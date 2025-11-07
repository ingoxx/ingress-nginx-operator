package limitreq

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
)

const (
	enableLimitReqAnnotations = "enable-limit-req"
	limitConfigAnnotations    = "set-limit-config"
)

type RequestLimitIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type ZoneConfig struct {
	LimitKey string `json:"limit_key"`
	ZoneName string `json:"zone_name"`
	Capacity string `json:"capacity"`
	Rate     string `json:"rate"` // 10r/s, 10r/m
}

type ReqConfig struct {
	ZoneName string `json:"zone_name"`
	Burst    int    `json:"burst"`
	Delay    bool   `json:"delay"`
}

type ZoneRepConfig struct {
	LimitZone []*ZoneConfig `json:"limit_zone"`
	LimitReq  []*ReqConfig  `json:"limit_req"`
	Name      string        `json:"name"`
}

type SetLimitConfig struct {
	Backends []*ZoneRepConfig `json:"backends"`
}

type Config struct {
	Bs                 SetLimitConfig
	LimitConfig        string `json:"limit-config"`
	EnableRequestLimit bool   `json:"enable-request-limit"`
}

var RequestLimitIngAnnotations = parser.AnnotationsContents{
	enableLimitReqAnnotations: {
		Doc: "set true or false to use limitreq.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableLimitReqAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	limitConfigAnnotations: {
		Doc: "nginx request limit, same as the official configuration requirements of nginx, must be in JSON format",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var lq = new(SetLimitConfig)
				if err := jsonParser.JSONToStruct(s, lq); err != nil {
					return err
				}

				if parser.IsZeroStruct(lq) {
					return cerr.NewInvalidIngressAnnotationsError(limitConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
}

func NewRequestLimitIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &RequestLimitIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *RequestLimitIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}

	config.EnableRequestLimit, err = parser.GetBoolAnnotations(enableLimitReqAnnotations, r.ingress, RequestLimitIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.LimitConfig, err = parser.GetStringAnnotation(limitConfigAnnotations, r.ingress, RequestLimitIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if err := r.validate(config); err != nil {
		return config, err
	}

	return config, err
}

func (r *RequestLimitIng) validate(config *Config) error {
	if config.EnableRequestLimit {
		if config.LimitConfig == "" {
			return cerr.NewMissIngressFieldValueError(limitConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}
		var lq SetLimitConfig
		if err := jsonParser.JSONToStruct(config.LimitConfig, &lq); err != nil {
			return err
		}

		if parser.IsZeroStruct(lq) {
			return cerr.NewInvalidIngressAnnotationsError(limitConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		config.Bs = lq
	}

	return nil
}

func (r *RequestLimitIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, RequestLimitIngAnnotations, r.ingress)
}
