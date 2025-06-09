package allowiplist

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"strconv"
)

const (
	enableIpWhiteListAnnotations = "enable-ip-whitelist"
	setIpWhiteListAnnotations    = "set-ip-whitelist"
)

type enableIpWhiteListIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	EnableIpWhiteList bool     `json:"enable-ip-whitelist"`
	SetIpWhiteList    string   `json:"set-ip-whitelist"`
	Backend           []string `json:"backend"`
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
	setIpWhiteListAnnotations: {
		Doc: "nginx stream, support cross namespace, must be in JSON format, example: {\"backends\": [ {\"name\": \"svcName-1\", \"namespace\": \"web\", \"port\": 8080}, {\"name\": \"svcName-2\", \"namespace\": \"api\", \"port\": 8081}... ]}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {

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
	return nil, nil
}

func (r *enableIpWhiteListIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableIpWhiteListIngAnnotations, r.ingress)
}
