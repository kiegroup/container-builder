package api

import (
	"github.com/kiegroup/container-builder/util/defaults"
)

func NewBuildahBuild(platformBuild PlatformBuild, publishImage string, buildName string) *Build {
	result := &Build{
		Spec: BuildSpec{
			Tasks: []Task{
				{Buildah: &BuildahTask{
					BaseTask: BaseTask{Name: "buildah-task"},
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
