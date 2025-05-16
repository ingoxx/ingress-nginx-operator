package service

type K8sConfigMapCert interface {
	GetConfigMapData(name string) ([]byte, error)
}
