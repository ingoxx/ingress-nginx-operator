package service

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourcesSecret interface {
	GetTlsData(key client.ObjectKey) (map[string][]byte, error)
	GetSecret() (*corev1.Secret, error)
}
