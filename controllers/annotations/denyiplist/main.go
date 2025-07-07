package denyiplist

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
)

const (
	enableIpBlackListAnnotations = "enable-ip-blacklist"
	setIpBlackConfigAnnotations  = "set-ip-black-config"
)

type enableIpBlackListIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type IpDenyConfig struct {
	Backend []string `json:"backend"`
	Ip      []string `json:"ip"`
}

type IpDenyBackendsConfig struct {
	Backends []ingress.IpListBackendsConfig
}

type Config struct {
	EnableIpBlackList bool   `json:"enable-ip-blacklist"`
	SetIpBlackConfig  string `json:"set-ip-black-config"`
	DenyIpConfig      IpDenyBackendsConfig
}

var enableIpBlackListIngAnnotations = parser.AnnotationsContents{
	enableIpBlackListAnnotations: {
		Doc: "set true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableIpBlackListAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	setIpBlackConfigAnnotations: {
		Doc: "nginx deny ip access, must be in JSON format, example: {\"ip\": [\"2.2.2.2\"], \"backend\": [\"svc-name\"]}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var lq = new(IpDenyBackendsConfig)
				if err := jsonParser.JSONToStruct(s, lq); err != nil {
					return err
				}

				if parser.IsZeroStruct(lq) {
					return cerr.NewInvalidIngressAnnotationsError(setIpBlackConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
}

func NewEnableIpBlackListIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &enableIpBlackListIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *enableIpBlackListIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}
	config.EnableIpBlackList, err = parser.GetBoolAnnotations(enableIpBlackListAnnotations, r.ingress, enableIpBlackListIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SetIpBlackConfig, err = parser.GetStringAnnotation(setIpBlackConfigAnnotations, r.ingress, enableIpBlackListIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if err := r.validate(config); err != nil {
		return config, err
	}

	return config, err
}

func (r *enableIpBlackListIng) validate(config *Config) error {
	if config.EnableIpBlackList {
		if config.SetIpBlackConfig == "" {
			return cerr.NewMissIngressFieldValueError(setIpBlackConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		var lq = new(IpDenyBackendsConfig)
		if err := jsonParser.JSONToStruct(config.SetIpBlackConfig, lq); err != nil {
			return err
		}

		if parser.IsZeroStruct(lq) {
			return cerr.NewInvalidIngressAnnotationsError(setIpBlackConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		config.DenyIpConfig = *lq
	}

	return nil
}

func (r *enableIpBlackListIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableIpBlackListIngAnnotations, r.ingress)
}
