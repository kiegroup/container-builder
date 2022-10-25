package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

// IsOptionEnabled return whether if the PublishStrategyOptions is enabled or not
func (b PlatformBuildSpec) IsOptionEnabled(option string) bool {
	//Key defined in builder/kaniko.go
	if enabled, ok := b.PublishStrategyOptions[option]; ok {
		res, err := strconv.ParseBool(enabled)
		if err != nil {
			return false
		}
		return res
	}
	return false
}

// GetTimeout returns the specified duration or a default one
func (b PlatformBuildSpec) GetTimeout() metav1.Duration {
	if b.Timeout == nil {
		return metav1.Duration{}
	}
	return *b.Timeout
}
