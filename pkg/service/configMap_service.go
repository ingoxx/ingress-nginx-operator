package service

import v1 "k8s.io/api/core/v1"

type K8sResourceConfigMap interface {
	GetConfigMapData(name string) ([]byte, error)
	UpdateConfigMap(name, ns, key string, data []byte) (string, error)
	GetNgxConfigMap(name string) (map[string]string, error)
	GetCmName() string
	GetCm() (*v1.ConfigMap, error)
	ClearCmData(string) error
	DeleteConfigMap() error
}
