package kubernetes

import (
	"github.com/kiegroup/container-builder/util"
	corev1 "k8s.io/api/core/v1"
)

func PodSecurityDefaults() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsNonRoot: util.Pbool(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}
