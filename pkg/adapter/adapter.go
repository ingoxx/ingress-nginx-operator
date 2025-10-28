package adapter

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAdapter struct {
	Ingress   service.K8sResourcesIngress
	Secret    service.K8sResourcesSecret
	Issuer    service.K8sResourcesIssuer
	Cert      service.K8sResourcesCert
	ConfigMap service.K8sResourceConfigMap
	Svc       service.K8sResourcesSvc
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

func (r ResourceAdapter) GetService(key client.ObjectKey) (*corev1.Service, error) {
	return r.Ingress.GetService(key)
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

func (r ResourceAdapter) GetConfigMapData(name string) ([]byte, error) {
	return r.ConfigMap.GetConfigMapData(name)
}

func (r ResourceAdapter) UpdateConfigMap(name, ns, key string, data []byte) (string, error) {
	return r.ConfigMap.UpdateConfigMap(name, ns, key, data)
}

func (r ResourceAdapter) GetNgxConfigMap(name string) (map[string]string, error) {
	return r.ConfigMap.GetNgxConfigMap(name)
}

func (r ResourceAdapter) GetAnyBackendName(svc *v1.ServiceBackendPort, namespace string) string {
	return r.Ingress.GetAnyBackendName(svc, namespace)
}

func (r ResourceAdapter) GetDaemonSetNameLabel() string {
	return r.Ingress.GetDaemonSetNameLabel()
}

func (r ResourceAdapter) GetDeployNameLabel() string {
	return r.Ingress.GetDeployNameLabel()
}

func (r ResourceAdapter) GetBackendPorts(key client.ObjectKey) ([]*v1.ServiceBackendPort, error) {
	return r.Ingress.GetBackendPorts(key)
}

func (r ResourceAdapter) GetDaemonSvcName() string {
	return r.Ingress.GetDaemonSvcName()
}

func (r ResourceAdapter) GetDeploySvcName() string {
	return r.Ingress.GetDeploySvcName()
}

func (r ResourceAdapter) GetDaemonSetLabel() string {
	return r.Ingress.GetDaemonSetLabel()
}

func (r ResourceAdapter) GetDeployLabel() string {
	return r.Ingress.GetDeployLabel()
}

func (r ResourceAdapter) GetDefaultBackend() (*v1.ServiceBackendPort, error) {
	return r.Ingress.GetDefaultBackend()
}

func (r ResourceAdapter) CheckDefaultBackend() error {
	return r.Ingress.CheckDefaultBackend()
}

func (r ResourceAdapter) CheckHost(host string) bool {
	return r.Ingress.CheckHost(host)
}

func (r ResourceAdapter) UpdateIngress(ing *v1.Ingress) error {
	return r.Ingress.UpdateIngress(ing)
}

func (r ResourceAdapter) GetCmName() string {
	return r.ConfigMap.GetCmName()
}
