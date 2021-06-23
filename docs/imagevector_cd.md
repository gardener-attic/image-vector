# Image Vectors and Component Descriptors

Image vectors describe images that are deployed by a specific service.
These images are consequently part of the productive deployment of a service.
Therefore, they *MUST* be part of the component's ComponentDescriptor and can be documented, scanned, signed and transported. <br>
See [here](./imagevector_cd_parse.md) how images describes in image vector are added to a component descriptor. 

During the deployment the ComponentDescriptor should be the source of truth for all deployed artifacts.
This means that the images defined in a ComponentDescriptor have to be used by the service. <br>
See [here](./imagevector_cd_overwrite.md) how the image vector overwrite for a service is calculated using it's ComponentDescriptor.


#### Image Vector Generic Images - Component Descriptor flow

Gardener is a Kubernetes-as-a-Service offering that is able to manage Kubernetes clusters across cloudproviders.

These managed Kubernetes clusters can be provisioned with different versions which consequently results in different oci images per version.
The available versions for clusters that can be provisioned with Gardener depend on the actual Landscape configuration.
This means that the available cluster versions can differ from landscape to landscape, so Gardener must be able to describe these images in a generic way. 

This is done by using the image vector's [generic image entries](./imagevector.md#generic-image-entries).

The consequence of these generic image entries is that the actual used images can only determined if the available versions are defined.

See the diagram below that describes the old high level flow how the image vector of Gardener (`images.yaml`) is used within a landscape.
That landscape defines the Kubernetes Version in a separate component called `LSS`.

Using the generic image entries and the Kubernetes Versions result in a ComponentDescriptor that contains the actual images for the K8s versions.
This ComponentDescriptor can now be used to scan, sign, transport, etc.. the images (see `LSSD`).

During the deployment Gardener's ComponentDescriptor, image vector and the LSS Component Descriptor is used to generate the overwrite.

![old](img/old_image_vector_generation.png)
