package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"
	"golang.org/x/net/context"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
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

func (i *IngressServiceImpl) GetIngressHosts(ctx context.Context, namespace, name string) []string {
	var hosts []string

	return hosts
}

func (i *IngressServiceImpl) GetIngressBackends(ctx context.Context, namespace, name string) ([]v12.Service, error) {
	var svs []v12.Service

	return svs, nil
}
