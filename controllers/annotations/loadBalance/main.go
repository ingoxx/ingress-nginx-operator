package loadBalance

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
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
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	LbConfig []*ingress.Backends `json:"lb-config"`
	LbPolicy string              `json:"lb-policy"`
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
		Doc: "nginx lb config, same as the official configuration requirements of nginx, must be in JSON format, example: {\"svc-name\": \"max_fails=3 fail_timeout=30s weight=80;...\"}",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				data, err := jsonParser.JSONToMap(s)
				if err != nil {
					return err
				}

				for k := range data {
					if _, err := ing.GetBackend(k); err != nil {
						return cerr.NewInvalidIngressAnnotationsError(lbConfigAnnotations, ing.GetName(), ing.GetNameSpace())
					}
				}
			}

			return nil
		},
	},
}

func NewLoadBalanceIng(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &loadBalanceIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *loadBalanceIng) Parse() (interface{}, error) {
	var err error
	var config = new(Config)

	config.LbPolicy, err = parser.GetStringAnnotation(lbPolicyAnnotations, r.ingress, loadBalanceAnnotations)
	if !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	lbConfig, err := parser.GetStringAnnotation(lbConfigAnnotations, r.ingress, loadBalanceAnnotations)
	if !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	upstreamConfig, err := r.resources.GetUpstreamConfig()
	if err != nil {
		return config, err
	}

	tls, err := r.resources.GetTlsFile()
	if err != nil {
		return config, err
	}

	if lbConfig != "" {
		data, err := jsonParser.JSONToMap(lbConfig)
		if err != nil {
			return config, err
		}

		for av := range data {
			for _, uv := range upstreamConfig {
				uv.Cert = tls[uv.Host]
				//if av == uv.Services.Name {
				//	uv.Services.Name = fmt.Sprintf("%s %s", r.resources.GetBackendName(uv.Services), data[av])
				//}

				for _, sv := range uv.ServiceBackend {
					if av == sv.Services.Name {
						sv.Services.Name = fmt.Sprintf("%s %s", r.resources.GetBackendName(sv.Services), data[av])
					}
				}
			}
		}
	} else {
		for _, uv := range upstreamConfig {
			uv.Cert = tls[uv.Host]
			//uv.Services.Name = r.resources.GetBackendName(uv.Services)
			for _, sv := range uv.ServiceBackend {
				sv.Services.Name = r.resources.GetBackendName(sv.Services)
			}
		}

	}

	config.LbConfig = upstreamConfig

	return config, nil
}

func (r *loadBalanceIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, loadBalanceAnnotations, r.ingress)
}
