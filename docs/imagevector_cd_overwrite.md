# Parse a component descriptor and generate an Image Vectors Overwrite

This document describes the process how a component descriptor (and optional dependent component descriptors) can be parsed and an image vector overwrite is generated.<br>
A dependent component descriptor is only needed if generic images are defined (see [way #3](#3-as-generic-images-from-the-component-descriptors-labels) for details.).

See the [go docs](https://github.com/gardener/image-vector/blob/main/pkg/overwrite.go#L100) for a description how the go function can be used.

Images can be defined in a component descriptor in 3 different ways:

#### 1. as 'ociImage' resource
   
The image is defined a default resource of type 'ociImage' with a access of type 'ociRegistry'.
   
It is expected that the resource contains the following labels to be identified as image vector image.
The resulting image overwrite will contain the repository and the tag/digest from the access method.
<pre>
resources:
- name: pause-container
  version: "3.1"
  type: ociImage
  relation: external
  extraIdentity:
    "imagevector-gardener-cloud+tag": "3.1"
  labels:
  - name: imagevector.gardener.cloud/name
    value: pause-container
  - name: imagevector.gardener.cloud/repository
    value: gcr.io/google_containers/pause-amd64
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  - name: imagevector.gardener.cloud/target-version
    value: "< 1.16"
  access:
    type: ociRegistry
    imageReference: gcr.io/google_containers/pause-amd64:3.1
</pre>

#### 2. as component reference: 

The images are defined in a label "imagevector.gardener.cloud/images".

The resulting image overwrite will contain all images defined in the images label.
Their repository and tag/digest will be matched from the resources defined in the actual component's resources.

By default the actual resource is matched using the `resourceId` where the referenced component descriptor's resources are serached for a resource matching the id.

If no resourceId is defined the resource is tried to be matched
1. by using the name of the image entry and the name of the resource.
2. by using the original image reference defined by the gardenerci label `cloud.gardener.cnudie/migration/original_ref`.
2. by using resource's imageReference.

<pre>
componentReferences:
- name: cluster-autoscaler-abc
  componentName: github.com/gardener/autoscaler
  version: v0.10.1
  labels:
  - name: imagevector.gardener.cloud/images
    value:
      images:
      - name: cluster-autoscaler
        resourceId:
          name: cluster-autoscaler
        repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
        tag: "v0.10.1"
</pre>

#### 3. as **generic images** from the component descriptor's labels

Generic images are images that do not directly result in a resource.
They will be matched with another component descriptor that actually defines the images.
The other component descriptor MUST have the extraIdentity "imagevector-gardener-cloud+repository" in order to be matched.

<pre>
meta:
  schemaVersion: 'v2'
component:
  labels:
  - name: imagevector.gardener.cloud/images
    value:
      images:
      - name: hyperkube
        repository: k8s.gcr.io/hyperkube
        targetVersion: "< 1.19"
</pre>

<pre>
meta:
  schemaVersion: 'v2'
component:
  resources:
  - name: hyperkube
    version: "v1.19.4"
    type: ociImage
    extraIdentity:
      "imagevector-gardener-cloud+tag": "v1.19.4"
      "imagevector-gardener-cloud+repository": "k8s.gcr.io/hyperkube"
    labels:
    - name: imagevector.gardener.cloud/repository
      value: k8s.gcr.io/hyperkube
    access:
	  type: ociRegistry
	  imageReference: my-registry/hyperkube:v1.19.4
</pre>
