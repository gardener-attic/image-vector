# Image Vector

Image vectors describe container images that are deployed into a cluster by a service.

For example [Gardener](https://github.com/gardener/gardener) and other extensions are using the image vector to describe images that are deployed into seed and shoot clusters.

Image vectors can also be overwritten which means that image locations can be replaced.
This is useful if images are copied to private or regional registries.

### Structure

An image vector is a list of image entries.<br>
An image entry consists of 
- **a semantic name** that is used by the application to identify the image.
- **a repository** that describes the oci image repository.
- **an optional tag** that describes the tag or digest of the image. <br>
  The actual image location is calculated using the repository and the tag: `<repository>:<tag>` or `<repository>@<tag>` for digests. <br>
  See [generic image entries](#generic-image-entries) for a description if the tag is not defined.
- **an optional target version** that describes the version of the Kubernetes cluster that image is targeted for. <br>
  The version is a semver constraint as defined by the [mastermind semver lib](https://github.com/Masterminds/semver#checking-version-constraints) so that multiple Kubernetes version can be matched.
- **an optional source repository** that describes the source github repository or docker file. This is only used for documentation purposes and has no effect for the evaluation.
  
##### Example

In this example the service will use the `pause-container` with tag `3.0` for all clusters with Kubernetes version `1.15.x`, and tag `3.1` for all clusters with Kubernetes >= 1.16.

```yaml
images:
- name: pause-container
  sourceRepository: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  repository: gcr.io/google_containers/pause-amd64
  tag: "3.0"
  version: 1.15.x
- name: pause-container
  sourceRepository: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  repository: gcr.io/google_containers/pause-amd64
  tag: "3.1"
  version: ">= 1.16"
...
```

### Generic image entries

The tag of an image entry is optional as described in the [image vector structure](#structure).

The default behavior for an undefined tag is that the target kubernetes version is used as tag.

*Example*: 

In this example the service would deploy the image `k8s.gcr.io/hyperkube:v1.15.5` if deployed to a Kubernetes cluster with version `v1.15.5` and the image `k8s.gcr.io/hyperkube:v1.21.0` if deployed to a Kubernetes cluster with version `v1.21.0`

```yaml
images:
- name: pause-container
  repository: k8s.gcr.io/hyperkube
...
```

> Note: This behavior can differ from service to service.
> Please consult the respective services documentation for their interpretation.
