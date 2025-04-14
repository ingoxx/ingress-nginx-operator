package service

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NginxTmplData interface {
	GetName() string
	GetNameSpace() string
	GetTlsData(key client.ObjectKey) (map[string][]byte, error)
	GetService(name string) (*corev1.Service, error)
	GetBackendPort(svc *corev1.Service) int32
	GetUpstreamName(paths []v1.HTTPIngressPath, ing interface{}) string
	GetSecret() (*corev1.Secret, error)
	GetDefaultBackendPort(svc *corev1.Service) int32
}
