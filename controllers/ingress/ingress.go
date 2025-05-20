package ingress

import v1 "k8s.io/api/networking/v1"

type IngConfig interface {
	GetIngAnnConfig()
}

type Tls struct {
	TlsKey string `json:"tls_key"`
	TlsCrt string `json:"tls_crt"`
}

type Backends struct {
	Services             []*v1.ServiceBackendPort `json:"services"`
	IngAnnotationsConfig IngConfig                `json:"ing_annotations_config"`
	Cert                 Tls                      `json:"cert"`
	Host                 string                   `json:"host"`
	Upstream             string                   `json:"upstream"`
	Path                 string                   `json:"path"`
	PathType             string                   `json:"path_type"`
}

type StreamBackendList struct {
	Backends []*StreamBackend
}

type StreamBackend struct {
	Name      string
	Namespace string
}
