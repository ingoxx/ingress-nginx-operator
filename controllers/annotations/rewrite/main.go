package rewrite

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	v1 "k8s.io/api/networking/v1"
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteEnableRegexAnnotation = "enable-regex"
)

type rewrite struct {
}

type Config struct {
	RewriteTarget string `json:"rewrite-target"`
	EnableRegex   bool   `json:"enable-regex"`
}

var rewriteAnnotations = parser.AnnotationsContents{
	rewriteTargetAnnotation: {
		Doc: "",
	},
	rewriteEnableRegexAnnotation: {
		Doc: "",
	},
}

func NewRewrite() parser.IngressAnnotationsParser {
	return &rewrite{}
}

func (r *rewrite) Parse(*v1.Ingress) (interface{}, error) {

	return nil, nil
}

func (r *rewrite) Validate(map[string]string) error {
	return nil
}
