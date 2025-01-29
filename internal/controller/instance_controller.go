/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fizzbuzzv1alpha1 "github.com/hegerdes/fizz-buzz-operator/api/v1alpha1"
)

const (
	// InstanceFinalizer is the finalizer used to delete the resources
	InstanceFinalizer = "cache.example.com/finalizer"

	// Definitions to manage status conditions
	// typeAvailableInstance represents the status of the Deployment reconciliation
	typeAvailableInstance = "Available"
	// typeDegradedInstance represents the status used when the custom resource is deleted and the finalizer operations are yet to occur.
	typeDegradedInstance = "Degraded"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=fizz-buzz.hegerdes.com,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fizz-buzz.hegerdes.com,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fizz-buzz.hegerdes.com,resources=instances/finalizers,verbs=update

// Extra rbac permissions
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Instance instance
	// The purpose is check if the Custom Resource for the Kind Instance
	// is applied on the cluster if not we return nil to stop the reconciliation
	Instance := &fizzbuzzv1alpha1.Instance{}
	err := r.Get(ctx, req.NamespacedName, Instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("Instance resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Instance")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status is available
	if len(Instance.Status.Conditions) == 0 {
		meta.SetStatusCondition(&Instance.Status.Conditions, metav1.Condition{Type: typeAvailableInstance, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, Instance); err != nil {
			log.Error(err, "Failed to update Instance status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the Instance Custom Resource after updating the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raising the error "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, Instance); err != nil {
			log.Error(err, "Failed to re-fetch Instance")
			return ctrl.Result{}, err
		}
	}

	// Let's add a finalizer. Then, we can define some operations which should
	// occur before the custom resource is deleted.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(Instance, InstanceFinalizer) {
		log.Info("Adding Finalizer for Instance")
		if ok := controllerutil.AddFinalizer(Instance, InstanceFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, Instance); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the Instance instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isInstanceMarkedToBeDeleted := Instance.GetDeletionTimestamp() != nil
	if isInstanceMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(Instance, InstanceFinalizer) {
			log.Info("Performing Finalizer Operations for Instance before delete CR")

			// Let's add here a status "Downgrade" to reflect that this resource began its process to be terminated.
			meta.SetStatusCondition(&Instance.Status.Conditions, metav1.Condition{Type: typeDegradedInstance,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", Instance.Name)})

			if err := r.Status().Update(ctx, Instance); err != nil {
				log.Error(err, "Failed to update Instance status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before removing the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			r.doFinalizerOperationsForInstance(Instance)

			// TODO(user): If you add operations to the doFinalizerOperationsForInstance method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.

			// Re-fetch the Instance Custom Resource before updating the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raising the error "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, Instance); err != nil {
				log.Error(err, "Failed to re-fetch Instance")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&Instance.Status.Conditions, metav1.Condition{Type: typeDegradedInstance,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", Instance.Name)})

			if err := r.Status().Update(ctx, Instance); err != nil {
				log.Error(err, "Failed to update Instance status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for Instance after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(Instance, InstanceFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for Instance")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, Instance); err != nil {
				log.Error(err, "Failed to remove finalizer for Instance")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	fizzBuzzLimit := Instance.Spec.Limit
	// loop over all fizz-buzz numbers using range
	for current := 1; current <= fizzBuzzLimit; current++ {
		podName := strings.ToLower(fmt.Sprintf("%s-seq%d-%s", Instance.Spec.Prefix, current, fizzbuzz(current)))
		// Check if the pods already exists, if not create a new one
		found := &corev1.Pod{}
		err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: Instance.Namespace}, found)
		if err != nil && apierrors.IsNotFound(err) {
			// Define a new pods
			pod, err := r.podForInstance(Instance, current)
			if err != nil {
				log.Error(err, "Failed to define new Pod resource for Instance")

				// The following implementation will update the status
				meta.SetStatusCondition(&Instance.Status.Conditions, metav1.Condition{Type: typeAvailableInstance,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create Pod for the custom resource (%s): (%s)", Instance.Name, err)})

				if err := r.Status().Update(ctx, Instance); err != nil {
					log.Error(err, "Failed to update Instance status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			log.Info("Creating a new Pod",
				"Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err = r.Create(ctx, pod); err != nil {
				log.Error(err, "Failed to create new Pod",
					"Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
				return ctrl.Result{}, err
			}

			// Pod created successfully
			// We will requeue the reconciliation so that we can ensure the state
			// and move forward for the next operations
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Pod")
			// Let's return the error for the reconciliation be re-trigged again
			return ctrl.Result{}, err
		}
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&Instance.Status.Conditions, metav1.Condition{Type: typeAvailableInstance,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Deployment for custom resource (%s) with %d replicas created successfully", Instance.Name, Instance.Spec.Limit)})

	if err := r.Status().Update(ctx, Instance); err != nil {
		log.Error(err, "Failed to update Instance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// podForInstance returns a Instance Deployment object
func (r *InstanceReconciler) podForInstance(
	Instance *fizzbuzzv1alpha1.Instance, current int) (*corev1.Pod, error) {
	ls := labelsForInstance(Instance.Name)
	containerCommand := fmt.Sprintf("while true; do cowsay \"%s: %s\"; sleep 60; done", Instance.Spec.Prefix, fizzbuzz(current))
	podName := strings.ToLower(fmt.Sprintf("%s-seq%d-%s", Instance.Spec.Prefix, current, fizzbuzz(current)))

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: Instance.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &[]bool{true}[0],
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
			},
			Containers: []corev1.Container{{
				Image:           Instance.Spec.Image,
				Name:            "cowsay",
				ImagePullPolicy: corev1.PullIfNotPresent,
				// Ensure restrictive context for the container
				// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
				SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             &[]bool{true}[0],
					RunAsUser:                &[]int64{1001}[0],
					ReadOnlyRootFilesystem:   &[]bool{true}[0],
					AllowPrivilegeEscalation: &[]bool{false}[0],
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{
							"ALL",
						},
					},
				},
				Command: []string{"sh", "-c", containerCommand},
			}},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(Instance, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}

// finalizeInstance will perform the required operations before delete the CR.
func (r *InstanceReconciler) doFinalizerOperationsForInstance(cr *fizzbuzzv1alpha1.Instance) {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	// Note: It is not recommended to use finalizers with the purpose of deleting resources which are
	// created and managed in the reconciliation. These ones, such as the Deployment created on this reconcile,
	// are defined as dependent of the custom resource. See that we use the method ctrl.SetControllerReference.
	// to set the ownerRef which means that the Deployment will be deleted by the Kubernetes API.
	// More info: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/

	// The following implementation will raise an event
	r.Recorder.Event(cr, "Warning", "Deleting",
		fmt.Sprintf("Custom Resource %s is being deleted from the namespace %s",
			cr.Name,
			cr.Namespace))
}

// returns the name of an instance based on fizzbuzz rules
func fizzbuzz(number int) string {
	if number%15 == 0 {
		return "FizzBuzz"
	} else if number%3 == 0 {
		return "Fizz"
	} else if number%5 == 0 {
		return "Buzz"
	}
	return fmt.Sprintf("%d", number)
}

// labelsForInstance returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func labelsForInstance(currentInstance string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "Instance-operator",
		"app.kubernetes.io/instance":   currentInstance,
		"app.kubernetes.io/managed-by": "InstanceController",
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fizzbuzzv1alpha1.Instance{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
