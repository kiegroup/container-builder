//go:build integration
// +build integration

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
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// --------------------------- TEST SUITE -----------------
type DockerTestSuite struct {
	suite.Suite
	LocalRegistry DockerLocalRegistry
	RegistryID    string
	Docker        Docker
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (suite *DockerTestSuite) SetupSuite() {
	dockerRegistryContainer, registryID, docker := SetupDockerSocket()
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
		DockerTearDown(suite.LocalRegistry)
	} else {
		suite.LocalRegistry.StopRegistry()
	}
	purged, _ := suite.Docker.PurgeContainer("", REGISTRY)
	logrus.Infof("Purged containers %t", purged)
}

// --------------------------- TESTS -----------------

func (suite *DockerTestSuite) TestImagesOperationsOnDockerRegistryForTest() {
	registryContainer, err := GetRegistryContainer()
	assert.NotNil(suite.T(), registryContainer)
	assert.Nil(suite.T(), err)
	repos, err := registryContainer.GetRepositories()
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), len(repos) == 0)
	assert.Nil(suite.T(), suite.Docker.PullImage(TEST_IMAGE), "Pull image failed")
	assert.Nil(suite.T(), suite.Docker.TagImage(TEST_IMAGE_TAG, TEST_IMAGE_LOCAL_TAG), "Tag image failed")
	assert.Nil(suite.T(), suite.Docker.PushImage(TEST_IMAGE_LOCAL_TAG, REGISTRY_CONTAINER_URL_FROM_DOCKER_SOCKET, "", ""), "Push image in the DOcker container failed")
	//give the time to update the registry status
	time.Sleep(2 * time.Second)
	repos, err = registryContainer.GetRepositories()
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), len(repos) == 1)

	digest, erroDIgest := registryContainer.Connection.ManifestDigest(TEST_IMAGE, LATEST_TAG)
	assert.Nil(suite.T(), erroDIgest)
	assert.NotNil(suite.T(), digest)
	assert.NotNil(suite.T(), registryContainer.DeleteImage(TEST_IMAGE, LATEST_TAG), "Delete Image not allowed")
}
