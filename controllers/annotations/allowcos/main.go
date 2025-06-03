package allowcos

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"strconv"
)

const (
	enableCosAnnotations = "enable-cos"
)

type enableCosIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	EnableCos bool `json:"enable_cos"`
}

var enableCosIngAnnotations = parser.AnnotationsContents{
	enableCosAnnotations: {
		Doc: "optional, true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableCosAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
}

func NewEnableCosIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &enableCosIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (s *enableCosIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}

	config.EnableCos, err = parser.GetBoolAnnotations(enableCosAnnotations, s.ingress, enableCosIngAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	return config, err
}

func (s *enableCosIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableCosIngAnnotations, s.ingress)
}
