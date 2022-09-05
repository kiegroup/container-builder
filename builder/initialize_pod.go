package builder

import (
	"context"

	"github.com/pkg/errors"
	"github.com/ricardozanini/kogito-builder/api"
)

func newInitializePodAction() Action {
	return &initializePodAction{}
}

type initializePodAction struct {
	baseAction
}

// Name returns a common name of the action.
func (action *initializePodAction) Name() string {
	return "initialize-pod"
}

// CanHandle tells whether this action can handle the build.
func (action *initializePodAction) CanHandle(build *api.Build) bool {
	return build.Status.Phase == "" || build.Status.Phase == api.BuildPhaseInitialization
}

// Handle handles the builds.
func (action *initializePodAction) Handle(ctx context.Context, build *api.Build) (*api.Build, error) {
	if err := deleteBuilderPod(ctx, action.client, build); err != nil {
		return nil, errors.Wrap(err, "cannot delete build pod")
	}

	pod, err := getBuilderPod(ctx, action.client, build)
	if err != nil || pod != nil {
		// We return and wait for the pod to be deleted before de-queue the build pod.
		return nil, err
	}

	build.Status.Phase = api.BuildPhaseScheduling

	return build, nil
}
