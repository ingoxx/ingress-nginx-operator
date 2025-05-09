package service

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourcesSecret interface {
	GetTlsData(key client.ObjectKey) (map[string][]byte, error)
	GetSecret(key client.ObjectKey) (*corev1.Secret, error)
	GetTlsFile() (map[string]ingress.Tls, error)
}
