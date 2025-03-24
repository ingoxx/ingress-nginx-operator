package client

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sClientPod interface {
	CreatePod(ctx context.Context, pod *corev1.Pod) error
	GetPod(ctx context.Context, namespace, name string) (*corev1.PodList, error)
	DeletePod(ctx context.Context, namespace, name string) error
}

type K8sClientCertIssuer interface {
	CreateCert(ctx context.Context, pod *corev1.Pod) error
	GetCert(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	DeleteCert(ctx context.Context, namespace, name string) error
	CreateIssuer(ctx context.Context, pod *corev1.Pod) error
	GetIssuer(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	DeleteIssuer(ctx context.Context, namespace, name string) error
}

type K8sClientDeployment interface {
	CreateDeployment(ctx context.Context, pod *corev1.Pod) error
	GetDeployment(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	DeleteDeployment(ctx context.Context, namespace, name string) error
}
