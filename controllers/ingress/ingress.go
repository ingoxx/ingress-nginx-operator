package ingress

import v1 "k8s.io/api/networking/v1"

type IngConfig interface {
	GetIngAnnConfig()
}

type Servers struct {
	Cert                 Tls       `json:"cert"`
	IngAnnotationsConfig IngConfig `json:"ing_annotations_config"`
	Bks                  *Backends `json:"bks"`
	Host                 string    `json:"host"`
}

type Tls struct {
	TlsKey string `json:"tls_key"`
	TlsCrt string `json:"tls_crt"`
}

type Backends struct {
	Services             []*v1.ServiceBackendPort `json:"services"`
	IngAnnotationsConfig IngConfig                `json:"ing_annotations_config"`
	Host                 string                   `json:"host"`
	Upstream             string                   `json:"upstream"`
	Path                 string                   `json:"path"`
	PathType             string                   `json:"path_type"`
}
