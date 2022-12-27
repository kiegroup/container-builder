/*
 * Copyright 2022 Red Hat, Inc. and/or its affiliates.
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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kiegroup/container-builder/api"
	"github.com/kiegroup/container-builder/builder"
	"github.com/kiegroup/container-builder/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
Usage example
*/

func main() {
	cli, err := client.NewOutOfClusterClient("")

	dockerFile, err := os.ReadFile("examples/dockerfiles/Kogito.dockerfile")
	if err != nil {
		panic("Can't read dockerfile")
	}
	source, err := os.ReadFile("examples/sources/kogitogreetings.sw.json")
	if err != nil {
		panic("Can't read source file")
	}

	if err != nil {
		fmt.Println("Failed to create client")
		fmt.Println(err.Error())
	}
	platform := api.PlatformBuild{
		ObjectReference: api.ObjectReference{
			Namespace: "kogito-builder",
			Name:      "testPlatform",
		},
		Spec: api.PlatformBuildSpec{
			BuildStrategy:   api.BuildStrategyPod,
			PublishStrategy: api.PlatformBuildPublishStrategyKaniko,
			Registry: api.RegistrySpec{
				Insecure: true,
			},
			Timeout: &metav1.Duration{
				Duration: 5 * time.Minute,
			},
		},
	}

	build, err := builder.NewScheduler(builder.BuilderInfo{FinalImageName: "greetings:latest", BuildUniqueName: "kogito-test", Platform: platform}).
		WithResource("Dockerfile", dockerFile).WithResource("greetings.sw.json", source).
		WithClient(cli).
		Schedule()
	if err != nil {
		fmt.Println(err.Error())
		panic("Can't create build")
	}

	// from now the Reconcile method can be called until the build is finished
	for build.Status.Phase != api.BuildPhaseSucceeded &&
		build.Status.Phase != api.BuildPhaseError &&
		build.Status.Phase != api.BuildPhaseFailed {
		fmt.Printf("\nBuild status is %s", build.Status.Phase)
		build, err = builder.FromBuild(build).WithClient(cli).Reconcile()
		if err != nil {
			fmt.Println("Failed to run test")
			panic(fmt.Errorf("build %v just failed", build))
		}
		time.Sleep(10 * time.Second)
	}

}
