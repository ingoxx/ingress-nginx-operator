package stream

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"k8s.io/apimachinery/pkg/types"
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
		Doc: "nginx stream, support cross namespace, example: \"backends\": [ {\"name\": \"svc-1\", \"namespace\": \"web\", \"port\": 8080}, {\"name\": \"svc-2\", \"namespace\": \"api\", \"port\": 8081} ]",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			var bks = new(ingress.StreamBackendList)
			if s != "" {
				if err := jsonParser.JSONToStruct(s, bks); err != nil {
					return err
				}

				var isExistsSvc string

				for _, v := range bks.Backends {
					if isExistsSvc == v.Name {
						return cerr.NewDuplicateValueErrorError(v.Name, ing.GetName(), ing.GetNameSpace())
					}

					if isExistsSvc == "" {
						isExistsSvc = v.Name
					}

					var isExistsPort bool
					key := types.NamespacedName{Name: v.Name, Namespace: v.Namespace}
					if _, err := ing.GetService(key); err != nil {
						return err
					}

					ports, err := ing.GetBackendPorts(key)
					if err != nil {
						return err
					}

					for _, p := range ports {
						if p.Number == v.Port {
							isExistsPort = true
						}
					}

					if !isExistsPort {
						return cerr.NewInvalidSvcPortError(v.Name, ing.GetName(), ing.GetNameSpace())
					}
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
