package api

import (
	"k8s.io/apimachinery/pkg/types"
)

// ObjectReference is a subset of the kubernetes k8s.io/apimachinery/pkg/apis/meta/v1.Object interface.
// Objects in this API not necessarily represent Kubernetes objects, but this structure can help when needed.
type ObjectReference struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (o *ObjectReference) GetName() string {
	return o.Name
}

func (o *ObjectReference) GetNamespace() string {
	return o.Namespace
}

func (o *ObjectReference) GetObjectKey() types.NamespacedName {
	return types.NamespacedName{Name: o.Name, Namespace: o.Namespace}
}
