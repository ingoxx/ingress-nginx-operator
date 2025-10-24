package service

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourcesIngress interface {
	GetIngress(ctx context.Context, req client.ObjectKey) (*v1.Ingress, error)
	GetHosts() []string
	GetBackend(string) (*v1.ServiceBackendPort, error)
	GetDefaultBackend() (*v1.ServiceBackendPort, error)
	GetService(client.ObjectKey) (*corev1.Service, error)
	GetBackendPort(*corev1.Service) int32
	GetDefaultBackendPort(*corev1.Service) int32
	GetUpstreamConfig() ([]*ingress.Backends, error)
	CheckController() error
	GetAnnotations() map[string]string
	CheckService() error
	CheckHosts() error
	CheckPath([]v1.HTTPIngressPath) error
	GetName() string
	GetNameSpace() string
	GetRules() []v1.IngressRule
	GetTls() []v1.IngressTLS
	CheckTlsHosts() bool
	GetIngressObjectMate() metav1.ObjectMeta
	GetBackendName(*v1.ServiceBackendPort) string
	GetPaths() []string
	GetPathType(string) (string, error)
	GetAnyBackendName(*v1.ServiceBackendPort, string) string
	GetDaemonSetNameLabel() string
	GetDeployNameLabel() string
	GetBackendPorts(client.ObjectKey) ([]*v1.ServiceBackendPort, error)
	GetDaemonSvcName() string
	GetDeploySvcName() string
	GetDaemonSetLabel() string
	GetDeployLabel() string
	CheckDefaultBackend() error
	CheckHost(string) bool
	UpdateIngress(*v1.Ingress) error
}
