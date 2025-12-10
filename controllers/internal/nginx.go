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
	"golang.org/x/net/context"
	"io"
	v1 "k8s.io/api/networking/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"
)

const (
	queueSize      = 90                     // 队列缓冲大小（可根据内存 / QPS 调整）
	workerCount    = 30                     // worker 数量（并发处理数）
	enqueueTimeout = 100 * time.Millisecond // 当队列已满时，尝试放入队列的超时时间（实现 backpressure）
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
	IsDel            bool
}

func NewNginxController() *NginxController {
	return &NginxController{
		mux: new(sync.Mutex),
	}
}

func (nc *NginxController) Run(data service.ResourcesMth, config *annotations.IngressAnnotationsConfig) error {
	nc.mux.Lock()
	defer nc.mux.Unlock()

	nc.allResourcesData = data
	nc.config = config

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

	// nginx.conf中的limitconn功能
	if nc.config.EnableConnLimit.EnableConnLimit {
		b3, err := json.Marshal(&nc.config.EnableConnLimit.Bs.Backends)
		if err != nil {
			return err
		}

		cm3, err := nc.allResourcesData.UpdateConfigMap(nc.allResourcesData.GetCmName(), nc.allResourcesData.GetNameSpace(), constants.LimitConnKey, b3)
		if err != nil {
			return err
		}

		if _, err := nc.getLimitConnData(cm3); err != nil {
			return err
		}

	} else {
		if err := nc.allResourcesData.ClearCmData(constants.LimitConnKey); err != nil {
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
	if s2 != "" && ok {
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

	if err := nc.multiRun(c); err != nil {
		return err
	}

	return nil
}

// generateServerTmpl 生成conf.d/下的各个子配置
func (nc *NginxController) generateServerTmpl(cfg *Config, ip string) error {
	var buffer bytes.Buffer
	var nUrl string

	serverTemp, err := nc.renderTemplateData(cfg.ServerTmpl)
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&buffer, cfg); err != nil {
		return err
	}

	if nc.IsDel {
		nUrl = constants.NginxConfDelUrl
	} else {
		nUrl = constants.NginxConfUpUrl
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, constants.HealthPort, nUrl)
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
func (nc *NginxController) generateNgxConfTmpl(cfg *Config, ip string) error {
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

	url := fmt.Sprintf("http://%s:%d%s", ip, constants.HealthPort, constants.NginxConfUpUrl)
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

// http
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

	client := &http.Client{Timeout: time.Second * time.Duration(3)}
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

func (nc *NginxController) worker(ctx context.Context, task chan string, wg *sync.WaitGroup, cfg *Config, errs chan error) {
	defer wg.Done()

	for {
		select {
		case ip := <-task:
			if err := nc.generateNgxConfTmpl(cfg, ip); err != nil {
				errs <- err
				return
			}

			if err := nc.generateServerTmpl(cfg, ip); err != nil {
				errs <- err
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (nc *NginxController) multiRun(cfg *Config) error {
	var wg *sync.WaitGroup
	var tasks = make(chan string, len(nc.podsIp))
	var errs = make(chan error, len(nc.podsIp))
	var te error
	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(3))
	defer cancel()

	wg.Add(len(nc.podsIp))
	for i := 0; i < len(nc.podsIp); i++ {
		go nc.worker(ctx, tasks, wg, cfg, errs)
	}

	for _, v := range nc.podsIp {
		select {
		case tasks <- v:
		case <-time.After(enqueueTimeout):
		}
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		te = errors.Join(e)
	}

	if te != nil {
		return te
	}

	return nil
}
