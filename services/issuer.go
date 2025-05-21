package services

import (
	"context"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// IssuerServiceImpl 实现 IssuerService 接口
type IssuerServiceImpl struct {
	ing  common.Generic
	ctx  context.Context
	cert service.K8sResourcesCert
}

// NewIssuerServiceImpl 创建 Service 实例
func NewIssuerServiceImpl(ctx context.Context, clientSet common.Generic, cert service.K8sResourcesCert) *IssuerServiceImpl {
	return &IssuerServiceImpl{ctx: ctx, ing: clientSet, cert: cert}
}

func (i *IssuerServiceImpl) issuerGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "issuers"}
}

func (i *IssuerServiceImpl) issuerUnstructuredData() *unstructured.Unstructured {
	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Issuer",
			"metadata": map[string]interface{}{
				"name":      i.cert.IssuerObjectKey(),
				"namespace": i.ing.GetNameSpace(),
			},
			"spec": map[string]interface{}{
				"selfSigned": map[string]interface{}{},
			},
		},
	}

	return issuer
}

func (i *IssuerServiceImpl) GetIssuer(ctx context.Context, namespace, name string) error {
	if _, err := i.ing.GetDynamicClientSet().Resource(i.issuerGVR()).Namespace(i.ing.GetNameSpace()).Get(ctx, i.cert.IssuerObjectKey(), metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			if err := i.CreateIssuer(ctx, namespace, name); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

func (i *IssuerServiceImpl) CreateIssuer(ctx context.Context, namespace, name string) error {
	if _, err := i.ing.GetDynamicClientSet().Resource(i.issuerGVR()).Namespace(i.ing.GetNameSpace()).Create(ctx, i.issuerUnstructuredData(), metav1.CreateOptions{}); err != nil {
		klog.Error(fmt.Sprintf("CreateIssuer error %v", err))
		return err
	}

	return nil
}

func (i *IssuerServiceImpl) DeleteIssuer(ctx context.Context, namespace, name string) error {
	if err := i.ing.GetDynamicClientSet().Resource(i.issuerGVR()).Namespace(i.ing.GetNameSpace()).Delete(ctx, i.cert.IssuerObjectKey(), metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (i *IssuerServiceImpl) UpdateIssuer(ctx context.Context, namespace, name string) error {
	if _, err := i.ing.GetDynamicClientSet().Resource(i.issuerGVR()).Namespace(i.ing.GetNameSpace()).Update(ctx, i.issuerUnstructuredData(), metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (i *IssuerServiceImpl) CheckIssuer() error {
	return i.GetIssuer(i.ctx, i.ing.GetNameSpace(), i.ing.GetName())
}
