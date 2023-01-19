package kubernetes

import (
	"context"
	"os"
	"strings"

	"github.com/kiegroup/container-builder/client"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/container-builder/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type registrySecret struct {
	fileName    string
	mountPath   string
	destination string
	refEnv      string
}

func newBuildPod(ctx context.Context, c client.Client, build *api.Build) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: build.Namespace,
			Name:      buildPodName(build),
			Labels: map[string]string{
				"kie.kogito.org/buildContext": build.Name,
				"kie.kogito.org/component":    "builder",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	for _, task := range build.Spec.Tasks {
		switch {
		case task.Kaniko != nil:
			err := addKanikoTaskToPod(ctx, c, build, task.Kaniko, pod)
			if err != nil {
				return nil, err
			}
		case task.Buildah != nil:
			err := addBuildahTaskToPod(ctx, c, build, task.Buildah, pod)
			if err != nil {
				return nil, err
			}
		}
	}

	return pod, nil
}

func buildPodName(build *api.Build) string {
	return "kogito-" + strings.ToLower(build.Name) + "-builder"
}

func getBuilderPod(ctx context.Context, c client.Client, build *api.Build) (*corev1.Pod, error) {
	pod := corev1.Pod{}
	err := c.Get(ctx, types.NamespacedName{Name: buildPodName(build), Namespace: build.Namespace}, &pod)
	if err != nil && k8serrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pod, nil
}

func deleteBuilderPod(ctx context.Context, c client.Client, build *api.Build) error {
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: build.Namespace,
			Name:      buildPodName(build),
		},
	}

	err := c.Delete(ctx, &pod)
	if err != nil && k8serrors.IsNotFound(err) {
		return nil
	}

	return err
}

func proxyFromEnvironment() []corev1.EnvVar {
	var envVars []corev1.EnvVar

	if httpProxy, ok := os.LookupEnv("HTTP_PROXY"); ok {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "HTTP_PROXY",
			Value: httpProxy,
		})
	}

	if httpsProxy, ok := os.LookupEnv("HTTPS_PROXY"); ok {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "HTTPS_PROXY",
			Value: httpsProxy,
		})
	}

	if noProxy, ok := os.LookupEnv("NO_PROXY"); ok {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "NO_PROXY",
			Value: noProxy,
		})
	}

	return envVars
}
