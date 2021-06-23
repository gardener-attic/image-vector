// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package pkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
)

// GenerateImageOverwrite parses a component descriptor and returns the defined image vector.
//
// Images can be defined in a component descriptor in 3 different ways:
// 1. as 'ociImage' resource: The image is defined a default resource of type 'ociImage' with a access of type 'ociRegistry'.
//    It is expected that the resource contains the following labels to be identified as image vector image.
//    The resulting image overwrite will contain the repository and the tag/digest from the access method.
// <pre>
// resources:
// - name: pause-container
//   version: "3.1"
//   type: ociImage
//   relation: external
//   extraIdentity:
//     "imagevector-gardener-cloud+tag": "3.1"
//   labels:
//   - name: imagevector.gardener.cloud/name
//     value: pause-container
//   - name: imagevector.gardener.cloud/repository
//     value: gcr.io/google_containers/pause-amd64
//   - name: imagevector.gardener.cloud/source-repository
//     value: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
//   - name: imagevector.gardener.cloud/target-version
//     value: "< 1.16"
//   access:
//     type: ociRegistry
//     imageReference: gcr.io/google_containers/pause-amd64:3.1
// </pre>
//
// 2. as component reference: The images are defined in a label "imagevector.gardener.cloud/images".
//    The resulting image overwrite will contain all images defined in the images label.
//    Their repository and tag/digest will be matched from the resources defined in the actual component's resources.
//
//   Note: The images from the label are matched to the resources using their name and version. The original image reference do not exit anymore.
//
// <pre>
// componentReferences:
// - name: cluster-autoscaler-abc
//   componentName: github.com/gardener/autoscaler
//   version: v0.10.1
//   labels:
//   - name: imagevector.gardener.cloud/images
//     value:
//       images:
//       - name: cluster-autoscaler
//         repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
//         tag: "v0.10.1"
// </pre>
//
// 3. as generic images from the component descriptor labels.
//   Generic images are images that do not directly result in a resource.
//   They will be matched with another component descriptor that actually defines the images.
//   The other component descriptor MUST have the "imagevector.gardener.cloud/name" label in order to be matched.
//
// <pre>
// meta:
//   schemaVersion: 'v2'
// component:
//   labels:
//   - name: imagevector.gardener.cloud/images
//     value:
//       images:
//       - name: hyperkube
//         repository: k8s.gcr.io/hyperkube
//         targetVersion: "< 1.19"
// </pre>
//
// <pre>
// meta:
//   schemaVersion: 'v2'
// component:
//   resources:
//   - name: hyperkube
//     version: "v1.19.4"
//     type: ociImage
//     extraIdentity:
//       "imagevector-gardener-cloud+tag": "v1.19.4"
//     labels:
//     - name: imagevector.gardener.cloud/name
//       value: hyperkube
//     - name: imagevector.gardener.cloud/repository
//       value: k8s.gcr.io/hyperkube
//     access:
//	   type: ociRegistry
//	   imageReference: my-registry/hyperkube:v1.19.4
// </pre>
func GenerateImageOverwrite(ctx context.Context,
	compResolver ctf.ComponentResolver,
	cd *cdv2.ComponentDescriptor,
	list *cdv2.ComponentDescriptorList) (*ImageVector, error) {
	imageVector := &ImageVector{}

	// parse all images from the component descriptors resources
	images, err := parseImagesFromResources(cd.Resources)
	if err != nil {
		return nil, err
	}
	imageVector.Images = append(imageVector.Images, images...)

	images, err = parseImagesFromComponentReferences(ctx, compResolver, cd)
	if err != nil {
		return nil, err
	}
	imageVector.Images = append(imageVector.Images, images...)

	images, err = parseGenericImages(cd, list)
	if err != nil {
		return nil, err
	}
	imageVector.Images = append(imageVector.Images, images...)

	return imageVector, nil
}

