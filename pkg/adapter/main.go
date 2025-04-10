package adapter

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAdapter struct {
	Ingress service.K8sResourcesIngress
	Secret  service.K8sResourcesSecret
}

func (r ResourceAdapter) GetName() string {
	return r.Ingress.GetName()
}

func (r ResourceAdapter) GetNameSpace() string {
	return r.Ingress.GetNameSpace()
}

func (r ResourceAdapter) GetTlsData(key client.ObjectKey) (map[string][]byte, error) {
	return r.Secret.GetTlsData(key)
}
func (r ResourceAdapter) GetService(name string) (*corev1.Service, error) {
	return r.Ingress.GetService(name)
}
func (r ResourceAdapter) GetBackendPort(data interface{}) (uint16, error) {
	return r.Ingress.GetBackendPort(data)
}
func (r ResourceAdapter) GetUpstreamName(paths []v1.HTTPIngressPath, ing interface{}) string {
	return r.Ingress.GetUpstreamName(paths, ing)
}

func (r ResourceAdapter) GetSecret() (*corev1.Secret, error) {
	return r.Secret.GetSecret()
}
