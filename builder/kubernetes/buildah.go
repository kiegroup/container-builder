package kubernetes

import (
	"context"
	"fmt"
	"github.com/kiegroup/container-builder/api"
	"github.com/kiegroup/container-builder/util/defaults"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const BuildahPlatform = "BuildahPlatform"
const BuildahImage = "BuildahImage"

const (
	builderDir    = "/builder"
	builderVolume = "container-builder"
	ContextDir    = "context"
)

type registryConfigMap struct {
	fileName    string
	mountPath   string
	destination string
}

var (
	serviceCABuildahRegistryConfigMap = registryConfigMap{
		fileName:    "service-ca.crt",
		mountPath:   "/etc/containers/certs.d",
		destination: "service-ca.crt",
	}

	buildahRegistryConfigMaps = []registryConfigMap{
		serviceCABuildahRegistryConfigMap,
	}
)

var (
	plainDockerBuildahRegistrySecret = registrySecret{
		fileName:    corev1.DockerConfigKey,
		mountPath:   "/buildah/.docker",
		destination: "config.json",
	}
	standardDockerBuildahRegistrySecret = registrySecret{
		fileName:    corev1.DockerConfigJsonKey,
		mountPath:   "/buildah/.docker",
		destination: "config.json",
		refEnv:      "REGISTRY_AUTH_FILE",
	}

	buildahRegistrySecrets = []registrySecret{
		plainDockerBuildahRegistrySecret,
		standardDockerBuildahRegistrySecret,
	}
)

func addBuildahTaskToPod(ctx context.Context, c ctrl.Reader, build *api.Build, task *api.BuildahTask, pod *corev1.Pod) error {
	var bud []string

	bud = []string{
		"buildah",
		"bud",
		"--storage-driver=vfs",
	}

	if task.Platform != "" {
		bud = append(bud, []string{
			"--platform",
			task.Platform,
		}...)
	}

	bud = append(bud, []string{
		"--pull-always",
		"-f",
		"Dockerfile",
		"-t",
		task.Image,
		".",
	}...)

	push := []string{
		"buildah",
		"push",
		"--storage-driver=vfs",
		"--digestfile=/dev/termination-log",
		task.Image,
		"docker://" + task.Image,
	}

	if task.Verbose != nil && *task.Verbose {
		bud = append(bud[:2], append([]string{"--log-level=debug"}, bud[2:]...)...)
		push = append(push[:2], append([]string{"--log-level=debug"}, push[2:]...)...)
	}

	env := make([]corev1.EnvVar, 0)
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	if task.Registry.CA != "" {
		config, err := getRegistryConfigMap(ctx, c, build.Namespace, task.Registry.CA, buildahRegistryConfigMaps)
		if err != nil {
			return err
		}
		addRegistryConfigMap(task.Registry.CA, config, &volumes, &volumeMounts)
		// This is easier to use the --cert-dir option, otherwise Buildah defaults to looking up certificates
		// into a directory named after the registry address
		bud = append(bud[:2], append([]string{"--cert-dir=/etc/containers/certs.d"}, bud[2:]...)...)
		push = append(push[:2], append([]string{"--cert-dir=/etc/containers/certs.d"}, push[2:]...)...)
	}

	var auth string
	if task.Registry.Secret != "" {
		secret, err := getRegistrySecret(ctx, c, build.Namespace, task.Registry.Secret, buildahRegistrySecrets)
		if err != nil {
			return err
		}
		if secret == plainDockerBuildahRegistrySecret {
			// Handle old format and make it compatible with Buildah
			auth = "(echo '{ \"auths\": ' ; cat /buildah/.docker/config.json ; echo \"}\") > /tmp/.dockercfg"
			env = append(env, corev1.EnvVar{
				Name:  "REGISTRY_AUTH_FILE",
				Value: "/tmp/.dockercfg",
			})
		}
		addRegistrySecret(task.Registry.Secret, secret, &volumes, &volumeMounts, &env)
	}

	if task.Registry.Insecure {
		bud = append(bud[:2], append([]string{"--tls-verify=false"}, bud[2:]...)...)
		push = append(push[:2], append([]string{"--tls-verify=false"}, push[2:]...)...)
	}

	env = append(env, proxyFromEnvironment()...)

	args := []string{
		strings.Join(bud, " "),
		strings.Join(push, " "),
	}
	if auth != "" {
		args = append([]string{auth}, args...)
	}

	image := task.ExecutorImage
	if image == "" {
		image = fmt.Sprintf("%s:v%s", defaults.BuildahDefaultImageName, defaults.BuildahVersion)
	}

	container := corev1.Container{
		Name:            task.Name,
		Image:           image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c"},
		Args:            []string{strings.Join(args, " && ")},
		Env:             env,
		WorkingDir:      filepath.Join(builderDir, build.Name, ContextDir),
		VolumeMounts:    volumeMounts,
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, volumes...)

	addContainerToPod(build, container, pod)

	// Make sure there is one container defined
	pod.Spec.Containers = pod.Spec.InitContainers[len(pod.Spec.InitContainers)-1 : len(pod.Spec.InitContainers)]
	pod.Spec.InitContainers = pod.Spec.InitContainers[:len(pod.Spec.InitContainers)-1]

	return nil
}

func addRegistryConfigMap(name string, config registryConfigMap, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	*volumes = append(*volumes, corev1.Volume{
		Name: "registry-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  config.fileName,
						Path: config.destination,
					},
				},
			},
		},
	})

	*volumeMounts = append(*volumeMounts, corev1.VolumeMount{
		Name:      "registry-config",
		MountPath: config.mountPath,
		ReadOnly:  true,
	})
}

