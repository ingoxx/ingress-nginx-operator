package stream

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	v1 "k8s.io/api/networking/v1"
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

// BackendList annotations中的序列化结构
type BackendList struct {
	Backends []*Backend `json:"backends"`
}

type Backend struct {
	Name              string `json:"name"`
	StreamBackendName string `json:"-"`
	Port              int32  `json:"port"`
}

type Config struct {
	EnableStream      bool       `json:"enable-stream"`
	SetStreamConfig   string     `json:"set-stream-config"`
	StreamBackendList []*Backend `json:"stream-backend"`
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
		Doc: "nginx stream, must be in JSON format, example: {\"backends\": [ {\"name\": \"svcName-1\", \"port\": 8080}, {\"name\": \"svcName-2\", \"port\": 8081}... ]}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var bks = new(BackendList)
				if err := jsonParser.JSONToStruct(s, bks); err != nil {
					return err
				}

				if parser.IsZeroStruct(bks) {
					return cerr.NewInvalidIngressAnnotationsError(setStreamConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}

				var isExistsSvc string
				for _, v := range bks.Backends {
					if isExistsSvc == v.Name {
						return cerr.NewDuplicateValueError(v.Name, ing.GetName(), ing.GetNameSpace())
					}

					if isExistsSvc == "" {
						isExistsSvc = v.Name
					}

					var isExistsPort bool
					key := types.NamespacedName{Name: v.Name, Namespace: ing.GetNameSpace()}
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
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.SetStreamConfig, err = parser.GetStringAnnotation(setStreamConfigAnnotations, r.ingress, enableStreamIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if verr := r.validate(config); verr != nil {
		return config, verr
	}

	return config, err
}

func (r *enableStreamIng) validate(config *Config) error {
	if config.EnableStream {
		var bks = new(BackendList)

		if config.SetStreamConfig == "" {
			return cerr.NewInvalidIngressAnnotationsError(enableStreamAnnotations+","+setStreamConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		if err := jsonParser.JSONToStruct(config.SetStreamConfig, bks); err != nil {
			return err
		}

		if parser.IsZeroStruct(bks) {
			return cerr.NewInvalidIngressAnnotationsError(setStreamConfigAnnotations, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

		for _, v := range bks.Backends {
			sp := &v1.ServiceBackendPort{
				Name:   v.Name,
				Number: v.Port,
			}

			v.StreamBackendName = r.ingress.GetBackendName(sp)

		}

		config.StreamBackendList = bks.Backends
	}

	return nil
}

func (r *enableStreamIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableStreamIngAnnotations, r.ingress)
}
