package loadBalance

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	v12 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
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

				var isInIng bool
				for _, v := range bks.Backends {

					key := types.NamespacedName{Name: v.Name, Namespace: ing.GetNameSpace()}
					svc, err := ing.GetService(key)
					if err != nil {
						return err
					}

					var isExistPort bool
					port := ing.GetSvcPort(svc)
					for _, p := range port {
						if p == v.Port {
							isExistPort = true
							break
						}
					}

					if !isExistPort {
						return fmt.Errorf("service '%s' port error", svc.Name)
					}

					backendPort := ing.GetBackendPort(svc)
					if backendPort > 0 {
						isInIng = true
					}
				}

				if !isInIng {
					return fmt.Errorf("the backend is not in ingress '%s'", ing.GetName())
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

	var isE = make(map[string]struct{})
	for _, v1 := range upstreamConfig {
		v1.Cert = tls[v1.Host]
		var st = make([]string, 0, len(upstreamConfig))

		if len(bks.Backends) > 0 {
			// 这里是使用upstream功能，支持自定义权重，策略，超时等设置
			for _, v2 := range bks.Backends {
				svc := &v12.ServiceBackendPort{
					Name:   v2.Name,
					Number: v2.Port,
				}

				if v1.Host == v2.Host {
					bn := fmt.Sprintf("%s %s", r.resources.GetBackendName(svc), v2.Config)
					_, ok := isE[v2.Name]
					if !ok {
						st = append(st, bn)
						isE[v2.Name] = struct{}{}
					}
				}
			}

			v1.StreamServeName = st
		} else {
			// 这里是普通的后端地址：http://svc.namespace.svc:port
			v1.Upstream = ""
			for _, ib := range v1.ServiceBackend {
				ib.BackendDns = r.resources.GetBackendName(ib.Services)
			}
		}

	}

	config.LbConfig = upstreamConfig

	return config, err
}

func (r *loadBalanceIng) validate(config *Config) error {
	for _, v1 := range config.LbConfig {
		if len(v1.ServiceBackend) > 1 {
			v1.ServiceBackend = v1.ServiceBackend[:1]
		}
	}

	return nil
}

func (r *loadBalanceIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, loadBalanceAnnotations, r.ingress)
}
