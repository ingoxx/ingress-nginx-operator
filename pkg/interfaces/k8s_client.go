package interfaces

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// K8sClientSet 通用接口
type K8sClientSet interface {
	GetClientSet() *kubernetes.Clientset
	GetDynamicClientSet() dynamic.Interface
}
