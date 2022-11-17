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
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func (d Docker) StartRegistry() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
	}

	ctx := context.Background()
	if d.IsRegistryRunning() {
		return true
	}
	if !d.IsRegistryImagePresent() {
		_, err := cli.ImagePull(context.Background(), REGISTRY, types.ImagePullOptions{})
		if err != nil {
			fmt.Println(err)
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        REGISTRY,
		ExposedPorts: nat.PortSet{"5000": struct{}{}},
	},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("5000"): {{HostIP: "127.0.0.1", HostPort: "5000"}}},
		},
		nil,
		nil,
		REGISTRY)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Starting Registry container...")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (d Docker) StopRegistry() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println(err)
	}

	for _, container := range containers {
		if container.Image == REGISTRY {
			sec, _ := time.ParseDuration("10sec")
			fmt.Println("Stop registry container")
			err = cli.ContainerStop(context.Background(), container.ID, &sec)
			if err != nil {
				fmt.Println(err)
			}
			_ = cli.ContainerKill(context.Background(), container.ID, "")

			err = cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{})
			return err == nil
		}
	}
	return false
}

func (d Docker) IsRegistryRunning() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
		return false
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, container := range containers {
		if container.Image == REGISTRY {
			fmt.Println("Registry container already running...")
			return true
		}
	}
	return false
}

func (d Docker) IsRegistryImagePresent() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
		return false
	}
	imageList, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false
	}
	for _, imagex := range imageList {
		if imagex.RepoTags[0] == REGISTRY || strings.HasPrefix(imagex.RepoDigests[0], REGISTRY) {
			return true
		}
	}
	return false
}

func (d Docker) RemoveRegistryContainerAndImage() {
	d.StopRegistry()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
	}
	containerList, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})

	for _, container := range containerList {
		if container.Image == REGISTRY {
			err = cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{})
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	imageList, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		fmt.Println(err)
	}
	for _, imagex := range imageList {
		if imagex.RepoTags[0] == REGISTRY {
			_, err := cli.ImageRemove(context.Background(), imagex.ID, types.ImageRemoveOptions{})
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}
