# fizz-buzz-operator
The most famous coding challenge for interviews that proves that you build ... an advanced hello-world?  

Now show off your skills by adding Kubernetes to your fizz-buzz tool-belt.

## Description
A simple operator that creates the desired number of pods while naming them after the fizz-buzz rules.
## Getting Started
Just deploy the oerator via helm or kubectl:
```bash
# Via helm
helm install hackerman
# Or raw manifests
kubectl apply -f https://raw.githubusercontent.com/hegerdes/fizz-buzz-operator/main/charts/install.yaml
```

Then you can deploy the ``fizz-buzz Instance`` cr:
```yaml
apiVersion: fizz-buzz.hegerdes.com/v1alpha1
kind: Instance
metadata:
  name: hire-me
spec:
  prefix: hire-me
  limit: 15
  # Optional
  image: docker.io/rancher/cowsay:latest
```

### Prerequisites
- kubectl version v1.25+.
- Access to a Kubernetes v1.25+ cluster.

### To Deploy on the cluster


## Development
**Prerequisites:**  
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.25+.
- Access to a Kubernetes v1.25+ cluster.

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/fizz-buzz-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/fizz-buzz-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

**Delete the APIs(CRDs) and controller from the cluster:**

```sh
make uninstall && make undeploy
```

## Project Distribution
Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/fizz-buzz-operator:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/hegerdes/fizz-buzz-operator/main/charts/install.yaml
```

## Contributing
This a a fun project. If you have ideas on how the simplest problems can be solved in a even more complicated way, feel free to open a PR or Issue. But since this was just a finger-tipps excises please do not expect long term commitment for this project.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
