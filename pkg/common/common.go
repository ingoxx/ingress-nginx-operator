package common

import "github.com/ingoxx/ingress-nginx-operator/pkg/service"

// Generic 聚合通用接口
type Generic interface {
	K8sClientSet
	OperatorClientSet
	service.K8sResourcesIngress
}
