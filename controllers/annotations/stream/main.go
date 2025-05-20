package stream

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
)

const (
	enableStreamAnnotations    = "enable-stream"
	setStreamConfigAnnotations = "set-stream-config"
)

type enableStreamIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	EnableStream      bool                     `json:"enable-stream"`
	SetStreamConfig   string                   `json:"set-stream-config"`
	StreamBackendList []*ingress.StreamBackend `json:"stream-backend"`
}

var enableStreamIngAnnotations = parser.AnnotationsContents{
	enableStreamAnnotations: {
		Doc: "set true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableStreamAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	setStreamConfigAnnotations: {
		Doc: "nginx stream, support cross namespace, example: \"backends\": [ {\"name\": \"svc-1\", \"namespace\": \"web\"}, {\"name\": \"svc-2\", \"namespace\": \"api\"} ]",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			var bks = new(ingress.StreamBackendList)
			if s != "" {
				if err := jsonParser.JSONToStruct(s, bks); err != nil {
					return err
				}
			}

			return nil
		},
	},
}

func NewEnableStreamIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &enableStreamIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *enableStreamIng) Parse() (interface{}, error) {
	var err error
	var config = new(Config)

	config.EnableStream, err = parser.GetBoolAnnotations(enableStreamAnnotations, r.ingress, enableStreamIngAnnotations)
	if !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SetStreamConfig, err = parser.GetStringAnnotation(setStreamConfigAnnotations, r.ingress, enableStreamIngAnnotations)
	if !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	return config, err
}

func (r *enableStreamIng) validate(config *Config) error {
	var bks = new(ingress.StreamBackendList)
	if config.EnableStream && config.SetStreamConfig == "" {
		return cerr.NewInvalidIngressAnnotationsError(enableStreamAnnotations+","+setStreamConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
	}

	if err := jsonParser.JSONToStruct(config.SetStreamConfig, bks); err != nil {
		return err
	}

	config.StreamBackendList = bks.Backends

	return nil
}

func (r *enableStreamIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableStreamIngAnnotations, r.ingress)
}