func getRegistryConfigMap(ctx context.Context, c ctrl.Reader, ns, name string, registryConfigMaps []registryConfigMap) (registryConfigMap, error) {
	config := corev1.ConfigMap{}
	err := c.Get(ctx, ctrl.ObjectKey{Namespace: ns, Name: name}, &config)
	if err != nil {
		return registryConfigMap{}, err
	}
	for _, k := range registryConfigMaps {
		if _, ok := config.Data[k.fileName]; ok {
			return k, nil
		}
	}
	return registryConfigMap{}, errors.New("unsupported registry config map")
}

func getRegistrySecret(ctx context.Context, c ctrl.Reader, ns, name string, registrySecrets []registrySecret) (registrySecret, error) {
	secret := corev1.Secret{}
	err := c.Get(ctx, ctrl.ObjectKey{Namespace: ns, Name: name}, &secret)
	if err != nil {
		return registrySecret{}, err
	}
	for _, k := range registrySecrets {
		if _, ok := secret.Data[k.fileName]; ok {
			return k, nil
		}
	}
	return registrySecret{}, errors.New("unsupported secret type for registry authentication")
}

func addRegistrySecret(name string, secret registrySecret, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount, env *[]corev1.EnvVar) {
	*volumes = append(*volumes, corev1.Volume{
		Name: "registry-secret",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: name,
				Items: []corev1.KeyToPath{
					{
						Key:  secret.fileName,
						Path: secret.destination,
					},
				},
			},
		},
	})

	*volumeMounts = append(*volumeMounts, corev1.VolumeMount{
		Name:      "registry-secret",
		MountPath: secret.mountPath,
		ReadOnly:  true,
	})

	if secret.refEnv != "" {
		*env = append(*env, corev1.EnvVar{
			Name:  secret.refEnv,
			Value: filepath.Join(secret.mountPath, secret.destination),
		})
	}
}

func addContainerToPod(build *api.Build, container corev1.Container, pod *corev1.Pod) {
	if hasBuilderVolume(pod) {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      builderVolume,
			MountPath: filepath.Join(builderDir, build.Name),
		})
	}

	pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)
}

func hasBuilderVolume(pod *corev1.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == builderVolume {
			return true
		}
	}
	return false
}
