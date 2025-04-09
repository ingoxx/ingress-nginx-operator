package annotations

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/rewrite"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressAnnotationsKey struct {
	metav1.ObjectMeta
	Rewrite rewrite.Config
}

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

func (e *Extractor) Extract() *IngressAnnotationsKey {
	in := &IngressAnnotationsKey{
		ObjectMeta: e.ingress.ObjectMeta,
	}

	return in
}
