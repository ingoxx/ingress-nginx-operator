package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
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

	if err := nc.check(); err != nil {
		return err
	}

	if err := nc.generateBackendCfg(); err != nil {
		return err
	}

	return nil
}

// check 先检查nginx.conf中的配置是否重复配置
func (nc *NginxController) check() error {
	name := fmt.Sprintf("%s-%s-ngx-map", nc.allResourcesData.GetName(), nc.allResourcesData.GetNameSpace())
	cm, err := nc.allResourcesData.GetNgxConfigMap(nc.allResourcesData.GetNameSpace())
	if err != nil {
		return err
	}

	// nginx.conf中的stream功能
	var nb []*stream.Backend
	if nc.config.EnableStream.EnableStream {
		if len(cm) == 0 {
			nb = nc.config.EnableStream.StreamBackendList
		} else {
			es := cm[constants.StreamKey]
			if es == "" {
				nb = nc.config.EnableStream.StreamBackendList
			} else {
				if err := json.Unmarshal([]byte(es), &nb); err != nil {
					return err
				}

				nc.config.EnableStream.StreamBackendList = nc.isUniquePort(nb, nc.config.EnableStream.StreamBackendList)
				nb = nc.config.EnableStream.StreamBackendList
			}
		}

		b, err := json.Marshal(&nb)
		if err != nil {
			return err
		}

		_, err = nc.allResourcesData.UpdateConfigMap(name, constants.StreamKey, b)
		if err != nil {
			return err
		}
	}

	// nginx.conf中的limitreq功能
	var lb []*limitreq.ZoneRepConfig
	if nc.config.EnableReqLimit.EnableRequestLimit {
		if len(cm) == 0 {
			lb = nc.config.EnableReqLimit.Bs.Backends
		} else {
			es := cm[constants.LimitReqKey]
			if es == "" {
				lb = nc.config.EnableReqLimit.Bs.Backends
			} else {
				if err := json.Unmarshal([]byte(es), &lb); err != nil {
					return err
				}

				nc.config.EnableReqLimit.Bs.Backends = nc.isUniqueZone(lb, nc.config.EnableReqLimit.Bs.Backends)
				lb = nc.config.EnableReqLimit.Bs.Backends
			}
		}

		b, err := json.Marshal(&lb)
		if err != nil {
			return err
		}
		_, err = nc.allResourcesData.UpdateConfigMap(name, constants.LimitReqKey, b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (nc *NginxController) isUniquePort(cmBackend []*stream.Backend, anBackend []*stream.Backend) []*stream.Backend {
	keySet := make(map[int32]struct{})

	for _, b := range cmBackend {
		keySet[b.Port] = struct{}{}
	}

	for _, b := range anBackend {
		if _, exists := keySet[b.Port]; !exists {
			cmBackend = append(cmBackend, b)
		}
	}

	return cmBackend
}

func (nc *NginxController) isUniqueZone(cmBackend []*limitreq.ZoneRepConfig, anBackend []*limitreq.ZoneRepConfig) []*limitreq.ZoneRepConfig {
	keySet := make(map[string]struct{})

	for _, b1 := range cmBackend {
		for _, b2 := range b1.LimitZone {
			keySet[b2.ZoneName] = struct{}{}
		}

	}

	for _, b1 := range anBackend {
		for _, b2 := range b1.LimitZone {
			if _, exists := keySet[b2.ZoneName]; !exists {
				cmBackend = append(cmBackend, b1)
			}
		}

	}

	return cmBackend
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
