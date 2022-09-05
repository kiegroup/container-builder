package builder

import (
	"context"

	"github.com/ricardozanini/kogito-builder/api"
	"github.com/ricardozanini/kogito-builder/client"
	"github.com/ricardozanini/kogito-builder/util/log"
)

type Action interface {
	log.Injectable
	client.Injectable
	// Name returns user-friendly name for the action
	Name() string

	// CanHandle returns true if the action can handle the build
	CanHandle(build *api.Build) bool

	// Handle executes the handling function
	Handle(ctx context.Context, build *api.Build) (*api.Build, error)
}

type baseAction struct {
	client client.Client
	L      log.Logger
}

// TODO: implement our client wrapper

func (action *baseAction) InjectClient(client client.Client) {
	action.client = client
}

func (action *baseAction) InjectLogger(log log.Logger) {
	action.L = log
}
