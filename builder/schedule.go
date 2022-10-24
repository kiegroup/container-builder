package builder

import (
	"context"

	"github.com/kiegroup/container-builder/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newScheduleAction() Action {
	return &scheduleAction{}
}

type scheduleAction struct {
	baseAction
}

// Name returns a common name of the action.
func (action *scheduleAction) Name() string {
	return "schedule"
}

// CanHandle tells whether this action can handle the build.
func (action *scheduleAction) CanHandle(build *api.Build) bool {
	return build.Status.Phase == api.BuildPhaseScheduling
}

// Handle handles the builds.
func (action *scheduleAction) Handle(ctx context.Context, build *api.Build) (*api.Build, error) {
	// TODO do any work required between initialization and scheduling, like enqueueing builds
	now := metav1.Now()
	build.Status.StartedAt = &now
	build.Status.Phase = api.BuildPhasePending

	return build, nil
}
