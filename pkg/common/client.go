package common

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sClientSet 通用接口
type K8sClientSet interface {
	GetClientSet() *kubernetes.Clientset
	GetDynamicClientSet() dynamic.Interface
}

// OperatorClientSet 通用接口
type OperatorClientSet interface {
	GetClient() client.Client
}
