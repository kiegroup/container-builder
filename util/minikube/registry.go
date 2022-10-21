// Package minikube contains utilities for Minikube deployments
package minikube

import (
	"context"
	"strconv"

	"github.com/ricardozanini/container-builder/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	registryNamespace = "kube-system"
)

// FindRegistry returns the Minikube addon registry location if any.
func FindRegistry(ctx context.Context, c client.Client) (*string, error) {
	svcs := corev1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
	}
	err := c.List(ctx, &svcs,
		k8sclient.InNamespace(registryNamespace),
		k8sclient.MatchingLabels{
			"kubernetes.io/minikube-addons": "registry",
		})
	if err != nil {
		return nil, err
	}
	if len(svcs.Items) == 0 {
		return nil, nil
	}
	svc := svcs.Items[0]
	ip := svc.Spec.ClusterIP
	portStr := ""
	if len(svc.Spec.Ports) > 0 {
		port := svc.Spec.Ports[0].Port
		if port > 0 && port != 80 {
			portStr = ":" + strconv.FormatInt(int64(port), 10)
		}
	}
	registry := ip + portStr
	return &registry, nil
}
