package service

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type K8sResourcesCert interface {
	CreateCert() (*unstructured.Unstructured, error)
	GetCert() (*unstructured.Unstructured, error)
	DeleteCert() error
	UpdateCert(*unstructured.Unstructured) error
	CheckCert() error
	CertObjectKey() string
	SecretObjectKey() string
	IssuerObjectKey() string
}
