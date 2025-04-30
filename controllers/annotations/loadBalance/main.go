package loadBalance

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strings"
)

var (
	Policy = []string{"ip_hash", "random", "least_conn", "hash $request_uri consistent"}
)

const (
	lbPolicyAnnotation  = "lb-policy"
	lbWeightAnnotation  = "lb-weight"
	lbHAAnnotation      = "lb-ha"
	lbStandbyAnnotation = "lb-standby"
)

type loadBalanceIng struct {
	ingress service.K8sResourcesIngress
}

type Config struct {
	LbPolicy   map[string]interface{} `json:"lb-policy"`
	LbWeight   map[string]interface{} `json:"lb-weight"`
	LbHa       map[string]interface{} `json:"lb-ha"`
	LbStandby  map[string]interface{} `json:"lb-standby"`
	StreamName string                 `json:"stream_name"`
}

var loadBalanceAnnotations = parser.AnnotationsContents{
	lbPolicyAnnotation: {
		Doc: fmt.Sprintf("nginx lb policy, the value of the flag must be selected from here: %v.", strings.Join(Policy, ",")),
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				for _, v := range Policy {
					if s == v {
						return nil
					}
				}

				return cerr.NewInvalidIngressAnnotationsError(s, ing.GetName(), ing.GetNameSpace())
			}

			return nil
		},
	},
	lbWeightAnnotation: {
		Doc: "nginx lb weight, such as: backend01 weight 80;backend02 weight 20.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				data, err := jsonParser.JSONToMap(s)
				if err != nil {
					return err
				}

				for k, _ := range data {
					ing.GetBackend(k)
				}

			}

			return nil
		},
	},
	lbHAAnnotation: {
		Doc: "",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {

			}

			return nil
		},
	},
	lbStandbyAnnotation: {
		Doc: "",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {

			}

			return nil
		},
	},
}

func NewLoadBalanceIng(ingress service.K8sResourcesIngress) parser.IngressAnnotationsParser {
	return &loadBalanceIng{
		ingress: ingress,
	}
}

func (r *loadBalanceIng) Parse(ing service.K8sResourcesIngress) (interface{}, error) {
	return nil, nil
}

func (r *loadBalanceIng) Validate(ing map[string]string) error {
	return nil
}
