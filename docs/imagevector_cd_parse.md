# Parse Image Vectors and add to component descriptors

This document describes the process how images, defined by an image vector, are added to a component descriptor.

See the [go docs](https://github.com/gardener/image-vector/blob/main/pkg/imagevector.go#L190) for a description how the go function can be used.

There are 4 different scenarios how images are added to the component descriptor.
#### 1. The image is defined with a tag

If an image entry contains a tag the default behavior is to add it directly as oci image resource.

<pre>
images:
- name: pause-container
  sourceRepository: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  repository: gcr.io/google_containers/pause-amd64
  tag: "3.1"
</pre>

<pre>
meta:
  schemaVersion: 'v2'
...
resources:
- name: pause-container
  version: "3.1"
  type: ociImage
  extraIdentity:
    "imagevector-gardener-cloud+tag": "3.1"
  labels:
  - name: imagevector.gardener.cloud/name
    value: pause-container
  - name: imagevector.gardener.cloud/repository
    value: gcr.io/google_containers/pause-amd64
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  access:
    type: ociRegistry
    imageReference: gcr.io/google_containers/pause-amd64:3.1
</pre>

#### 2. The image is defined by another component

If an image entry contains a tag but is defined by another component, the resource is not directly added as resource but a reference to the component.
That reference consists of a `componentReference` to that component and a label ("imagevector.gardener.cloud/images").
The referenced component name is defined by the `sourceRepository` and the version is defined by the images `tag`.

Images that are defined by other components can be specified
1. when the image's repository matches the given "--component-prefixes"
2. the image is labeled with "imagevector.gardener.cloud/component-reference"

If the component reference is not yet defined it will be automatically added.
If multiple images are defined for the same component reference they are added to the images list in the label.

<pre>
images:
- name: cluster-autoscaler
  sourceRepository: github.com/gardener/autoscaler
  repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
  targetVersion: "< 1.16"
  tag: "v0.10.0"
  labels: # recommended bbut only needed when "--component-prefixes" is not defined
  - name: imagevector.gardener.cloud/component-reference
    value:
      name: cla # defaults to image.name
      componentName: github.com/gardener/autoscaler # defaults to image.sourceRepository
      version: v0.10.0 # defaults to image.version
</pre>

<pre>
meta:
  schemaVersion: 'v2'
...
componentReferences:
- name: cla
  componentName: github.com/gardener/autoscaler
  version: v0.10.0
  extraIdentity:
    imagevector-gardener-cloud+tag: v0.10.0
  labels:
  - name: imagevector.gardener.cloud/images
    value:
	  images:
	  - name: cluster-autoscaler
	    repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
	    sourceRepository: github.com/gardener/autoscaler
	    tag: v0.10.0
	    targetVersion: '< 1.16'
</pre>

#### 3. The image is a generic dependency

A generic dependency image is not part of a component descriptor's resource but will be added as label ("imagevector.gardener.cloud/images") to the component descriptor.
Generic dependencies do not have a tag as their actual tag depends on other external factors.

Generic dependencies can be defined by
1. image entries without a tag
2. defined as "--generic-dependency=<image name>"
3. the label "imagevector.gardener.cloud/generic"

<pre>
images:
- name: hyperkube
  sourceRepository: github.com/kubernetes/kubernetes
  repository: k8s.gcr.io/hyperkube
  targetVersion: "< 1.19"
  labels: # only needed if "--generic-dependency" is not set
  - name: imagevector.gardener.cloud/generic
</pre>

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
	    sourceRepository: github.com/kubernetes/kubernetes
	    targetVersion: '< 1.19'
</pre>

#### 4. The image has no tag and it's repository matches an already defined resource in the component descriptor.

This usually means that the image is build as part of the build pipeline and the version depends on the current component.
In this case only labels are added to the existing resource

<pre>
images:
- name: gardenlet
  sourceRepository: github.com/gardener/gardener
  repository: eu.gcr.io/gardener-project/gardener/gardenlet
</pre>

<pre>
meta:
  schemaVersion: 'v2'
...
resources:
- name: gardenlet
  version: "v0.0.0"
  type: ociImage
  relation: local
  labels:
  - name: imagevector.gardener.cloud/name
    value: gardenlet
  - name: imagevector.gardener.cloud/repository
    value: eu.gcr.io/gardener-project/gardener/gardenlet
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/gardener/gardener
  access:
    type: ociRegistry
    imageReference: eu.gcr.io/gardener-project/gardener/gardenlet:v0.0.0
</pre>
