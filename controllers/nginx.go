package controllers

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/apimachinery/pkg/types"
)

type NginxController struct {
	data   service.NginxTemplateData
	config *annotations.IngressAnnotationsConfig
}

func NewNginxController(data service.NginxTemplateData, config *annotations.IngressAnnotationsConfig) *NginxController {
	return &NginxController{
		data:   data,
		config: config,
	}
}

func (nc *NginxController) Run() error {
	return nil
}

func (nc *NginxController) generateCfg() error {
	return nil
}

func (nc *NginxController) generateBackendTmpl() error {
	return nil
}

func (nc *NginxController) generateDefaultBackendTmpl() error {
	return nil
}

func (nc *NginxController) getDefaultBackendCfg() error {
	rs := nc.data.GetRules()
	var servers = make([]*ingress.Servers, len(rs))
	is := &ingress.Servers{
		IngAnnotationsConfig: nc.config,
	}

	fmt.Println(servers, is)
	return nil
}

func (nc *NginxController) generateTlsFile() (map[string]ingress.Tls, error) {
	if len(nc.data.GetTls()) > 0 {
		cs, err := nc.caSigned()
		if err != nil {
			return cs, err
		}

		return nil, nil
	}

	ss, err := nc.selfSigned()
	if err != nil {
		return ss, err
	}

	return nil, nil
}

func (nc *NginxController) selfSigned() (map[string]ingress.Tls, error) {
	key := types.NamespacedName{Name: nc.data.SecretObjectKey(), Namespace: nc.data.GetNameSpace()}
	_, err := nc.data.GetSecret(key)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (nc *NginxController) caSigned() (map[string]ingress.Tls, error) {
	return nil, nil
}
