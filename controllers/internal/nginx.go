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
	podsIp           []string
}

func NewNginxController(data service.ResourcesMth, config *annotations.IngressAnnotationsConfig) *NginxController {
	return &NginxController{
		allResourcesData: data,
		config:           config,
		mux:              new(sync.Mutex),
		httpReqChan:      make(chan NginxConfig),
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
	// nginx.conf中的stream功能
	if nc.config.EnableStream.EnableStream {
		b, err := json.Marshal(&nc.config.EnableStream.StreamBackendList)
		if err != nil {
			return err
		}

		cm1, err := nc.allResourcesData.UpdateConfigMap(nc.allResourcesData.GetCmName(), nc.allResourcesData.GetNameSpace(), constants.StreamKey, b)
		if err != nil {
			return err
		}

		var tnb []*stream.Backend
		if err := json.Unmarshal([]byte(cm1), &tnb); err != nil {
			return err
		}

		nc.config.EnableStream.StreamBackendList = tnb
	}

	// nginx.conf中的limitreq功能
	if nc.config.EnableReqLimit.EnableRequestLimit {
		b2, err := json.Marshal(&nc.config.EnableReqLimit.Bs.Backends)
		if err != nil {
			return err
		}

		cm2, err := nc.allResourcesData.UpdateConfigMap(nc.allResourcesData.GetCmName(), nc.allResourcesData.GetNameSpace(), constants.LimitReqKey, b2)
		if err != nil {
			return err
		}

		var lb []*limitreq.ZoneRepConfig
		if err := json.Unmarshal([]byte(cm2), &lb); err != nil {
			return err
		}

		nc.config.EnableReqLimit.Bs.Backends = lb
	}

	ips, err := nc.allResourcesData.GetAllEndPoints()
	if err != nil {
		return err
	}

	nc.podsIp = ips

	return nil
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

	for _, ip := range nc.podsIp {
		url := fmt.Sprintf("http://%s:%d%s", ip, constants.HealthPort, constants.NginxConfUpUrl)
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
	}

	//url := fmt.Sprintf("http://%s.%s.svc:%d%s", constants.DeploySvcName, nc.allResourcesData.GetNameSpace(), constants.HealthPort, constants.NginxConfUpUrl)
	//file := NginxConfig{
	//	FileName:  fmt.Sprintf("%s/%s_%s.conf", constants.NginxConfDir, nc.allResourcesData.GetName(), nc.allResourcesData.GetNameSpace()),
	//	Url:       url,
	//	FileBytes: buffer.Bytes(),
	//}
	//
	//if err := nc.updateNginxTls(url); err != nil {
	//	return err
	//}
	//
	//if err := nc.updateNginxConfig(file); err != nil {
	//	return err
	//}

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

	for _, ip := range nc.podsIp {
		url := fmt.Sprintf("http://%s:%d%s", ip, constants.HealthPort, constants.NginxConfUpUrl)
		file := NginxConfig{
			FileName:  constants.NginxMainConf,
			Url:       url,
			FileBytes: buffer.Bytes(),
		}

		if err := nc.updateNginxConfig(file); err != nil {
			return err
		}
	}

	//url := fmt.Sprintf("http://%s.%s.svc:%d%s", constants.DeploySvcName, nc.allResourcesData.GetNameSpace(), constants.HealthPort, constants.NginxConfUpUrl)
	//file := NginxConfig{
	//	FileName:  constants.NginxMainConf,
	//	Url:       url,
	//	FileBytes: buffer.Bytes(),
	//}
	//
	//if err := nc.updateNginxConfig(file); err != nil {
	//	return err
	//}

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
