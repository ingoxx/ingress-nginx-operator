package services

import (
	"context"
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IssuerServiceImpl 实现 IssuerService 接口
type IssuerServiceImpl struct {
	K8sClient   interfaces.K8sClientSet
	OperatorCli interfaces.OperatorClientSet
	ctx         context.Context
}

// NewIssuerServiceImpl 创建 Service 实例
func NewIssuerServiceImpl(ctx context.Context, client interfaces.K8sClientSet) *PodServiceImpl {
	return &PodServiceImpl{K8sClient: client}
}

func (i *IssuerServiceImpl) issuerGVK() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "issuers"}
}

func (i *IssuerServiceImpl) issuerUnstructured() *unstructured.Unstructured {
	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Issuer",
		},
	}

	return issuer
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
