/*
Copyright Copyright 2022 Red Hat, Inc. and/or its affiliates.
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
package docker

import (
	"github.com/kiegroup/container-builder/cleaner"
	test "github.com/kiegroup/container-builder/test/registry"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// --------------------------- TEST SUITE -----------------

type DockerTestSuite struct {
	suite.Suite
	LocalRegistry cleaner.DockerLocalRegistry
	RegistryID    string
	Docker        cleaner.Docker
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (suite *DockerTestSuite) SetupSuite() {
	dockerRegistryContainer, registryID, docker := cleaner.SetupDockerSocket()
	if len(registryID) > 0 {
		suite.LocalRegistry = dockerRegistryContainer
		suite.RegistryID = registryID
		suite.Docker = docker
	} else {
		assert.FailNow(suite.T(), "Initialization failed %s", registryID)
	}
}

func (suite *DockerTestSuite) TearDownSuite() {
	registryID := suite.LocalRegistry.GetRegistryRunningID()
	if len(registryID) > 0 {
		cleaner.DockerTearDown(suite.LocalRegistry)
	} else {
		suite.LocalRegistry.StopRegistry()
	}
	purged, _ := suite.Docker.PurgeContainer("", cleaner.REGISTRY)
	logrus.Infof("Purged containers %t", purged)
}

// -------------------------------------- TESTS -----------------------------

func (suite *DockerTestSuite) TestDockerRegistry() {
	assert.Truef(suite.T(), suite.RegistryID != "", "Registry not started")
	assert.Truef(suite.T(), suite.LocalRegistry.IsRegistryImagePresent(), "Registry image not present")
	assert.Truef(suite.T(), suite.LocalRegistry.GetRegistryRunningID() == suite.RegistryID, "Registry container not running")
	assert.True(suite.T(), suite.LocalRegistry.Connection.DaemonHost() == "unix:///var/run/docker.sock")
}

func (suite *DockerTestSuite) TestPullTagPush() {
	assert.Truef(suite.T(), suite.RegistryID != "", "Registry not started")
	registryContainer, err := cleaner.GetRegistryContainer()
	assert.Nil(suite.T(), err)
	repos := test.CheckRepositoriesSize(suite.T(), 0, registryContainer)
	logrus.Info("Empty Repo Size = ", len(repos))

	result := dockerPullTagPushOnRegistryContainer(suite)
	logrus.Info("Pull Tag and Push Image on Registry Container successful = ", result)

	// Give some time to the registry to refresh status
	time.Sleep(2 * time.Second)
	repos = test.CheckRepositoriesSize(suite.T(), 1, registryContainer)
	logrus.Info("Repo Size after pull image = ", len(repos))
}

func dockerPullTagPushOnRegistryContainer(suite *DockerTestSuite) bool {
	dockerSocketConn, errSock := cleaner.GetDockerConnection()
	if errSock != nil {
		assert.FailNow(suite.T(), "Cant get docker socket")
	}
	d := cleaner.Docker{Connection: dockerSocketConn}
	if !suite.LocalRegistry.IsRegistryImagePresent() {
		err := d.PullImage(cleaner.TEST_IMAGE_TAG)
		if err != nil {
			assert.Fail(suite.T(), "Pull Image Failed", err)
			return false
		}
		logrus.Info("Pull image")
	}
	err := d.TagImage(cleaner.TEST_IMAGE_TAG, cleaner.TEST_IMAGE_LOCAL_TAG)
	if err != nil {
		assert.Fail(suite.T(), "Tag Image Failed", err)
		return false
	}
	logrus.Info("Tag image")
	err = d.PushImage(cleaner.TEST_IMAGE_LOCAL_TAG, cleaner.REGISTRY_CONTAINER_URL_FROM_DOCKER_SOCKET, "", "")
	if err != nil {
		assert.Fail(suite.T(), "Push Image Failed", err)
		return false
	}
	logrus.Info("Push image")
	return true
}
