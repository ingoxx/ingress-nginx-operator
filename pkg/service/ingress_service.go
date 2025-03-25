package service

import (
	"golang.org/x/net/context"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
)

type K8sResourcesIngress interface {
	GetIngress(ctx context.Context, namespace, name string) (v1.Ingress, error)
	GetIngressHosts(ctx context.Context, namespace, name string) []string
	GetIngressBackends(ctx context.Context, namespace, name string) ([]v12.Service, error)
}
