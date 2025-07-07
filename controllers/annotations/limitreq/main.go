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
	resources service.ResourcesMth
}

type ReqLimitConfig struct {
	LimitReqZone []string `json:"limit_req_zone"`
	LimitReq     []string `json:"limit_req"`
	Backend      string   `json:"backend"`
}

type ReqBackendsConfig struct {
	Backends []ReqLimitConfig `json:"backends"`
}

type Config struct {
	EnableRequestLimit bool   `json:"enable-request-limit"`
	SetLimitConfig     string `json:"set-limit-config"`
	ReqBackendsConfig
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
		Doc: "nginx request limit, same as the official configuration requirements of nginx, must be in JSON format, example: {\"path\": [\"svc_name\"], \"limit_req_zone\": \"$binary_remote_addr$request_uri zone=per_ip_uri:10m rate=5r/s;\"}.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var lq = new(ReqBackendsConfig)
				if err := jsonParser.JSONToStruct(s, lq); err != nil {
					return err
				}

				if parser.IsZeroStruct(lq) {
					return cerr.NewInvalidIngressAnnotationsError(setLimitConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}

				//if lq.LimitReq == "" || lq.LimitReqZone == "" || len(lq.Backend) == 0 {
				//	return cerr.NewInvalidIngressAnnotationsError(setLimitConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				//}
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

	config.SetLimitConfig, err = parser.GetStringAnnotation(setLimitConfigAnnotations, r.ingress, RequestLimitIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if verr := r.validate(config); verr != nil {
		return config, verr
	}

	return config, err
}

func (r *RequestLimitIng) validate(config *Config) error {
	if config.EnableRequestLimit {
		if config.SetLimitConfig == "" {
			return cerr.NewMissIngressFieldValueError(setLimitConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}
		var lq = new(ReqBackendsConfig)
		if err := jsonParser.JSONToStruct(config.SetLimitConfig, lq); err != nil {
			return err
		}

		if parser.IsZeroStruct(lq) {
			return cerr.NewInvalidIngressAnnotationsError(setLimitConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		config.ReqBackendsConfig = *lq
	}

	return nil
}

func (r *RequestLimitIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, RequestLimitIngAnnotations, r.ingress)
}
