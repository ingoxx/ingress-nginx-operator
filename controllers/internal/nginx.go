package internal

import (
	"bytes"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/klog/v2"
	"os"
	"sync"
	"text/template"
)

type Config struct {
	ServersBuffer bytes.Buffer
	Annotations   *annotations.IngressAnnotationsConfig
	ServerTmpl    string
	NginxConfTmpl string
	ConfName      string
}

type NginxController struct {
	resData service.ResourcesMth
	config  *annotations.IngressAnnotationsConfig
	mux     *sync.Mutex
}

func NewNginxController(data service.ResourcesMth, config *annotations.IngressAnnotationsConfig) *NginxController {
	return &NginxController{
		resData: data,
		config:  config,
		mux:     new(sync.Mutex),
	}
}

func (nc *NginxController) Run() error {
	nc.mux.Lock()
	defer nc.mux.Unlock()
	if err := nc.generateBackendCfg(); err != nil {
		return err
	}

	return nil
}

//func (nc *NginxController) updateBackendCfg() (*annotations.IngressAnnotationsConfig, error) {
//	tls, err := nc.resData.GetTlsFile()
//	if err != nil {
//		return nil, err
//	}
//
//	lbConfig := nc.config.LoadBalance.LbConfig
//	for _, v := range lbConfig {
//		for h := range tls {
//			if v.Host == h {
//				v.Cert = tls[h]
//				break
//			}
//		}
//	}
//
//	return nc.config, nil
//}

func (nc *NginxController) generateBackendCfg() error {
	//cfg, err := nc.updateBackendCfg()
	//if err != nil {
	//	return err
	//}

	c := &Config{
		ServerTmpl:    constants.NginxServerTmpl,
		NginxConfTmpl: constants.NginxTmpl,
		Annotations:   nc.config,
		ConfName:      constants.NginxConfDir,
	}

	// 生成nginx.conf配置
	if err := nc.generateNgxConfTmpl(c); err != nil {
		return err
	}

	// 生成conf.d/下的各个子配置
	if err := nc.generateServerTmpl(c); err != nil {
		return err
	}

	return nil
}

func (nc *NginxController) generateServerTmpl(cfg *Config) error {
	serverTemp, err := nc.renderTemplateData(cfg.ServerTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&cfg.ServersBuffer, cfg); err != nil {
		return err
	}

	fmt.Println("server.conf >>> ", cfg.ServersBuffer.String())

	return nil
}

func (nc *NginxController) generateNgxConfTmpl(cfg *Config) error {
	serverTemp, err := nc.renderTemplateData(cfg.NginxConfTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&cfg.ServersBuffer, cfg); err != nil {
		return err
	}

	fmt.Println("nginx.conf >>> ", cfg.ServersBuffer.String())

	return nil
}

func (nc *NginxController) renderTemplateData(file string) (*template.Template, error) {
	var tmp = new(template.Template)
	b, err := os.ReadFile(file)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("tmpelate file '%s' not found", file))
		return tmp, err
	}

	tmp, err = template.New("server").Parse(string(b))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template '%s'", file))
		return tmp, err
	}

	return tmp, nil
}

func (nc *NginxController) updateNginxConfig() error {
	return nil
}

func (nc *NginxController) generateDefaultBackendTmpl() error {
	return nil
}

func (nc *NginxController) getDefaultBackendCfg() error {

	return nil
}
