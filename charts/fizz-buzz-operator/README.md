# fizz-buzz-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

A Helm chart for Kubernetes

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| hegerdes | <hegerdes@outlook.de> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for pod. |
| commonLabels | object | `{}` | Labels applied to all manifests. |
| defaultEnvs | list | `[]` | List of default ENVs. No need to change |
| extraContainers | list | `[]` | Any additional containers. |
| extraDeploy | list | `[]` | Extra manifests |
| fullnameOverride | string | `""` | Override full release name. |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy of that image. |
| image.repository | string | `"hegerdes/fizz-buzz-operator"` | The container registry and image to use. |
| image.tag | string | `"latest"` | The image tag and/or sha. |
| imagePullSecrets | list | `[]` | Any repository secrets needed to pull the image. |
| initContainers | list | `[]` | Any additional init containers. |
| nameOverride | string | `""` | Override the application name. |
| nodeSelector | object | `{}` | Node selector for pod. |
| podAnnotations | object | `{}` | Extra annotations for the pod. |
| podContainerPort | int | `8080` | App and Container note. Change also in ENVs. |
| podEnvs | list | `[]` | List of ENVs to configure the app. |
| podSecurityContext | object | `{"fsGroup":2000}` | PodSecurity settings that will be applied to all containers. |
| rbac.enabled | bool | `true` | Enable (Cluster)Role and (Cluster)RoleBinding creation. |
| replicaCount | int | `1` |  |
| resources | object | `{}` | Resources for the container. |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}` | Security settings for the container. |
| service | object | `{"annotations":{},"internalTrafficPolicy":"Cluster","ipFamilyPolicy":"SingleStack","port":8080,"type":"ClusterIP"}` | How the service is exposed. |
| service.annotations | object | `{}` | Annotations for the service. |
| service.internalTrafficPolicy | string | `"Cluster"` | Service traffic policy. |
| service.ipFamilyPolicy | string | `"SingleStack"` | Service IP family. |
| service.port | int | `8080` | Service and container port. |
| service.type | string | `"ClusterIP"` | Service type. |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created. |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. |
| tolerations | list | `[]` | Tolerations for pod. |
| useUserNamespaces | bool | `false` | If pod should use user namespaces. Must be supported by CRI. See https://kubernetes.io/docs/tasks/configure-pod-container/user-namespaces/ |
| volumeMounts | list | `[]` | Volume mount's for container. |
| volumes | list | `[]` | Volumes where data should be persisted. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
