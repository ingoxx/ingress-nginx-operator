package service

type K8sResourceConfigMap interface {
	GetConfigMapData(name string) ([]byte, error)
	UpdateConfigMap(name, ns, key string, data []byte) (string, error)
	GetNgxConfigMap(name string) (map[string]string, error)
}
