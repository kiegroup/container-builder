package registry

import (
	"context"

	"github.com/kiegroup/container-builder/api"
	"github.com/kiegroup/container-builder/client"
	"github.com/kiegroup/container-builder/util/minikube"
)

// RetrieveAddress will retrieve a docker registry address from a RegistrySpec struct if present, otherwhise will try tu use the minikube one if available
// if no RegistrySPec is set and we are not on minikube we will return the minikube.FindRegistry error
func RetrieveAddress(registryTask api.RegistrySpec, ctx context.Context, c client.Client) (string, error) {
	if registryTask.Address != "" {
		return registryTask.Address, nil
	} else {
		reg, err := minikube.FindRegistry(ctx, c)
		return *reg, err
	}
}
