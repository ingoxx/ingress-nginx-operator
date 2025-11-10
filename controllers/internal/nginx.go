package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitconn"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"io"
	v1 "k8s.io/api/networking/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
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
	podsIp           []string
}

func NewNginxController(data service.ResourcesMth, config *annotations.IngressAnnotationsConfig) *NginxController {
	return &NginxController{
		allResourcesData: data,
		config:           config,
		mux:              new(sync.Mutex),
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

// checkPublicCfg 先检查nginx.conf中的配置是否重复配置
func (nc *NginxController) checkPublicCfg() error {
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

		if _, err := nc.getStreamData(cm1); err != nil {
			return err
		}

	} else {
		if err := nc.allResourcesData.ClearCmData(constants.StreamKey); err != nil {
			if !kerr.IsNotFound(err) {
				return err
			}
		}
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

		if _, err := nc.getLimitReqData(cm2); err != nil {
			return err
		}

	} else {
		if err := nc.allResourcesData.ClearCmData(constants.LimitReqKey); err != nil {
			if !kerr.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (nc *NginxController) getStreamData(data string) ([]*stream.Backend, error) {
	var tnb []*stream.Backend
	if err := json.Unmarshal([]byte(data), &tnb); err != nil {
		return tnb, err
	}

	return tnb, nil
}

func (nc *NginxController) getLimitReqData(data string) ([]*limitreq.ZoneRepConfig, error) {
	var lb []*limitreq.ZoneRepConfig
	if err := json.Unmarshal([]byte(data), &lb); err != nil {
		return lb, err
	}

	return lb, nil
}

func (nc *NginxController) getLimitConnData(data string) ([]*limitconn.ZoneConnConfig, error) {
	var lb []*limitconn.ZoneConnConfig
	if err := json.Unmarshal([]byte(data), &lb); err != nil {
		return lb, err
	}

	return lb, nil
}

func (nc *NginxController) generateBackendCfg() error {
	var err error

	if err := nc.checkPublicCfg(); err != nil {
		return err
	}

	cm, err := nc.allResourcesData.GetNgxConfigMap(nc.allResourcesData.GetNameSpace())
	if err != nil {
		return err
	}

	s, ok := cm[constants.StreamKey]
	if s != "" && ok {
		nc.config.EnableStream.EnableStream = true
		data, err := nc.getStreamData(s)
		if err != nil {
			return err
		}
		nc.config.EnableStream.StreamBackendList = data
	}

	s1, ok := cm[constants.LimitReqKey]
	if s != "" && ok {
		nc.config.EnableReqLimit.EnableRequestLimit = true
		data, err := nc.getLimitReqData(s1)
		if err != nil {
			return err
		}
		nc.config.EnableReqLimit.Bs.Backends = data
	}

	s2, ok := cm[constants.LimitConnKey]
	if s != "" && ok {
		nc.config.EnableConnLimit.EnableConnLimit = true
		data, err := nc.getLimitConnData(s2)
		if err != nil {
			return err
		}
		nc.config.EnableConnLimit.Bs.Backends = data
	}

	if err := nc.allResourcesData.CheckSvc(); err != nil {
		return err
	}

	if err := nc.allResourcesData.CheckDeploy(); err != nil {
		return err
	}

	nc.podsIp, err = nc.allResourcesData.GetAllEndPoints()
	if err != nil {
		return err
	}

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
