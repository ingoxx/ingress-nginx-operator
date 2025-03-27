package services

import (
	"context"
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CertServiceImpl 实现 CertService 接口
type CertServiceImpl struct {
	K8sClient interfaces.K8sClientSet
}

// NewCertServiceImpl 创建 Service 实例
func NewCertServiceImpl(client interfaces.K8sClientSet) *PodServiceImpl {
	return &PodServiceImpl{K8sClient: client}
}

func (c *CertServiceImpl) certGVK() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
}

func (c *CertServiceImpl) certUnstructured() *unstructured.Unstructured {
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
		},
	}

	return certificate
}

func (c *CertServiceImpl) certUnstructuredData(namespace, name string, hosts []string) *unstructured.Unstructured {
	certUnstructured := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"dnsNames": hosts,
				"issuerRef": map[string]interface{}{
					"kind": "Issuer",
					"name": name + "-issuer",
				},
				"secretName": name + "-secret",
			},
		},
	}

	return certUnstructured
}

func (c *CertServiceImpl) CreateCert(ctx context.Context, namespace, name string) error {

	return nil
}

func (c *CertServiceImpl) GetCert(ctx context.Context, namespace, name string) error {

	return nil
}

func (c *CertServiceImpl) DeleteCert(ctx context.Context, namespace, name string) error {
	return nil
}

func (c *CertServiceImpl) UpdateCert(ctx context.Context, namespace, name string) error {
	return nil
}

func (c *CertServiceImpl) ResourceData() error {
	return nil
}
