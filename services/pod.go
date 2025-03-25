package services

import (
	"context"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodServiceImpl 实现 PodService 接口
type PodServiceImpl struct {
	K8sClient interfaces.K8sClientSet
}

// NewPodServiceImpl 创建 Service 实例
func NewPodServiceImpl(client interfaces.K8sClientSet) *PodServiceImpl {
	return &PodServiceImpl{K8sClient: client}
}

// GetPods 获取 Pod 列表
func (s *PodServiceImpl) GetPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	pods, err := s.K8sClient.GetClientSet().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing pods: %w", err)
	}
	return pods, nil
}

func (s *PodServiceImpl) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	_, err := s.K8sClient.GetClientSet().CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *PodServiceImpl) DeletePod(ctx context.Context, namespace, name string) error {
	if err := s.K8sClient.GetClientSet().CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
