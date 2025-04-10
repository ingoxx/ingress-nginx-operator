package rewrite

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteEnableRegexAnnotation = "enable-regex"
)

type rewrite struct {
	ingress *v1.Ingress
}

type Config struct {
	RewriteTarget string `json:"rewrite-target"`
	EnableRegex   bool   `json:"enable-regex"`
}

var rewriteAnnotations = parser.AnnotationsContents{
	rewriteTargetAnnotation: {
		Doc: "",
		Validator: func(s string, ing *v1.Ingress) error {
			if s != "" && !parser.IsRegex(s) {
				return cerr.NewInvalidIngressAnnotationsError(s, ing.Name, ing.Namespace)
			}
			return nil
		},
	},
	rewriteEnableRegexAnnotation: {
		Doc: "",
		Validator: func(s string, ing *v1.Ingress) error {
			if s != "" {
				if s == "false" || s == "true" {
					return nil
				}

				return cerr.NewInvalidIngressAnnotationsError(s, ing.Name, ing.Namespace)
			}

			return nil
		},
	},
}

func NewRewrite(ingress *v1.Ingress) parser.IngressAnnotationsParser {
	return &rewrite{
		ingress: ingress,
	}
}

func (r *rewrite) Parse(ing *v1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.RewriteTarget, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing, rewriteAnnotations)
	if err != nil {
		if cerr.IsInvalidIngressAnnotationsError(err) {
			klog.Warningf("%s is invalid value, defaulting to empty", rewriteTargetAnnotation)
		}
	}

	config.EnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, ing, rewriteAnnotations)
	if err != nil {
		if cerr.IsInvalidIngressAnnotationsError(err) {
			klog.Warningf("%s is invalid value, defaulting to false", rewriteEnableRegexAnnotation)
		}
	}

	return config, nil
}

func (r *rewrite) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, rewriteAnnotations, r.ingress)
}
