package kubernetes

import (
	"github.com/kiegroup/container-builder/util"
	corev1 "k8s.io/api/core/v1"
)

// This is the security defualt to run without privileges
func ContainerSecurityDefaults() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: util.Pbool(false),
		Privileged:               util.Pbool(false),
		RunAsNonRoot:             util.Pbool(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{corev1.Capability("ALL")},
		},
	}
}
