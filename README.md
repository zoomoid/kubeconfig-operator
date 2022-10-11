# kubeconfig-operator

The Kubeconfig operator is meant to ease generation of Kubeconfig files for users of the cluster. Cluster-admins can create a
custom resource `Kubeconfig`, and the operator will handle the steps required to create a fully-signed and bound kubeconfig
YAML to be delivered to the new cluster user (because some of the steps can get a bit tedious, like handling OpenSSL certificate signing
requests by hand).

## Description

The operator consists of a Custom Resource Definition for the Kubeconfig request, and two controllers.

The custom resource serves as a request for a new Kubeconfig, which on top of all includes a username.
You can also add further configuration, such as toggling auto-approval and selecting other parameters for the certificate
request. Have a look at `./config/samples` for some example Kubeconfigs.

The first controller reconciles all Kubeconfig custom resources, and acts as the manager of the workflow. It creates secrets, certificate signing
requests and cluster role bindings for the Kubeconfig object and is the owner of all created resources such that garbage collection works as expected. The second controller reconciles all certificate signing requests and auto-approves requests
that where annotated to be automatically approved.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/kubeconfig-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/kubeconfig-operator:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing

Contribution is welcome, and I'm up for any feature requests. Just fork the project and open a pull request with an accompanying issue
marked accordingly.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022 zoomoid.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
