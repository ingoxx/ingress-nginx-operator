package ssl

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
)

var (
	sslVerifyParams = []string{"off", "on"}
)

const (
	sslRedirectAnnotations          = "ssl-redirect"
	sslStaplingVerifyAnnotations    = "ssl-stapling-verify"
	sslStaplingAnnotations          = "ssl-stapling"
	sslStaplingConfigMapAnnotations = "ssl-trusted-config-map"
	sslVerifyAnnotations            = "ssl-verify"
	sslServerNameAnnotations        = "ssl-server-name"
	sslNameAnnotations              = "ssl-name"
)

type sslIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	SslRedirect        bool   `json:"ssl-redirect"`
	SSllStaplingVerify bool   `json:"ssl-stapling-verify"`
	SSlStapling        bool   `json:"ssl-stapling-stapling"`
	SSLTrustedCMName   string `json:"ssl-trusted-cm-name"`
	SSLTrustCertFile   string `json:"ssl-trust-cert-file"`
	SslVerify          string `json:"ssl-verify"`
	SslServerName      string `json:"ssl-server-name"`
	SslName            string `json:"ssl-name"`
}

var sslAnnotations = parser.AnnotationsContents{
	sslVerifyAnnotations: {
		Doc: "optional, off or on.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var isExist bool
				for _, v := range sslVerifyParams {
					if v == s {
						isExist = true
						break
					}
				}

				if !isExist {
					return cerr.NewInvalidIngressAnnotationsError(sslVerifyAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	sslServerNameAnnotations: {
		Doc: "optional, off or on.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var isExist bool
				for _, v := range sslVerifyParams {
					if v == s {
						isExist = true
						break
					}
				}

				if !isExist {
					return cerr.NewInvalidIngressAnnotationsError(sslServerNameAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	sslNameAnnotations: {
		Doc: "optional, a host in ingress.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var isExist bool

				hosts := ing.GetHosts()
				for _, v := range hosts {
					if v == s {
						isExist = true
					}
				}

				if !isExist {
					return cerr.NewInvalidIngressAnnotationsError(sslNameAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	sslRedirectAnnotations: {
		Doc: "optional, true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(sslRedirectAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
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

	config.SslRedirect, err = parser.GetBoolAnnotations(sslRedirectAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SslVerify, err = parser.GetStringAnnotation(sslVerifyAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SslServerName, err = parser.GetStringAnnotation(sslServerNameAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SslName, err = parser.GetStringAnnotation(sslNameAnnotations, s.ingress, sslAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

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

	if config.SslRedirect {
		if verr := s.validate(config); verr != nil {
			return config, verr
		}
	}

	return config, nil
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
