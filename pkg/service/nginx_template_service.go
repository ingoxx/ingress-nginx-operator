package service

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NginxTemplateData interface {
	GetName() string
	GetNameSpace() string
	GetTlsData(key client.ObjectKey) (map[string][]byte, error)
	GetService(name string) (*corev1.Service, error)
	GetBackendPort(svc *corev1.Service) int32
	GetUpstreamConfig() ([]*ingress.Backends, error)
	GetSecret(key client.ObjectKey) (*corev1.Secret, error)
	GetDefaultBackendPort(svc *corev1.Service) int32
	GetRules() []v1.IngressRule
	CertObjectKey() string
	SecretObjectKey() string
	GetTls() []v1.IngressTLS
	CheckTlsHosts() bool
	GetHosts() []string
	GetAnnotations() map[string]string
	GetBackendName(*v1.ServiceBackendPort) string
	GetPaths() []string
	GetTlsFile() (map[string]ingress.Tls, error)
	GetPathType(string) (string, error)
}
