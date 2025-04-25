package service

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type K8sResourcesCert interface {
	CreateCert() error
	GetCert() (*unstructured.Unstructured, error)
	DeleteCert() error
	UpdateCert(context.Context, *unstructured.Unstructured) error
	CheckCert() error
	CertObjectKey() string
	SecretObjectKey() string
	IssuerObjectKey() string
}
