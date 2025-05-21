package service

type K8sResourceConfigMap interface {
	GetConfigMapData(name string) ([]byte, error)
}
