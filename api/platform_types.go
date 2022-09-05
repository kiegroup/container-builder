package api

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type PlatformBuild struct {
	ObjectReference
	Spec PlatformBuildSpec
}

type PlatformBuildSpec struct {
	// the strategy to adopt for building an Integration base image
	BuildStrategy BuildStrategy `json:"buildStrategy,omitempty"`
	// the strategy to adopt for publishing an Integration base image
	PublishStrategy PlatformBuildPublishStrategy `json:"publishStrategy,omitempty"`
	// a base image that can be used as base layer for all images.
	// It can be useful if you want to provide some custom base image with further utility software
	BaseImage string `json:"baseImage,omitempty"`
	// the image registry used to push/pull built images
	Registry RegistrySpec `json:"registry,omitempty"`
	// how much time to wait before time out the build process
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	//
	PublishStrategyOptions map[string]string `json:"PublishStrategyOptions,omitempty"`
}

// PlatformBuildPublishStrategy defines the strategy used to package and publish an Integration base image
type PlatformBuildPublishStrategy string

const (
	// PlatformBuildPublishStrategyKaniko uses Kaniko project (https://github.com/GoogleContainerTools/kaniko)
	// in order to push the incremental images to the image repository. It can be used with `pod` BuildStrategy.
	PlatformBuildPublishStrategyKaniko PlatformBuildPublishStrategy = "Kaniko"
)
