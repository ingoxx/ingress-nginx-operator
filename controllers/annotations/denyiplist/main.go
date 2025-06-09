package denyiplist

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"strconv"
)

const (
	enableIpBlackListAnnotations = "enable-ip-blacklist"
	setIpBlackListAnnotations    = "set-ip-blacklist"
)

type enableIpBlackListIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	EnableIpBlackList bool     `json:"enable-ip-blacklist"`
	SetIpBlackList    string   `json:"set-ip-blacklist"`
	Backend           []string `json:"backend"`
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
	setIpBlackListAnnotations: {
		Doc: "nginx stream, support cross namespace, must be in JSON format, example: {\"backends\": [ {\"name\": \"svcName-1\", \"namespace\": \"web\", \"port\": 8080}, {\"name\": \"svcName-2\", \"namespace\": \"api\", \"port\": 8081}... ]}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {

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
	return nil, nil
}

func (r *enableIpBlackListIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, enableIpBlackListIngAnnotations, r.ingress)
}
