package api

import (
	"github.com/kiegroup/container-builder/util/defaults"
	"path"
)

func NewKanikoBuild(platformBuild PlatformBuild, publishImage string, buildName string) *Build {
	result := &Build{
		Spec: BuildSpec{
			Tasks: []Task{
				{Kaniko: &KanikoTask{
					BaseTask: BaseTask{Name: "KanikoTask"},
					PublishTask: PublishTask{
						ContextDir: path.Join("/builder", buildName, "context"),
						BaseImage:  platformBuild.Spec.BaseImage,
						Image:      publishImage,
						Registry:   platformBuild.Spec.Registry,
					},
					Cache: KanikoTaskCache{},
				}},
			},
			Strategy: BuildStrategyPod,
			Timeout:  *platformBuild.Spec.Timeout,
		},
	}
	result.Name = buildName
	result.Namespace = platformBuild.Namespace
	return result
}

func NewBuildahBuild(platformBuild PlatformBuild, publishImage string, buildName string) *Build {
	result := &Build{
		Spec: BuildSpec{
			Tasks: []Task{
				{Buildah: &BuildahTask{
					BaseTask: BaseTask{Name: "BuildahTask"},
					PublishTask: PublishTask{
						BaseImage: platformBuild.Spec.BaseImage,
						Image:     publishImage,
						Registry:  platformBuild.Spec.Registry,
					},
					ExecutorImage: defaults.BuildahDefaultImageName,
				}},
			},
			Strategy: BuildStrategyPod,
			Timeout:  *platformBuild.Spec.Timeout,
		},
	}
	result.Name = buildName
	result.Namespace = platformBuild.Namespace
	return result
}
