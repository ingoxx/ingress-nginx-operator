package ingress

import v1 "k8s.io/api/networking/v1"

type IngConfig interface {
	GetIngAnnConfig()
}

// Tls nginx证书配置
type Tls struct {
	TlsKey string `json:"tls_key"`
	TlsCrt string `json:"tls_crt"`
}

type IngBackends struct {
	Services      *v1.ServiceBackendPort `json:"services"`
	Path          string                 `json:"path"`
	PathType      string                 `json:"path_type"`
	SvcName       string                 `json:"svc_name"`
	IsPathIsRegex bool                   `json:"is_path_is_regex"`
}

// Backends nginx配置结构，会生成对应的nginx配置
type Backends struct {
	ServiceBackend       []*IngBackends `json:"service_backend"`
	IngAnnotationsConfig IngConfig      `json:"ing_annotations_config"`
	Cert                 Tls            `json:"cert"`
	Host                 string         `json:"host"`
	Upstream             string         `json:"upstream"`
}

// LbConfigList annotations中的序列化结构
type LbConfigList struct {
	Backends []*LbConfig // 负载均衡后端列表
}

type LbConfig struct {
	Name   string // svc的name
	Config string // 负载均衡的参数配置，如：max_fails=3 fail_timeout=30s weight=20
}
