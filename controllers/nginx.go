package controllers

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
)

type NginxController struct {
	data   service.NginxTmplData
	config *annotations.IngressAnnotationsConfig
}

func NewNginxController(data service.NginxTmplData, config *annotations.IngressAnnotationsConfig) *NginxController {
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
	return nil
}

func (nc *NginxController) generateTlsFile() {}
