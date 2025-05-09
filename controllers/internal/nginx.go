package internal

import (
	"bytes"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/klog/v2"
	"os"
	"text/template"
)

type Config struct {
	ServersBuffer bytes.Buffer
	Annotations   *annotations.IngressAnnotationsConfig
	ServerTmpl    string
	ConfName      string
}

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
	return nc.generateBackendCfg()
}

func (nc *NginxController) updateBackendCfg() (*annotations.IngressAnnotationsConfig, error) {
	tls, err := nc.data.GetTlsFile()
	if err != nil {
		return nil, err
	}

	lbConfig := nc.config.LoadBalance.LbConfig
	for _, v := range lbConfig {
		for h := range tls {
			if v.Host == h {
				v.Cert = tls[h]
				break
			}
		}
	}

	return nc.config, nil
}

func (nc *NginxController) generateBackendCfg() error {
	cfg, err := nc.updateBackendCfg()
	if err != nil {
		return err
	}

	c := &Config{
		ServerTmpl:  constants.NginxServerTmpl,
		Annotations: cfg,
		ConfName:    constants.NginxConfDir,
	}

	if err := nc.generateBackendTmpl(c); err != nil {
		return err
	}

	return nil
}

func (nc *NginxController) generateBackendTmpl(cfg *Config) error {
	b, err := os.ReadFile(cfg.ServerTmpl)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("tmpelate file '%s' not found", cfg.ServerTmpl))
		return err
	}

	serverTemp, err := template.New("server").Parse(string(b))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template '%s'", cfg.ServerTmpl))
		return err
	}

	if err := serverTemp.Execute(&cfg.ServersBuffer, cfg); err != nil {
		return err
	}

	fmt.Println(cfg.ServersBuffer.String())

	return nil
}

func (nc *NginxController) generateDefaultBackendTmpl() error {
	return nil
}

func (nc *NginxController) getDefaultBackendCfg() error {

	return nil
}
