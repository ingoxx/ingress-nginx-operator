package limitconn

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
)

const (
	enableLimitConnAnnotations = "enable-limit-conn"
	limitConfigAnnotations     = "set-limit-conn-config"
)

type ConnLimitIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type ZoneConfig struct {
	LimitKey string `json:"limit_key"`
	ZoneName string `json:"zone_name"`
	Capacity string `json:"capacity"`
}

type ConnConfig struct {
	ZoneName string `json:"zone_name"`
	Burst    int    `json:"burst"`
}

type ZoneConnConfig struct {
	LimitZone []*ZoneConfig `json:"limit_zone"`
	LimitConn []*ConnConfig `json:"limit_conn"`
	Name      string        `json:"name"`
}

type SetLimitConfig struct {
	Backends []*ZoneConnConfig `json:"backends"`
}

type Config struct {
	Bs              SetLimitConfig
	LimitConfig     string `json:"limit-config"`
	EnableConnLimit bool   `json:"enable-conn-limit"`
}

var RequestLimitIngAnnotations = parser.AnnotationsContents{
	enableLimitConnAnnotations: {
		Doc: "set true or false to use limitconn.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableLimitConnAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	limitConfigAnnotations: {
		Doc: "nginx limit conn, same as the official configuration requirements of nginx, must be in JSON format",
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

func NewConnLimitIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &ConnLimitIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *ConnLimitIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}

	config.EnableConnLimit, err = parser.GetBoolAnnotations(enableLimitConnAnnotations, r.ingress, RequestLimitIngAnnotations)
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

func (r *ConnLimitIng) validate(config *Config) error {
	if config.EnableConnLimit {
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

func (r *ConnLimitIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, RequestLimitIngAnnotations, r.ingress)
}
