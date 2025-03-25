package service

import (
	"context"
)

type K8sResourcesCert interface {
	CreateCert(ctx context.Context, namespace, name string) error
	GetCert(ctx context.Context, namespace, name string) error
	DeleteCert(ctx context.Context, namespace, name string) error
	UpdateCert(ctx context.Context, namespace, name string) error
}
