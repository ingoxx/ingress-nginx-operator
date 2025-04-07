package annotations

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/rewrite"
	v1 "k8s.io/api/networking/v1"
)

type IngressAnnotationsKey struct {
	Rewrite rewrite.Config
}

type Extractor struct {
	annotations map[string]parser.IngressAnnotationsParser
}

func NewExtractor() *Extractor {
	return &Extractor{
		annotations: map[string]parser.IngressAnnotationsParser{
			"Rewrite": rewrite.NewRewrite(),
		},
	}
}

func (e *Extractor) Extract(ing *v1.Ingress) *IngressAnnotationsKey {
	return &IngressAnnotationsKey{}
}
