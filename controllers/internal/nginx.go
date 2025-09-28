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
	"path/filepath"
	"sync"
	"text/template"
)

type RespData struct {
	Msg    string `json:"msg"`
	Code   int    `json:"code"`
	Status int    `json:"status"`
}

type Config struct {
	Annotations      *annotations.IngressAnnotationsConfig
	DefaultBackend   *v1.ServiceBackendPort
	DefaultBackendAd string
	ServerTmpl       string
	NginxConfTmpl    string
	DefaultConfTmpl  string
	ConfDir          string
	DefaultPort      int32
}

type NginxConfig struct {
	FileName  string `json:"file_name"`
	Url       string `json:"-"`
	FileBytes []byte `json:"file_bytes"`
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

func (nc *NginxController) appendUniqueData() {}

func (nc *NginxController) getPublicNgxConfig() (map[string]string, error) {
	name := fmt.Sprintf("%s-%s-ngx-cm", nc.allResourcesData.GetName(), nc.allResourcesData.GetNameSpace())
	cm, err := nc.allResourcesData.GetNgxConfigMap(name)
	if err != nil {
		return cm, err
	}

	return cm, nil
}

// publicConfig 保存stream跟limit功能是公共配置
func (nc *NginxController) setPublicNgxConfig() (map[string]string, error) {
	var pd map[string]string
	name := fmt.Sprintf("%s-%s-ngx-cm", nc.allResourcesData.GetName(), nc.allResourcesData.GetNameSpace())
	if nc.config.EnableStream.EnableStream {

		b, err := json.Marshal(&nc.config.EnableStream)
		if err != nil {
			return pd, err
		}

		data, err := nc.allResourcesData.UpdateConfigMap(name, constants.StreamKey, b)
		if err != nil {
			return data, err
		}
	}

	if nc.config.EnableReqLimit.EnableRequestLimit {
		b, err := json.Marshal(&nc.config.EnableReqLimit)
		if err != nil {
			return nil, err
		}

		data, err := nc.allResourcesData.UpdateConfigMap(name, constants.StreamKey, b)
		if err != nil {
			return data, err
		}
	}

	return nil, nil

}

func (nc *NginxController) generateBackendCfg() error {
	c := &Config{
		ServerTmpl:    constants.NginxServerTmpl,
		NginxConfTmpl: constants.NginxTmpl,
		Annotations:   nc.config,
		ConfDir:       constants.NginxConfDir,
	}

	if err := nc.generateNgxConfTmpl(c); err != nil {
		return err
	}

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

	url := fmt.Sprintf("http://%s.%s.svc:%d%s", constants.DeploySvcName, nc.allResourcesData.GetNameSpace(), constants.HealthPort, constants.NginxConfUpUrl)
	file := NginxConfig{
		FileName:  fmt.Sprintf("%s/%s_%s.conf", constants.NginxConfDir, nc.allResourcesData.GetName(), nc.allResourcesData.GetNameSpace()),
		Url:       url,
		FileBytes: buffer.Bytes(),
	}

	if err := nc.updateNginxTls(url); err != nil {
		return err
	}

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
		cfg.DefaultPort = int32(constants.DefaultPort)
	}

	serverTemp, err := nc.renderTemplateData(cfg.NginxConfTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&buffer, cfg); err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s.%s.svc:%d%s", constants.DeploySvcName, nc.allResourcesData.GetNameSpace(), constants.HealthPort, constants.NginxConfUpUrl)
	file := NginxConfig{
		FileName:  constants.NginxMainConf,
		Url:       url,
		FileBytes: buffer.Bytes(),
	}

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
	var respData RespData
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.Url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", constants.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err := json.Unmarshal(body, &respData); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(respData.Msg)
	}

	if respData.Code != constants.HttpStatusOk {
		return errors.New(respData.Msg)
	}

	return nil
}

func (nc *NginxController) updateNginxTls(url string) error {
	tls := []string{
		filepath.Join(constants.NginxSSLDir, fmt.Sprintf("%s-%s", nc.allResourcesData.SecretObjectKey(), constants.NginxTlsCrt)),
		filepath.Join(constants.NginxSSLDir, fmt.Sprintf("%s-%s", nc.allResourcesData.SecretObjectKey(), constants.NginxTlsKey)),
		filepath.Join(constants.NginxSSLDir, fmt.Sprintf("%s-%s", nc.allResourcesData.SecretObjectKey(), constants.NginxTlsCa)),
	}
	for _, v := range tls {
		b, err := os.ReadFile(v)
		if err != nil {
			return err
		}
		file := NginxConfig{
			FileName:  v,
			Url:       url,
			FileBytes: b,
		}
		if err := nc.updateNginxConfig(file); err != nil {
			return err
		}
	}

	return nil
}
