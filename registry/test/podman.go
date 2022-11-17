/*
Copyright 2022.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/sirupsen/logrus"
)

func (p Podman) GetConnection() (context.Context, error) {
	// ROOTLESS access
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	socket := "unix:" + sockDir + "/podman/podman.sock"
	conn, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		logrus.Errorf("%s \n", err)
		return nil, err
	}
	return conn, err
}

func (p Podman) StartRegistry() bool {
	connection, _ := p.GetConnection()

	if p.IsRegistryRunning() {
		fmt.Println("Registry container already running...")
		return true
	}
	isImageRegistryPresent := p.IsRegistryImagePresent()

	if !isImageRegistryPresent {
		fmt.Println("Pulling Docker Registry...")
		_, err := images.Pull(connection, REGISTRY, nil)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	// Container create
	s := specgen.NewSpecGenerator(REGISTRY, false)
	s.Terminal = true
	r, err := containers.CreateWithSpec(connection, s, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("Starting Registry container...")
	err = containers.Start(connection, r.ID, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	wait := define.ContainerStateRunning
	_, err = containers.Wait(connection, r.ID, new(containers.WaitOptions).WithCondition([]define.ContainerStatus{wait}))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (p Podman) IsRegistryImagePresent() bool {
	connection, _ := p.GetConnection()
	imageList, err := images.List(connection, nil)
	if err != nil {
		return false
	}
	for _, imagex := range imageList {
		if strings.HasPrefix(imagex.Names[0], REGISTRY_FULL) {
			return true
		}
	}
	return false
}

func (p Podman) IsRegistryRunning() bool {
	connection, _ := p.GetConnection()
	containersList, err := containers.List(connection, nil)
	if err != nil {
		fmt.Println(err)
	}

	for _, container := range containersList {
		if strings.HasPrefix(container.Image, REGISTRY_FULL) {
			fmt.Println("Registry container already running...")
			return true
		}
	}
	return false
}

func (p Podman) StopRegistry() bool {
	connection, _ := p.GetConnection()
	containersList, err := containers.List(connection, nil)
	if err != nil {
		fmt.Println(err)
	}

	for _, container := range containersList {
		if container.Image == REGISTRY {
			fmt.Println("Registry container already running...")
			_ = containers.Stop(context.Background(), container.ID, nil)
			_ = containers.Kill(context.Background(), container.ID, nil)
			_, err = containers.Remove(context.Background(), container.ID, nil)
			if err != nil {
				fmt.Println(err)
				return false
			}
		}
	}
	return true
}

func (p Podman) RemoveRegistryContainerAndImage() {
	p.StopRegistry()
	connection, _ := p.GetConnection()
	containerList, _ := containers.List(connection, nil)
	for _, container := range containerList {
		if container.Image == REGISTRY {
			_ = containers.Stop(context.Background(), container.ID, nil)
			_ = containers.Kill(context.Background(), container.ID, nil)
			_, _ = containers.Remove(context.Background(), container.ID, nil)
		}
	}

	imageList, _ := images.List(connection, nil)

	for _, imagex := range imageList {
		if strings.HasPrefix(imagex.Names[0], REGISTRY_FULL) {
			imagesNames := []string{imagex.ID}
			images.Remove(context.Background(), imagesNames, nil)
		}
	}
}
