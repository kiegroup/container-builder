package builder

import (
	"github.com/kiegroup/container-builder/api"
	"github.com/kiegroup/container-builder/util"
	"github.com/kiegroup/container-builder/util/defaults"
	"github.com/kiegroup/container-builder/util/log"
)

type buildahScheduler struct {
	*scheduler
	BuildahTask *api.BuildahTask
}

type buildahSchedulerHandler struct {
}

var _ schedulerHandler = &buildahSchedulerHandler{}

func (k buildahSchedulerHandler) CreateScheduler(info BuilderInfo, buildCtx buildContext) Scheduler {
	buildahTask := api.BuildahTask{
		BaseTask: api.BaseTask{Name: "BuildahTask"},
		PublishTask: api.PublishTask{
			BaseImage: info.Platform.Spec.BaseImage,
			Image:     info.FinalImageName,
			Registry:  info.Platform.Spec.Registry,
		},
		ExecutorImage: defaults.BuildahDefaultImageName,
	}

	buildCtx.Build = &api.Build{
		Spec: api.BuildSpec{
			Tasks:    []api.Task{{Buildah: &buildahTask}},
			Strategy: api.BuildStrategyPod,
			Timeout:  *info.Platform.Spec.Timeout,
		},
	}
	buildCtx.Build.Name = info.BuildUniqueName
	buildCtx.Build.Namespace = info.Platform.Namespace

	sched := &buildahScheduler{
		&scheduler{
			builder: builder{
				L:       log.WithName(util.ComponentName),
				Context: buildCtx,
			},
			Resources: make([]resource, 0),
		},
		&buildahTask,
	}
	// we hold our own reference for the default methods to return the right object
	sched.Scheduler = sched
	return sched
}

func (k buildahSchedulerHandler) CanHandle(info BuilderInfo) bool {
	return info.Platform.Spec.BuildStrategy == api.BuildStrategyPod && info.Platform.Spec.PublishStrategy == api.PlatformBuildPublishStrategyKaniko
}

func (sk *buildahScheduler) Schedule() (*api.Build, error) {
	// verify if we really need this
	for _, task := range sk.builder.Context.Build.Spec.Tasks {
		if task.Buildah != nil {
			task.Buildah = sk.BuildahTask
			break
		}
	}
	return sk.scheduler.Schedule()
}
