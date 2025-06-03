package ssl

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"os"
	"path/filepath"
	"strconv"
)

const (
	sslStaplingVerifyAnnotations    = "ssl-stapling-verify"
	sslStaplingAnnotations          = "ssl-stapling"
	sslStaplingConfigMapAnnotations = "ssl-trusted-config-map"
)

type sslIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	SSllStaplingVerify bool   `json:"ssl-stapling-verify"`
	SSlStapling        bool   `json:"ssl-stapling-stapling"`
	SSLTrustedCMName   string `json:"ssl-trusted-cm-name"`
	SSLTrustCertFile   string `json:"ssl-trust-cert-file"`
}

var sslAnnotations = parser.AnnotationsContents{
	sslStaplingVerifyAnnotations: {
		Doc: "optional, true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(sslStaplingVerifyAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	sslStaplingAnnotations: {
		Doc: "optional, true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(sslStaplingVerifyAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	sslStaplingConfigMapAnnotations: {
		Doc: "optional, ConfigMap name.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			return nil
		},
	},
}

func NewSSL(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &sslIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (s *sslIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}

	config.SSllStaplingVerify, err = parser.GetBoolAnnotations(sslStaplingVerifyAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}
	config.SSlStapling, err = parser.GetBoolAnnotations(sslStaplingAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SSLTrustedCMName, err = parser.GetStringAnnotation(sslStaplingConfigMapAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if verr := s.validate(config); verr != nil {
		return config, verr
	}

	return config, err
}

func (s *sslIng) validate(config *Config) error {
	if config.SSllStaplingVerify {
		data, err := s.resources.GetConfigMapData(config.SSLTrustedCMName)
		if err != nil {
			return err
		}

		file := filepath.Join(constants.NginxSSLDir, s.resources.SecretObjectKey()+"-"+constants.NginxFullChain)
		if err := os.WriteFile(file, data, 0644); err != nil {
			return err
		}

		config.SSLTrustCertFile = file

	}
	return nil
}

func (s *sslIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, sslAnnotations, s.ingress)
}
