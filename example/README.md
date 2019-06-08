
This example project implements a controller for ConfigMapCount CRD, which looks like this:

```yaml
apiVersion: silly.example.org/v1alpha1
kind: ConfigMapCount
metadata:
  name: configmapcount-sample
  namespace: default
status:
  configMaps: 3
```

This ((not) very useful) API object counts ConfigMaps in the same namespace and puts the count into its `.status.configMaps`
field. 

### Pre-requisites

Have your Kubernetes cluster ready. One great option is [kind](https://kind.sigs.k8s.io/):

### Build and Install

If you made any changes to API types, you need to generate Go and CRD YAML code: 

```bash
make generate
```

To install the CRDs into your Kubernetes cluster:

```bash
make install
```

### Run the controller

Run against the configured Kubernetes cluster in ~/.kube/config

```bash
make run
```
