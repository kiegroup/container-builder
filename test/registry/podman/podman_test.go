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
package registry

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
type PodmanTestSuite struct {
	suite.Suite
	LocalRegistry cleaner.PodmanLocalRegistry
	RegistryID    string
	Podman        cleaner.Podman
}

func TestPodmanTestSuite(t *testing.T) {
	suite.Run(t, new(PodmanTestSuite))
}

func (suite *PodmanTestSuite) SetupSuite() {
	localRegistry, registryID, podman := cleaner.SetupPodmanSocket()
	if len(registryID) > 0 {
		suite.LocalRegistry = localRegistry
		suite.RegistryID = registryID
		suite.Podman = podman
	} else {
		assert.FailNow(suite.T(), "Initialization failed")
	}
}

func (suite *PodmanTestSuite) TearDownSuite() {
	registryID := suite.LocalRegistry.GetRegistryRunningID()
	if len(registryID) > 0 {
		cleaner.PodmanTearDown(suite.LocalRegistry)
	} else {
		suite.LocalRegistry.StopRegistry()
	}
	suite.Podman.PurgeContainer("", cleaner.REGISTRY_FULL)
}

// -------------------------------------- TESTS -----------------------------

func (suite *PodmanTestSuite) TestRegistry() {
	assert.Truef(suite.T(), suite.RegistryID != "", "Registry not started")
	assert.Truef(suite.T(), suite.LocalRegistry.IsRegistryImagePresent(), "Registry image not present")
	assert.Truef(suite.T(), suite.LocalRegistry.IsRegistryRunning(), "Registry container not running")
}

func (suite *PodmanTestSuite) TestPullTagPush() {

	assert.Truef(suite.T(), suite.RegistryID != "", "Registry not started")
	registryContainer, err := cleaner.GetRegistryContainer()
	assert.Nil(suite.T(), err)
	repos := test.CheckRepositoriesSize(suite.T(), 0, registryContainer)
	logrus.Info("Empty Repo Size = ", len(repos))

	result := podmanPullTagPushOnRegistryContainer(suite)
	logrus.Info("Pull Tag and Push Image on Registry Container successful = ", result)

	// Give some time to the registry to refresh status
	time.Sleep(2 * time.Second)
	repos = test.CheckRepositoriesSize(suite.T(), 1, registryContainer)
	logrus.Info("Repo Size after pull image = ", len(repos))
}

func podmanPullTagPushOnRegistryContainer(suite *PodmanTestSuite) bool {
	podmanSocketConn, errSock := cleaner.GetPodmanConnection()
	if errSock != nil {
		assert.FailNow(suite.T(), "Cant get podman socket")
	}
	p := cleaner.Podman{Connection: podmanSocketConn}
	if !suite.LocalRegistry.IsRegistryImagePresent() {
		_, err := p.PullImage(cleaner.TEST_IMAGE_TAG)
		if err != nil {
			assert.Fail(suite.T(), "Pull Image Failed", err)
			return false
		}
	}
	logrus.Info("Pull image")

	err := p.TagImage(cleaner.TEST_REPO+cleaner.TEST_IMAGE_TAG, cleaner.LATEST_TAG, cleaner.TEST_REGISTRY_REPO+cleaner.TEST_IMAGE)
	if err != nil {
		assert.Fail(suite.T(), "Tag Image Failed", err)
		return false
	}
	logrus.Info("Tag image")

	err = p.PushImage(cleaner.TEST_IMAGE_LOCAL_TAG, cleaner.TEST_IMAGE_LOCAL_TAG, "", "")
	if err != nil {
		assert.Fail(suite.T(), "Push Image Failed", err)
		return false
	}
	logrus.Info("Push image")
	return true
}
