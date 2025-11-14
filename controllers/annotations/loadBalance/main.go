package loadBalance

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
	"strings"
)

var (
	Policy = []string{"ip_hash", "random", "least_conn", "hash $request_uri consistent"}
)

const (
	lbPolicyAnnotations       = "lb-policy"
	lbConfigAnnotations       = "lb-config"
	enableLbConfigAnnotations = "enable-lb"
)

type loadBalanceIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	LbConfig []*ingress.Backends `json:"lb-config"`
	LbPolicy string              `json:"lb-policy"`
	EnableLb bool                `json:"enable-lb"`
}

var loadBalanceAnnotations = parser.AnnotationsContents{
	enableLbConfigAnnotations: {
		Doc: "set true or false to use limitconn.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableLbConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
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
		Doc: "nginx lb config, same as the official configuration requirements of nginx, must be in JSON format",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				var bks = new(ingress.LbConfigList)
				if err := jsonParser.JSONToStruct(s, bks); err != nil {
					return err
				}

				if parser.IsZeroStruct(bks) {
					return cerr.NewInvalidIngressAnnotationsError(lbConfigAnnotations, ing.GetName(), ing.GetNameSpace())
				}

				for _, v := range bks.Backends {
					if _, err := ing.GetBackend(v.Name); err != nil {
						//return cerr.NewInvalidIngressAnnotationsError(lbConfigAnnotations, ing.GetName(), ing.GetNameSpace())
						return err
					}

					if !ing.CheckHost(v.Host) {
						return cerr.NewIngressHostNotFoundError(v.Host, ing.GetName(), ing.GetNameSpace())
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
	var bks = new(ingress.LbConfigList)

	config.LbPolicy, err = parser.GetStringAnnotation(lbPolicyAnnotations, r.ingress, loadBalanceAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	lbConfig, err := parser.GetStringAnnotation(lbConfigAnnotations, r.ingress, loadBalanceAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	tls, err := r.resources.GetTlsFile()
	if err != nil {
		return config, err
	}

	upstreamConfig, err := r.resources.GetUpstreamConfig()
	if err != nil {
		return config, err
	}

	if lbConfig != "" {
		if err := jsonParser.JSONToStruct(lbConfig, bks); err != nil {
			return config, err
		}
	}

	for _, v1 := range upstreamConfig {
		v1.Cert = tls[v1.Host]
		for _, svc := range v1.ServiceBackend {
			if svc.Services.Name == "" {
				continue
			}
			svc.BackendDns = r.resources.GetBackendName(svc.Services)
			var updated bool
			for _, v3 := range bks.Backends {
				if svc.Services.Name == "" {
					continue
				}
				if svc.Services.Name == v3.Name && v1.Host == v3.Host {
					if svc.IsSingleService {
						svc.Services.Name = r.resources.GetBackendName(svc.Services)
					} else {
						svc.Services.Name = fmt.Sprintf("%s %s", r.resources.GetBackendName(svc.Services), v3.Config)
					}
					updated = true
					break // 找到后就退出
				}
			}
			if !updated {
				svc.Services.Name = r.resources.GetBackendName(svc.Services)
			}
		}
	}

	config.LbConfig = upstreamConfig

	return config, err
}

func (r *loadBalanceIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, loadBalanceAnnotations, r.ingress)
}