// parseImagesFromResources parse all images from the component descriptors resources
func parseImagesFromResources(resources []cdv2.Resource) ([]ImageEntry, error) {
	images := make([]ImageEntry, 0)
	for _, res := range resources {
		if res.GetType() != cdv2.OCIImageType || res.Access.GetType() != cdv2.OCIRegistryType {
			continue
		}
		var name string
		if ok, err := getLabel(res.GetLabels(), NameLabel, &name); !ok || err != nil {
			if err != nil {
				return nil, fmt.Errorf("unable to get name for %q: %w", res.GetName(), err)
			}
			continue
		}

		entry := ImageEntry{
			Name: name,
		}

		if err := parseResourceAccess(&entry, res); err != nil {
			return nil, err
		}

		// set additional information
		var targetVersion string
		if ok, err := getLabel(res.GetLabels(), TargetVersionLabel, &targetVersion); ok || err != nil {
			if err != nil {
				return nil, fmt.Errorf("unable to get target version for %q: %w", res.GetName(), err)
			}
			entry.TargetVersion = &targetVersion
		}
		var runtimeVersion string
		if ok, err := getLabel(res.GetLabels(), RuntimeVersionLabel, &runtimeVersion); ok || err != nil {
			if err != nil {
				return nil, fmt.Errorf("unable to get target version for %q: %w", res.GetName(), err)
			}
			entry.RuntimeVersion = &runtimeVersion
		}

		images = append(images, entry)
	}
	return images, nil
}

// parseImagesFromComponentReferences parse all images from the component descriptors references
func parseImagesFromComponentReferences(ctx context.Context, compResolver ctf.ComponentResolver, ca *cdv2.ComponentDescriptor) ([]ImageEntry, error) {
	images := make([]ImageEntry, 0)

	for _, ref := range ca.ComponentReferences {

		// only resolve the component reference if a images.yaml is defined
		imageVector := &ImageVector{}
		if ok, err := getLabel(ref.GetLabels(), ImagesLabel, imageVector); !ok || err != nil {
			if err != nil {
				return nil, fmt.Errorf("unable to parse images label from component reference %q: %w", ref.GetName(), err)
			}
			continue
		}

		refCD, err := compResolver.Resolve(ctx, ca.GetEffectiveRepositoryContext(), ref.ComponentName, ref.Version)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve component descriptor %q: %w", ref.GetName(), err)
		}

		// find the matching resource by name and version
		for _, image := range imageVector.Images {
			foundResources, err := refCD.GetResourcesByName(image.Name)
			if err != nil {
				return nil, fmt.Errorf("unable to find images for %q in component refernce %q: %w", image.Name, ref.GetName(), err)
			}
			for _, res := range foundResources {
				if res.GetVersion() != *image.Tag {
					continue
				}
				if err := parseResourceAccess(&image, res); err != nil {
					return nil, fmt.Errorf("unable to find images for %q in component refernce %q: %w", image.Name, ref.GetName(), err)
				}
				images = append(images, image)
			}
		}

	}

	return images, nil
}

// parseGenericImages parses the generic images of the component descriptor and matches all oci resources of the other component descriptors
func parseGenericImages(ca *cdv2.ComponentDescriptor, list *cdv2.ComponentDescriptorList) ([]ImageEntry, error) {
	images := make([]ImageEntry, 0)
	imageVector := &ImageVector{}
	if ok, err := getLabel(ca.GetLabels(), ImagesLabel, imageVector); !ok || err != nil {
		if err != nil {
			return nil, fmt.Errorf("unable to parse images label from component reference %q: %w", ca.GetName(), err)
		}
		return images, nil
	}

	for _, image := range imageVector.Images {
		if image.TargetVersion == nil {
			continue
		}
		constr, err := semver.NewConstraint(*image.TargetVersion)
		if err != nil {
			return nil, fmt.Errorf("unable to parse target version for %q: %w", image.Name, err)
		}

		for _, comp := range list.Components {
			resources, err := comp.GetResourcesByType(cdv2.OCIImageType)
			if err != nil {
				if errors.Is(err, cdv2.NotFound) {
					continue
				}
				return nil, fmt.Errorf("unable to get oci resources from %q: %w", comp.GetName(), err)
			}
			for _, res := range resources {
				var imageName string
				ok, err := getLabel(res.GetLabels(), NameLabel, &imageName)
				if err != nil {
					return nil, fmt.Errorf("unable to parse image name label from resource %q of component %q: %w", res.GetName(), ca.GetName(), err)
				}
				if !ok || imageName != image.Name {
					continue
				}
				semverVersion, err := semver.NewVersion(res.GetVersion())
				if err != nil {
					return nil, fmt.Errorf("unable to parse resource version from resource %q of component %q: %w", res.GetName(), ca.GetName(), err)
				}
				if !constr.Check(semverVersion) {
					continue
				}

				entry := ImageEntry{
					Name: image.Name,
				}
				if err := parseResourceAccess(&entry, res); err != nil {
					return nil, fmt.Errorf("unable to parse oci access from resource %q of component %q: %w", res.GetName(), ca.GetName(), err)
				}
				targetVersion := fmt.Sprintf("= %s", res.GetVersion())
				entry.TargetVersion = &targetVersion
				images = append(images, entry)
			}
		}

	}

	return images, nil
}
