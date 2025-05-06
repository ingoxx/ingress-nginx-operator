package loadBalance

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	v1 "k8s.io/api/networking/v1"
	"strings"
)

var (
	Policy = []string{"ip_hash", "random", "least_conn", "hash $request_uri consistent"}
)

const (
	lbPolicyAnnotations = "lb-policy"
	lbConfigAnnotations = "lb-config"
)

type loadBalanceIng struct {
	ingress service.K8sResourcesIngress
}

type Config struct {
	Upstream []UpstreamList `json:"upstream"`
	LbPolicy string         `json:"lb-policy"`
}

type UpstreamList struct {
	SvcList    []*v1.ServiceBackendPort `json:"svc-list"`
	StreamName string                   `json:"stream-name"`
}

var loadBalanceAnnotations = parser.AnnotationsContents{
	lbPolicyAnnotations: {
		Doc: fmt.Sprintf("nginx lb policy, the value of the flag must be selected from here: %v.", strings.Join(Policy, ",")),
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var isValidPolicy bool
				for _, v := range Policy {
					if s == v {
						isValidPolicy = true
					}
				}

				if !isValidPolicy {
					return cerr.NewInvalidIngressAnnotationsError(lbPolicyAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}
			return nil
		},
	},
	lbConfigAnnotations: {
		Doc: "nginx lb config, same as the official configuration requirements of nginx, must be in JSON format, like this: {\"svc-name\": \"max_fails=3 fail_timeout=30s weight=80;...\"}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				data, err := jsonParser.JSONToMap(s)
				if err != nil {
					return err
				}

				for k := range data["services"].(map[string]interface{}) {
					if _, err := ing.GetBackend(k); err != nil {
						return cerr.NewInvalidIngressAnnotationsError(k, ing.GetName(), ing.GetNameSpace())
					}
				}
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

func (r *loadBalanceIng) Parse() (interface{}, error) {
	var err error
	var config *Config
	//var upstream UpstreamList

	lbConfig, err := parser.GetStringAnnotation(lbConfigAnnotations, r.ingress, loadBalanceAnnotations)
	if err != nil {
		return config, err
	}

	data, err := jsonParser.JSONToMap(lbConfig)
	if err != nil {
		return config, err
	}

	for k := range data["services"].(map[string]interface{}) {
		_, err := r.ingress.GetBackend(k)
		if err != nil {
			return config, cerr.NewInvalidIngressAnnotationsError(k, r.ingress.GetName(), r.ingress.GetNameSpace())
		}

	}

	return config, nil
}

func (r *loadBalanceIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, loadBalanceAnnotations, r.ingress)
}
