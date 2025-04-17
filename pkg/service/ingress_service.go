package service

import (
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourcesIngress interface {
	GetIngress(ctx context.Context, req client.ObjectKey) (*v1.Ingress, error)
	GetHosts(namespace, name string) []string
	GetBackend(name string) (*v1.ServiceBackendPort, error)
	GetDefaultBackend() (*v1.ServiceBackendPort, error)
	GetService(name string) (*corev1.Service, error)
	GetBackendPort(svc *corev1.Service) int32
	GetDefaultBackendPort(svc *corev1.Service) int32
	GetUpstreamName(paths []v1.HTTPIngressPath, ing interface{}) string
	CheckController() error
	CheckService() error
	CheckHost(host v1.IngressRule) error
	CheckPath(path v1.HTTPIngressPath) error
	GetName() string
	GetNameSpace() string
}
