package service

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourcesSvc interface {
	GetSvc(key client.ObjectKey) (*corev1.Service, error)
	GetAllEndPoints() ([]string, error)
}
