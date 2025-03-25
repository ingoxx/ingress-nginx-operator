package service

import (
	"context"
	corev1 "k8s.io/api/core/v1"
)

type K8sResourcesPod interface {
	CreatePod(ctx context.Context, pod *corev1.Pod) error
	GetPods(ctx context.Context, namespace string) (*corev1.PodList, error)
	DeletePod(ctx context.Context, namespace, name string) error
}
