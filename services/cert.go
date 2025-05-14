package services

import (
	"context"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"slices"
)

// CertServiceImpl 实现 CertService 接口
type CertServiceImpl struct {
	issuer service.K8sResourcesIssuer
	ing    common.Generic
	ctx    context.Context
}

// NewCertServiceImpl 创建 Service 实例
func NewCertServiceImpl(ctx context.Context, ing common.Generic) *CertServiceImpl {
	c := &CertServiceImpl{ctx: ctx, ing: ing}
	c.issuer = NewIssuerServiceImpl(ctx, ing, c)

	return c
}

func (c *CertServiceImpl) certGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
}

func (c *CertServiceImpl) certUnstructuredData() *unstructured.Unstructured {
	certUnstructured := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      c.CertObjectKey(),
				"namespace": c.ing.GetNameSpace(),
			},
			"spec": map[string]interface{}{
				"dnsNames": c.ing.GetHosts(),
				"issuerRef": map[string]interface{}{
					"kind": "Issuer",
					"name": c.IssuerObjectKey(),
				},
				"secretName": c.SecretObjectKey(),
			},
		},
	}

	return certUnstructured
}

func (c *CertServiceImpl) CertObjectKey() string {
	return c.ing.GetName() + "-" + c.ing.GetNameSpace() + "-cert"
}

func (c *CertServiceImpl) SecretObjectKey() string {
	return c.ing.GetName() + "-" + c.ing.GetNameSpace() + "-secret"
}

func (c *CertServiceImpl) IssuerObjectKey() string {
	return c.ing.GetName() + "-" + c.ing.GetNameSpace() + "-issuer"
}

func (c *CertServiceImpl) GetCert() (*unstructured.Unstructured, error) {
	cert, err := c.ing.GetDynamicClientSet().Resource(c.certGVR()).Namespace(c.ing.GetNameSpace()).Get(c.ctx, c.CertObjectKey(), metav1.GetOptions{})
	if err != nil {
		return cert, err
	}

	return cert, nil
}

func (c *CertServiceImpl) CreateCert() (*unstructured.Unstructured, error) {
	cert, err := c.ing.GetDynamicClientSet().Resource(c.certGVR()).Namespace(c.ing.GetNameSpace()).Create(c.ctx, c.certUnstructuredData(), metav1.CreateOptions{})
	if err != nil {
		return cert, err
	}

	return cert, nil
}

func (c *CertServiceImpl) DeleteCert() error {
	return c.ing.GetDynamicClientSet().Resource(c.certGVR()).Namespace(c.ing.GetNameSpace()).Delete(c.ctx, c.CertObjectKey(), metav1.DeleteOptions{})
}

func (c *CertServiceImpl) UpdateCert(ctx context.Context, cert *unstructured.Unstructured) error {
	oh, found, err := unstructured.NestedStringSlice(cert.Object, "spec", "dnsNames")
	if err != nil {
		return fmt.Errorf("error parsing dnsNames in Certificate '%s', namespace '%s', %v", c.CertObjectKey(), c.ing.GetNameSpace(), err)
	}

	if !found {
		return fmt.Errorf("dnsNames not found in Certificate '%s', namespace '%s'", c.CertObjectKey(), c.ing.GetNameSpace())
	}

	nh := c.ing.GetHosts()

	hp := func(s1, s2 []string) bool {
		aCopy := slices.Clone(s1)
		bCopy := slices.Clone(s2)

		slices.Sort(aCopy)
		slices.Sort(bCopy)

		return slices.Equal(aCopy, bCopy)
	}

	if !hp(oh, nh) {
		if _, err := c.ing.GetDynamicClientSet().Resource(c.certGVR()).Namespace(c.ing.GetNameSpace()).Update(ctx, c.certUnstructuredData(), metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (c *CertServiceImpl) CheckCert() error {
	if err := c.issuer.CheckIssuer(); err != nil {
		klog.Error(fmt.Sprintf("CheckIssuer error %v", err))
		return err
	}

	cert, err := c.GetCert()
	if err != nil {
		klog.Error(fmt.Sprintf("GetCert error %v", err))
		cert, err = c.CreateCert()
		if err != nil {
			klog.Error(fmt.Sprintf("CreateCert error %v", err))
			return err
		}
	}

	if err := c.UpdateCert(c.ctx, cert); err != nil {
		klog.Error(fmt.Sprintf("UpdateCert error %v", err))
		return err
	}

	return nil
}
