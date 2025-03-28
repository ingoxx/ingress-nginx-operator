package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IngressServiceImpl 实现 IngressService 接口
type IngressServiceImpl struct {
	K8sClient interfaces.K8sClientSet
}

// NewIngressServiceImpl 创建 Service 实例
func NewIngressServiceImpl(client interfaces.K8sClientSet) *PodServiceImpl {
	return &PodServiceImpl{K8sClient: client}
}

func (i *IngressServiceImpl) GetIngress(ctx context.Context, namespace, name string) (v1.Ingress, error) {
	var ing v1.Ingress
	return ing, nil
}

func (i *IngressServiceImpl) GetHosts(ctx context.Context, namespace, name string) []string {
	var hosts []string

	return hosts
}

func (i *IngressServiceImpl) GetBackends(ctx context.Context, namespace, name string) ([]v1.IngressBackend, error) {
	var bks []v1.IngressBackend

	return bks, nil
}

func (i *IngressServiceImpl) GetBackend(ctx context.Context, namespace, name string) (v1.IngressBackend, error) {
	var bks v1.IngressBackend

	return bks, nil
}

func (i *IngressServiceImpl) GetDefaultService() (*corev1.Service, error) {
	var svc *corev1.Service
	return svc, nil
}

func (i *IngressServiceImpl) GetService(name string) (*corev1.Service, error) {
	var svc *corev1.Service
	return svc, nil
}

func (i *IngressServiceImpl) GetBackendPort(data interface{}) (uint16, error) {
	return 0, nil
}

func (i *IngressServiceImpl) GetSecret(key client.ObjectKey) (*corev1.Secret, error) {
	var sec *corev1.Secret
	return sec, nil
}

func (i *IngressServiceImpl) GetTlsData(key client.ObjectKey) (map[string][]byte, error) {
	return nil, nil
}

func (i *IngressServiceImpl) GetUpstreamName(paths []v1.HTTPIngressPath, ing interface{}) string {
	return ""
}

func (i *IngressServiceImpl) getUpstreamBackend(paths []v1.HTTPIngressPath) string {
	return ""
}
