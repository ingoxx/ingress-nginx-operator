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
	setLimitConfigAnnotations = "set-limit-config"
)

type RequestLimitIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesData
}

type Config struct {
	EnableRequestLimit bool   `json:"enable-request-limit"`
	SetLimitConfig     string `json:"set-limit-config"`
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
	setLimitConfigAnnotations: {
		Doc: "nginx request limit, same as the official configuration requirements of nginx, must be in JSON format, example: {\"limit_req_zone\": \"$binary_remote_addr$request_uri zone=per_ip_uri:10m rate=5r/s;\"}.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				data, err := jsonParser.JSONToMap(s)
				if err != nil {
					return err
				}

				for k := range data {
					v, ok := data[k]
					if !ok || v == nil {
						return cerr.NewInvalidIngressAnnotationsError(setLimitConfigAnnotations, ing.GetName(), ing.GetNameSpace())
					}
				}
			}

			return nil
		},
	},
}

func NewRequestLimitIng(ingress service.K8sResourcesIngress, resources service.ResourcesData) parser.IngressAnnotationsParser {
	return &RequestLimitIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *RequestLimitIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}

	config.EnableRequestLimit, err = parser.GetBoolAnnotations(enableLimitReqAnnotations, r.ingress, RequestLimitIngAnnotations)
	if !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SetLimitConfig, err = parser.GetStringAnnotation(setLimitConfigAnnotations, r.ingress, RequestLimitIngAnnotations)
	if !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if err = r.validate(config); err != nil {
		return config, err
	}

	return config, nil
}

func (r *RequestLimitIng) validate(config *Config) error {
	if config.EnableRequestLimit && config.SetLimitConfig == "" {
		return cerr.NewMissIngressFieldValueError(setLimitConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
	}

	return nil
}

func (r *RequestLimitIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, RequestLimitIngAnnotations, r.ingress)
}
