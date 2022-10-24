package builder

import (
	"context"

	"github.com/kiegroup/container-builder/api"
)

func newErrorAction() Action {
	return &errorAction{}
}

type errorAction struct {
	baseAction
}

// Name returns a common name of the action.
func (action *errorAction) Name() string {
	return "error"
}

// CanHandle tells whether this action can handle the build.
func (action *errorAction) CanHandle(build *api.Build) bool {
	return build.Status.Phase == api.BuildPhaseError
}

// Handle handles the builds.
func (action *errorAction) Handle(ctx context.Context, build *api.Build) (*api.Build, error) {
	return nil, nil
}
