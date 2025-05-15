package annotations

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/loadBalance"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/rewrite"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/klog/v2"
)

type IngressAnnotationsConfig struct {
	Rewrite     rewrite.Config
	LoadBalance loadBalance.Config
}

func (iac *IngressAnnotationsConfig) GetIngAnnConfig() {}

type Extractor struct {
	annotations map[string]parser.IngressAnnotationsParser
	ingress     service.K8sResourcesIngress
}

func NewExtractor(ing service.K8sResourcesIngress) *Extractor {
	return &Extractor{
		annotations: map[string]parser.IngressAnnotationsParser{
			"Rewrite":     rewrite.NewRewrite(ing),
			"LoadBalance": loadBalance.NewLoadBalanceIng(ing),
		},
		ingress: ing,
	}
}

func (e *Extractor) Extract() (*IngressAnnotationsConfig, error) {
	iak := new(IngressAnnotationsConfig)

	ia := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		if err := annotationParser.Validate(e.ingress.GetAnnotations()); err != nil {
			klog.Error(err)
			return nil, err
		}

		val, err := annotationParser.Parse()
		if err != nil {
			if cerr.IsMissIngressAnnotationsError(err) {
				continue
			}

			return nil, err
		}

		if val != nil {
			ia[name] = val
		}
	}

	if err := mergo.MapWithOverwrite(iak, ia); err != nil {
		klog.ErrorS(err, fmt.Sprintf("unexpected error merging extracted annotations, ingress '%s', namespace '%s'", e.ingress.GetName(), e.ingress.GetNameSpace()))
		return nil, err
	}

	return iak, nil
}
