package controllers

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/business/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
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

	return nil, nil
}
