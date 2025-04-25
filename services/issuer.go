package services

import (
	"context"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IssuerServiceImpl 实现 IssuerService 接口
type IssuerServiceImpl struct {
	ing common.Generic
	ctx context.Context
	as  service.CertAssociationResources
}

// NewIssuerServiceImpl 创建 Service 实例
func NewIssuerServiceImpl(ctx context.Context, clientSet common.Generic, as service.CertAssociationResources) *IssuerServiceImpl {
	return &IssuerServiceImpl{ctx: ctx, ing: clientSet, as: as}
}

func (i *IssuerServiceImpl) issuerGVK() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "issuers"}
}

func (i *IssuerServiceImpl) issuerUnstructuredData() *unstructured.Unstructured {
	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Issuer",
			"metadata": map[string]interface{}{
				"name":      i.as.IssuerObjectKey(),
				"namespace": i.ing.GetNameSpace(),
			},
			"spec": map[string]interface{}{
				"selfSigned": map[string]interface{}{},
			},
		},
	}

	return issuer
}

func (i *IssuerServiceImpl) CreateIssuer(ctx context.Context, namespace, name string) error {
	if _, err := i.ing.GetDynamicClientSet().Resource(i.issuerGVK()).Create(ctx, i.issuerUnstructuredData(), metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (i *IssuerServiceImpl) GetIssuer(ctx context.Context, namespace, name string) error {
	if _, err := i.ing.GetDynamicClientSet().Resource(i.issuerGVK()).Get(ctx, i.as.IssuerObjectKey(), metav1.GetOptions{}); err != nil {
		if err := i.CreateIssuer(ctx, namespace, name); err != nil {
			return err
		}
	}

	return nil
}

func (i *IssuerServiceImpl) DeleteIssuer(ctx context.Context, namespace, name string) error {
	if err := i.ing.GetDynamicClientSet().Resource(i.issuerGVK()).Delete(ctx, i.as.IssuerObjectKey(), metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (i *IssuerServiceImpl) UpdateIssuer(ctx context.Context, namespace, name string) error {
	if _, err := i.ing.GetDynamicClientSet().Resource(i.issuerGVK()).Update(ctx, i.issuerUnstructuredData(), metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (i *IssuerServiceImpl) CheckIssuer() error {
	return i.GetIssuer(i.ctx, i.ing.GetNameSpace(), i.ing.GetName())
}
