package internal

import (
	"bytes"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"text/template"
)

type ssl ingress.Tls

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

func (nc *NginxController) generateBackendCfg() error {
	c := &Config{
		ServerTmpl:  constants.NginxServerTmpl,
		Annotations: nc.config,
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

	serverTemp, err := template.New("serverMain").Parse(string(b))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template_nginx: %s", cfg.ServerTmpl))
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

func (nc *NginxController) generateTlsFile() (map[string]ssl, error) {
	if len(nc.data.GetTls()) > 0 {
		cs, err := nc.caSigned()
		if err != nil {
			return nil, err
		}

		return cs, nil
	}

	ss, err := nc.selfSigned()
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// selfSigned use cert-manager controller
func (nc *NginxController) selfSigned() (map[string]ssl, error) {
	var ss ssl
	var ht = make(map[string]ssl)

	key := types.NamespacedName{Name: nc.data.SecretObjectKey(), Namespace: nc.data.GetNameSpace()}
	data, err := nc.data.GetTlsData(key)
	if err != nil {
		return nil, nil
	}

	for k, v := range data {
		file := filepath.Join(constants.NginxSSLDir, nc.data.SecretObjectKey()+"-"+k)
		if err := os.WriteFile(file, v, 0644); err != nil {
			return nil, err
		}

		if k == constants.NginxTlsCrt {
			ss.TlsCrt = file
		} else if k == constants.NginxTlsKey {
			ss.TlsKey = file
		}
	}

	for _, v := range nc.data.GetHosts() {
		ht[v] = ss
	}

	return nil, nil
}

// caSigned ca signed
func (nc *NginxController) caSigned() (map[string]ssl, error) {
	if !nc.data.CheckTlsHosts() {
		return nil, cerr.NewNotFoundTlsHostError(nc.data.GetName(), nc.data.GetNameSpace())
	}

	var ss ssl
	var ht = make(map[string]ssl)

	for _, v := range nc.data.GetTls() {
		key := types.NamespacedName{Name: v.SecretName, Namespace: nc.data.GetNameSpace()}
		_, err := nc.data.GetSecret(key)
		if err != nil {
			return nil, err
		}

		data, err := nc.data.GetTlsData(key)
		if err != nil {
			return nil, err
		}

		for k, v := range data {
			file := filepath.Join(constants.NginxSSLDir, nc.data.SecretObjectKey())
			if err := os.WriteFile(file, v, 0644); err != nil {
				return nil, err
			}

			if k == constants.NginxTlsCrt {
				ss.TlsCrt = file
			} else if k == constants.NginxTlsKey {
				ss.TlsKey = file
			}
		}

		for _, h := range v.Hosts {
			ht[h] = ss
		}

	}

	return nil, nil
}
