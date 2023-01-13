//go:build integration_kaniko_docker

/*
 * Copyright 2023 Red Hat, Inc. and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package builder

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

func TestKanikoTestSuite(t *testing.T) {
	suite.Run(t, new(KanikoDockerTestSuite))
}

func (suite *KanikoDockerTestSuite) TestKanikoBuild() {

	mydir, err := os.Getwd()
	if err != nil {
		logrus.Error(err)
	}
	dockefileDir := mydir + "/../examples/dockerfiles"
	assert.Nil(suite.T(), suite.Docker.PullImage("gcr.io/kaniko-project/executor:latest"), "Pull image failed")
	config := KanikoVanillaConfig{
		DockerFilePath:         dockefileDir,
		VerbosityLevel:         "info",
		KanikoExecutorImage:    EXECUTOR_IMAGE,
		ContainerName:          "kaniko-build",
		DockerFileName:         "Kogito.dockerfile",
		RegistryFinalImageName: "localhost:5000/kaniko-test/kaniko-dockerfile_test_swf",
		ReadBuildOutput:        false,
	}
	logrus.Infof("Start Kaniko build")
	start := time.Now()
	imageID, error := KanikoBuild(suite.Docker.Connection, config)
	timeElapsed := time.Since(start)
	logrus.Infof("The Kaniko build took %s", timeElapsed)
	assert.Nil(suite.T(), error, "Build failed")
	assert.NotNil(suite.T(), imageID, error, "Build failed")
}
