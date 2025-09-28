package service

type K8sResourceConfigMap interface {
	GetConfigMapData(name string) ([]byte, error)
	UpdateConfigMap(name, key string, data []byte) (map[string]string, error)
	GetNgxConfigMap(name string) (map[string]string, error)
}
