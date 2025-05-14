package adapter

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAdapter struct {
	Ingress service.K8sResourcesIngress
	Secret  service.K8sResourcesSecret
	Issuer  service.K8sResourcesIssuer
	Cert    service.K8sResourcesCert
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

func (r ResourceAdapter) GetSecret(key client.ObjectKey) (*corev1.Secret, error) {
	return r.Secret.GetSecret(key)
}

func (r ResourceAdapter) GetService(name string) (*corev1.Service, error) {
	return r.Ingress.GetService(name)
}

func (r ResourceAdapter) GetBackendPort(svc *corev1.Service) int32 {
	return r.Ingress.GetBackendPort(svc)
}

func (r ResourceAdapter) GetDefaultBackendPort(svc *corev1.Service) int32 {
	return r.GetDefaultBackendPort(svc)
}

func (r ResourceAdapter) GetUpstreamConfig() ([]*ingress.Backends, error) {
	return r.Ingress.GetUpstreamConfig()
}

func (r ResourceAdapter) GetRules() []v1.IngressRule {
	return r.Ingress.GetRules()
}

func (r ResourceAdapter) CheckCert() error {
	return r.Cert.CheckCert()
}

func (r ResourceAdapter) CertObjectKey() string {
	return r.Cert.CertObjectKey()
}

func (r ResourceAdapter) SecretObjectKey() string {
	return r.Cert.SecretObjectKey()
}

func (r ResourceAdapter) GetTls() []v1.IngressTLS {
	return r.Ingress.GetTls()
}

func (r ResourceAdapter) CheckTlsHosts() bool {
	return r.Ingress.CheckTlsHosts()
}

func (r ResourceAdapter) GetHosts() []string {
	return r.Ingress.GetHosts()
}

func (r ResourceAdapter) GetAnnotations() map[string]string {
	return r.Ingress.GetAnnotations()
}

func (r ResourceAdapter) GetBackendName(bk *v1.ServiceBackendPort) string {
	return r.Ingress.GetBackendName(bk)
}

func (r ResourceAdapter) GetPaths() []string {
	return r.Ingress.GetPaths()
}

func (r ResourceAdapter) GetTlsFile() (map[string]ingress.Tls, error) {
	return r.Secret.GetTlsFile()
}

func (r ResourceAdapter) GetPathType(name string) (string, error) {
	return r.Ingress.GetPathType(name)
}
