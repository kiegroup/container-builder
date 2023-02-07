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

	builder "github.com/kiegroup/container-builder/builder/kubernetes"

	v1 "k8s.io/api/core/v1"
	resource2 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/container-builder/api"
	"github.com/kiegroup/container-builder/client"
)

/*
Usage example. Please note that you must have a valid Kubernetes environment up and running with the kogito-builder namespace
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

	argsWithoutProg := os.Args[1:]
	builderSelection := 1

	var registryAddress, registrySecret string
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "buildah" {
			builderSelection = 2
		}
		if len(argsWithoutProg) > 2 {
			registryAddress = argsWithoutProg[1]
			registrySecret = argsWithoutProg[2]
			fmt.Printf("Configured registry %s and secret %s.\n", registryAddress, registrySecret)
		}
	}

	var build *api.Build
	switch builderSelection {

	case 1:
		fmt.Println("Generating Kaniko build")
		build = generateKanikoBuild(err, dockerFile, source, cli)
		break
	case 2:
		fmt.Println("Generating Buildah build")
		build = generateBuildahBuild(err, dockerFile, source, cli, registryAddress, registrySecret)
		break

	}

	// from now the Reconcile method can be called until the build is finished
	for build.Status.Phase != api.BuildPhaseSucceeded &&
		build.Status.Phase != api.BuildPhaseError &&
		build.Status.Phase != api.BuildPhaseFailed {
		fmt.Printf("\nBuild status is %s", build.Status.Phase)
		build, err = builder.FromBuild(build).WithClient(cli).Reconcile()
		if err != nil {
			fmt.Println("Failed to run test")
			panic(fmt.Errorf("Build %v just failed", build))
		}
		time.Sleep(10 * time.Second)
	}

}

func generateBuildahBuild(err error, dockerFile []byte, source []byte, cli client.Client, address string, secret string) *api.Build {
	buildahPlatform := api.PlatformBuild{
		ObjectReference: api.ObjectReference{
			Namespace: "kogito-builder",
			Name:      "testBuildahPlatform",
		},
		Spec: api.PlatformBuildSpec{
			BuildStrategy:   api.BuildStrategyPod,
			PublishStrategy: api.PlatformBuildPublishStrategyBuildah,
			Registry: api.RegistrySpec{
				Insecure: true,
			},
			Timeout: &metav1.Duration{
				Duration: 5 * time.Minute,
			},
		},
	}
	if address != "" && secret != "" {
		buildahPlatform.Spec.Registry.Address = address
		buildahPlatform.Spec.Registry.Secret = secret
	}

	buildahBuild, err := builder.NewBuild(builder.BuilderInfo{FinalImageName: "greetings:latest", BuildUniqueName: "kogito-buildah-test", Platform: buildahPlatform}).
		WithResource("Dockerfile", dockerFile).WithResource("greetings.sw.json", source).
		WithClient(cli).
		Schedule()
	if err != nil {
		fmt.Println(err.Error())
		panic("Can't create kanikoBuild")
	}
	return buildahBuild
}

func generateKanikoBuild(err error, dockerFile []byte, source []byte, cli client.Client) *api.Build {
	kanikoPlatform := api.PlatformBuild{
		ObjectReference: api.ObjectReference{
			Namespace: "kogito-builder",
			Name:      "testBuildahPlatform",
		},
		Spec: api.PlatformBuildSpec{
			BuildStrategy:   api.BuildStrategyPod,
			PublishStrategy: api.PlatformBuildPublishStrategyBuildah,
			Registry: api.RegistrySpec{
				Insecure: true,
			},
			Timeout: &metav1.Duration{
				Duration: 5 * time.Minute,
			},
		},
	}

	cpuQty, _ := resource2.ParseQuantity("1")
	memQty, _ := resource2.ParseQuantity("4Gi")

	kanikoBuild, err := builder.NewBuild(builder.BuilderInfo{FinalImageName: "greetings:latest", BuildUniqueName: "kogito-test", Platform: kanikoPlatform}).
		WithResource("Dockerfile", dockerFile).WithResource("greetings.sw.json", source).
		WithAdditionalArgs([]string{"--build-arg=QUARKUS_PACKAGE_TYPE=mutable-jar", "--build-arg=QUARKUS_LAUNCH_DEVMODE=true", "--build-arg=SCRIPT_DEBUG=false"}).
		WithResourceRequirements(v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    cpuQty,
				v1.ResourceMemory: memQty,
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    cpuQty,
				v1.ResourceMemory: memQty,
			},
		}).
		WithClient(cli).
		Schedule()
	if err != nil {
		fmt.Println(err.Error())
		panic("Can't create kanikoBuild")
	}
	return kanikoBuild
}
