package services

import (
	"context"
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"
)

// IssuerServiceImpl 实现 IssuerService 接口
type IssuerServiceImpl struct {
	K8sClient interfaces.K8sClientSet
}

// NewIssuerServiceImpl 创建 Service 实例
func NewIssuerServiceImpl(client interfaces.K8sClientSet) *PodServiceImpl {
	return &PodServiceImpl{K8sClient: client}
}

func (i *IssuerServiceImpl) CreateIssuer(ctx context.Context, namespace, name string) error {
	return nil
}

func (i *IssuerServiceImpl) GetIssuer(ctx context.Context, namespace, name string) error {
	return nil
}

func (i *IssuerServiceImpl) DeleteIssuer(ctx context.Context, namespace, name string) error {
	return nil
}

func (i *IssuerServiceImpl) UpdateIssuer(ctx context.Context, namespace, name string) error {
	return nil
}
