package client

import (
	"context"
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sApiClient struct {
	dynamicClientSet dynamic.Interface
	ctx              context.Context
	client           *kubernetes.Clientset
}

// NewK8sApiClient 创建一个新的 Kubernetes 客户端
func NewK8sApiClient() (*K8sApiClient, error) {
	configSet, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("fail to create kubernetes cluster Config: %v", err)

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
		return nil, fmt.Errorf("failed to create Kubernetes dynamicClient: %w", err)
	}

	return &K8sApiClient{client: clientSet, dynamicClientSet: forConfig}, nil
}

func (k *K8sApiClient) GetClientSet() *kubernetes.Clientset {
	return k.client
}

func (k *K8sApiClient) GetDynamicClientSet() dynamic.Interface {
	return k.dynamicClientSet
}
