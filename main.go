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
Example of usage
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

	build, err := builder.NewBuild(platform, "greetings:latest", "kogito-test").
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
