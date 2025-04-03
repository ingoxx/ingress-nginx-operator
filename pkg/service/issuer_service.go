package service

import (
	"context"
)

type K8sResourcesIssuer interface {
	CreateIssuer(ctx context.Context, namespace, name string) error
	GetIssuer(ctx context.Context, namespace, name string) error
	DeleteIssuer(ctx context.Context, namespace, name string) error
	UpdateIssuer(ctx context.Context, namespace, name string) error
	CheckIssuer() error
}
