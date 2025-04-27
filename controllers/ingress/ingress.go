package ingress

type IngConfig interface {
	GetIngAnnConfig()
}

type Servers struct {
	Host                 string
	Cert                 Tls
	IngAnnotationsConfig IngConfig
	Bks                  []*Backends
}

type Backends struct {
	Path                 string
	IngAnnotationsConfig IngConfig
}

type Tls struct {
	TlsKey    string
	TlsCrt    string
	TlsNoPass bool
}
