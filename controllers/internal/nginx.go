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
	Annotations   *annotations.IngressAnnotationsConfig
	ServerTmpl    string
	NginxConfTmpl string
	ConfDir       string
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

func (nc *NginxController) generateBackendCfg() error {
	c := &Config{
		ServerTmpl:    constants.NginxServerTmpl,
		NginxConfTmpl: constants.NginxTmpl,
		Annotations:   nc.config,
		ConfDir:       constants.NginxConfDir,
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
	var buffer bytes.Buffer

	serverTemp, err := nc.renderTemplateData(cfg.ServerTmpl)
	if err != nil {
		return err
	}

	fmt.Println("ServerTmpl >>> ", cfg.ServerTmpl)

	if err := serverTemp.Execute(&buffer, cfg); err != nil {
		return err
	}

	fmt.Println("server.conf >>> ", buffer.String())

	return nil
}

func (nc *NginxController) generateNgxConfTmpl(cfg *Config) error {
	var buffer bytes.Buffer

	serverTemp, err := nc.renderTemplateData(cfg.NginxConfTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&buffer, cfg); err != nil {
		return err
	}

	fmt.Println("nginx.conf >>> ", buffer.String())

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
