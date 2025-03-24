package client

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type KubernetesApiClient struct {
	dynamicClientSet dynamic.Interface
	client           *kubernetes.Clientset
	ctx              context.Context
}

// NewK8sClient 创建一个新的 Kubernetes 客户端
func NewK8sClient() (*KubernetesApiClient, error) {
	configSet, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("fail to create kubernetes cluster Config: %v", err)
		}
		configSet = inClusterConfig
	}

	// 创建 ClientSet
	clientSet, err := kubernetes.NewForConfig(configSet)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// 创建 dynamicClientSet
	forConfig, err := dynamic.NewForConfig(configSet)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes dynamicClientSet: %w", err)
	}

	return &KubernetesApiClient{client: clientSet, dynamicClientSet: forConfig}, nil
}

func (c *KubernetesApiClient) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	_, err := c.client.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *KubernetesApiClient) GetPod(ctx context.Context, namespace, name string) (*corev1.PodList, error) {
	list, err := c.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return list, err
	}

	return nil, nil
}

func (c *KubernetesApiClient) DeletePod(ctx context.Context, namespace, name string) error {
	if err := c.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
