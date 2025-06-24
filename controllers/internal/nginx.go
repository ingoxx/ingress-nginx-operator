package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"io"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"sync"
	"text/template"
)

type Config struct {
	Annotations      *annotations.IngressAnnotationsConfig
	DefaultBackend   *v1.ServiceBackendPort
	DefaultBackendAd string
	ServerTmpl       string
	NginxConfTmpl    string
	DefaultConfTmpl  string
	ConfDir          string
}

type NginxConfig struct {
	FileName string
	FileByte []byte
}

type NginxController struct {
	allResourcesData service.ResourcesMth
	config           *annotations.IngressAnnotationsConfig
	mux              *sync.Mutex
	httpReqChan      chan NginxConfig
	limitReqChan     chan struct{}
	errorChan        chan error
}

func NewNginxController(data service.ResourcesMth, config *annotations.IngressAnnotationsConfig) *NginxController {
	return &NginxController{
		allResourcesData: data,
		config:           config,
		mux:              new(sync.Mutex),
		httpReqChan:      make(chan NginxConfig),
		limitReqChan:     make(chan struct{}, 10),
		errorChan:        make(chan error),
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

// generateServerTmpl 生成conf.d/下的各个子配置
func (nc *NginxController) generateServerTmpl(cfg *Config) error {
	var buffer bytes.Buffer

	serverTemp, err := nc.renderTemplateData(cfg.ServerTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&buffer, cfg); err != nil {
		return err
	}

	file := NginxConfig{
		FileName: fmt.Sprintf("%s/%s_%s.conf", constants.NginxConfDir, nc.allResourcesData.GetName(), nc.allResourcesData.GetNameSpace()),
		FileByte: buffer.Bytes(),
	}

	fmt.Println("server.conf >>> ", buffer.String(), file)

	if err := nc.updateNginxConfig(file); err != nil {
		return err
	}

	return nil
}

// generateNgxConfTmpl 生成nginx.conf配置
func (nc *NginxController) generateNgxConfTmpl(cfg *Config) error {
	var buffer bytes.Buffer

	backend, err := nc.allResourcesData.GetDefaultBackend()
	if err != nil {
		return err
	}

	if backend.Name != "" && backend.Number > 0 {
		cfg.DefaultBackend = backend
		cfg.DefaultBackendAd = nc.allResourcesData.GetBackendName(backend)
	}

	serverTemp, err := nc.renderTemplateData(cfg.NginxConfTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&buffer, cfg); err != nil {
		return err
	}

	file := NginxConfig{
		FileName: constants.NginxMainConf,
		FileByte: buffer.Bytes(),
	}

	fmt.Println("nginx.conf >>> ", buffer.String())

	if err := nc.updateNginxConfig(file); err != nil {
		return err
	}

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

func (nc *NginxController) updateNginxConfig(config NginxConfig) error {
	var respData map[string]interface{}
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	data := bytes.NewBuffer(b)

	url := fmt.Sprintf("%s/%s", constants.NginxConfUpUrl, constants.NginxConfUpUrl)
	resp, err := http.Post(url, "application/json", data)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(respData["msg"].(string))
	}

	return nil
}
