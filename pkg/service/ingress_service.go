package service

import (
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourcesIngress interface {
	GetIngress(ctx context.Context, namespace, name string) (v1.Ingress, error)
	GetHosts(ctx context.Context, namespace, name string) []string
	GetBackends(ctx context.Context, namespace, name string) ([]v1.IngressBackend, error)
	GetBackend(ctx context.Context, namespace, name string) (v1.IngressBackend, error)
	GetDefaultService() (*corev1.Service, error)
	GetService(name string) (*corev1.Service, error)
	GetBackendPort(data interface{}) (uint16, error)
	GetSecret(key client.ObjectKey) (*corev1.Secret, error)
	GetTlsData(key client.ObjectKey) (map[string][]byte, error)
	GetUpstreamName(paths []v1.HTTPIngressPath, ing interface{}) string
	CheckController() error
	CheckService() error
}
