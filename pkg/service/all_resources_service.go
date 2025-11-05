package service

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourcesMth interface {
	GetName() string
	GetNameSpace() string
	GetTlsData(client.ObjectKey) (map[string][]byte, error)
	GetService(client.ObjectKey) (*corev1.Service, error)
	GetBackendPort(*corev1.Service) int32
	GetUpstreamConfig() ([]*ingress.Backends, error)
	GetSecret(client.ObjectKey) (*corev1.Secret, error)
	GetDefaultBackendPort(*corev1.Service) int32
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
	GetConfigMapData(string) ([]byte, error)
	GetAnyBackendName(*v1.ServiceBackendPort, string) string
	GetDaemonSetNameLabel() string
	GetDeployNameLabel() string
	GetBackendPorts(client.ObjectKey) ([]*v1.ServiceBackendPort, error)
	GetDeploySvcName() string
	GetDaemonSvcName() string
	GetDaemonSetLabel() string
	GetDeployLabel() string
	GetDefaultBackend() (*v1.ServiceBackendPort, error)
	CheckDefaultBackend() error
	CheckHost(string) bool
	UpdateConfigMap(name, ns, key string, data []byte) (string, error)
	GetNgxConfigMap(name string) (map[string]string, error)
	UpdateIngress(ing *v1.Ingress) error
	GetCmName() string
	GetAllEndPoints() ([]string, error)
	NewIngress(ing *v1.Ingress)
	GetCm() (*corev1.ConfigMap, error)
	ClearCmData(string) error
}
