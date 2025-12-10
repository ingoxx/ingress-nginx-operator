package service

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type K8sResourcesIssuer interface {
	CreateIssuer(ctx context.Context, namespace, name string) error
	GetIssuer(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error)
	DeleteIssuer() error
	UpdateIssuer(ctx context.Context, issuer *unstructured.Unstructured) error
	CheckIssuer() error
}
