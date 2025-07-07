package allowiplist

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
)

const (
	enableIpWhiteListAnnotations = "enable-ip-whitelist"
	setIpWhiteConfigAnnotations  = "set-ip-white-config"
)

type enableIpWhiteListIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type IpWhiteConfig struct {
	Backend []string `json:"backend"`
	Ip      []string `json:"ip"`
}

type IpWhiteBackendsConfig struct {
	Backends []ingress.IpListBackendsConfig
}

type Config struct {
	EnableIpWhiteList bool   `json:"enable-ip-whitelist"`
	SetIpWhiteConfig  string `json:"set-ip-white-config"`
	AllowIpConfig     IpWhiteBackendsConfig
}

var enableIpWhiteListIngAnnotations = parser.AnnotationsContents{
	enableIpWhiteListAnnotations: {
		Doc: "set true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableIpWhiteListAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	setIpWhiteConfigAnnotations: {
		Doc: "nginx allow ip access, must be in JSON format, example: {\"ip\": [\"2.2.2.2\"], \"backend\": [\"svc-name\"]}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var lq = new(IpWhiteBackendsConfig)
				if err := jsonParser.JSONToStruct(s, lq); err != nil {
					return err
				}

				if parser.IsZeroStruct(lq) {
					return cerr.NewInvalidIngressAnnotationsError(setIpWhiteConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
}

func NewEnableIpWhiteListIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &enableIpWhiteListIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *enableIpWhiteListIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}
	config.EnableIpWhiteList, err = parser.GetBoolAnnotations(enableIpWhiteListAnnotations, r.ingress, enableIpWhiteListIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SetIpWhiteConfig, err = parser.GetStringAnnotation(setIpWhiteConfigAnnotations, r.ingress, enableIpWhiteListIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if err := r.validate(config); err != nil {
		return config, err
	}

	return config, err
}

func (r *enableIpWhiteListIng) validate(config *Config) error {
	if config.EnableIpWhiteList {
		if config.SetIpWhiteConfig == "" {
			return cerr.NewMissIngressFieldValueError(setIpWhiteConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		var lq = new(IpWhiteBackendsConfig)
		if err := jsonParser.JSONToStruct(config.SetIpWhiteConfig, lq); err != nil {
			return err
		}

		if parser.IsZeroStruct(lq) {
			return cerr.NewInvalidIngressAnnotationsError(setIpWhiteConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		config.AllowIpConfig = *lq
	}

	return nil
}

func (r *enableIpWhiteListIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableIpWhiteListIngAnnotations, r.ingress)
}
