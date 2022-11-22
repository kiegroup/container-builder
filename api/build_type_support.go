package api

import "path"

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
