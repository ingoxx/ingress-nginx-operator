package annotations

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/rewrite"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type IngressAnnotationsConfig struct {
	metav1.ObjectMeta
	Rewrite rewrite.Config
}

func (iac *IngressAnnotationsConfig) GetIngAnnConfig() {}

type Extractor struct {
	annotations map[string]parser.IngressAnnotationsParser
	ingress     *v1.Ingress
}

func NewExtractor(ing *v1.Ingress) *Extractor {
	return &Extractor{
		annotations: map[string]parser.IngressAnnotationsParser{
			"Rewrite": rewrite.NewRewrite(ing),
		},
	}
}

func (e *Extractor) Extract() (*IngressAnnotationsConfig, error) {
	iak := &IngressAnnotationsConfig{
		ObjectMeta: e.ingress.ObjectMeta,
	}

	ia := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		if err := annotationParser.Validate(e.ingress.GetAnnotations()); err != nil {
			klog.ErrorS(err, "")
			return nil, err
		}

		val, err := annotationParser.Parse(e.ingress)
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

	err := mergo.MapWithOverwrite(iak, ia)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("unexpected error merging extracted annotations, ingress '%s', namespace '%s'", e.ingress.Name, e.ingress.Namespace))
		return nil, err
	}

	return nil, nil
}
