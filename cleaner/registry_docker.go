/*
Copyright 2022 Red Hat, Inc. and/or its affiliates.
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
package cleaner

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"strings"
	"time"
)

func GetDockerConnection() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return cli, nil
}

func GetCustomDockerConnectionWithIP(ip string) (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithHost(ip))
}

func GetCustomRegistry() (DockerLocalRegistry, *client.Client, error) {
	connectionLocal, err := GetCustomDockerConnectionWithIP(REGISTRY_CONTAINER_URL_FROM_DOCKER_SOCKET)
	if err != nil {
		logrus.Error(err)
		return DockerLocalRegistry{}, nil, err
	}

	d := DockerLocalRegistry{Connection: connectionLocal}
	return d, connectionLocal, nil
}

func (d DockerLocalRegistry) getConnection() (*client.Client, error) {
	connectionLocal := d.Connection
	if connectionLocal == nil {
		return GetDockerConnection()
	}
	return connectionLocal, nil
}

func (d DockerLocalRegistry) StartRegistry() string {
	//wait until Podman registry shutdown in the podman tests
	for {
		time.Sleep(1 * time.Second)
		if IsPortAvailable("5000") {
			break
		}
	}

	ctx := context.Background()
	registryID := d.GetRegistryRunningID()
	logrus.Infof("registryID %s", registryID)
	if len(registryID) > 0 {
		logrus.Info("Registry already running")
		return registryID
	}

	if !d.IsRegistryImagePresent() {
		logrus.Info("Registry Image Pull")
		_, err := d.Connection.ImagePull(ctx, REGISTRY, types.ImagePullOptions{})
		if err != nil {
			fmt.Println(err)
			return ""
		}
	} else {
		logrus.Info("Registry Image ready")
	}

	logrus.Info("Registry Container Create")
	resp, err := d.Connection.ContainerCreate(ctx, &container.Config{
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
		logrus.Error(err)
	}

	logrus.Info("Starting Registry container")
	if err := d.Connection.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logrus.Info("error during Start registry")
		logrus.Error(err)
		return ""
	}

	// give some time to start
	logrus.Info("Waiting 4 sec")
	time.Sleep(4 * time.Second)
	return d.GetRegistryRunningID()
}

func (d DockerLocalRegistry) StopRegistry() bool {
	registryID := d.GetRegistryRunningID()
	if len(registryID) > 0 {
		logrus.Info("StopRegistry Kill registry container.ID " + registryID)
		ctx := context.Background()
		_ = d.Connection.ContainerKill(ctx, registryID, "SIGKILL")
		logrus.Info("StopRegistry Removing container ID " + registryID)
		err := d.Connection.ContainerRemove(ctx, registryID, types.ContainerRemoveOptions{})
		if err != nil {
			logrus.Info(err)
			return false
		}
	}
	return true
}

func (d DockerLocalRegistry) StopAndRemoveContainer(containerID string) bool {
	if len(containerID) > 0 {
		ctx := context.Background()
		logrus.Info("Docker StopAndRemoveContainer Kill registry container container.ID " + containerID)
		_ = d.Connection.ContainerKill(ctx, containerID, "SIGKILL")
		logrus.Info("Docker StopAndRemoveContainer Removing container ID " + containerID)
		err := d.Connection.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
		return err == nil
	}
	fmt.Println("Docker StopAndRemoveContainer Invalid ID " + containerID)
	return true
}

func (d DockerLocalRegistry) GetRegistryRunningID() string {
	containers, err := d.Connection.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, container := range containers {
		if container.Image == REGISTRY {
			return container.ID
		}
	}
	return ""
}

func (d DockerLocalRegistry) IsRegistryImagePresent() bool {

	imageList, err := d.Connection.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false
	}
	for _, imagex := range imageList {
		if imagex.RepoTags[0] == REGISTRY || (imagex.RepoDigests != nil && strings.HasPrefix(imagex.RepoDigests[0], REGISTRY)) {
			return true
		}
	}
	return false
}

func SetupDockerSocket() (DockerLocalRegistry, string, Docker) {
	dockerSocketConn, err := GetDockerConnection()
	if err != nil {
		logrus.Errorf("Can't get Docker socket")
		return DockerLocalRegistry{}, "", Docker{}
	}
	dockerSock := Docker{Connection: dockerSocketConn}
	dockerSock.PurgeContainer("", REGISTRY)

	d := DockerLocalRegistry{Connection: dockerSocketConn}
	if !d.IsRegistryImagePresent() {
		dockerSock.PullImage(TEST_IMAGE_TAG)
	}
	registryID := d.GetRegistryRunningID()
	if len(registryID) == 0 {
		registryID = d.StartRegistry()
	} else {
		logrus.Infof("Registry already up and running with ID %s", registryID)
	}
	return d, registryID, dockerSock

}

func DockerTearDown(dlr DockerLocalRegistry) {
	if len(dlr.GetRegistryRunningID()) > 0 {
		registryID := dlr.GetRegistryRunningID()
		dlr.StopAndRemoveContainer(registryID)
	} else {
		dlr.StopRegistry()
	}
}
