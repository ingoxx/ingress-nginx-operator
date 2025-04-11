package rewrite

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	v1 "k8s.io/api/networking/v1"
)

var (
	Flags = []string{"last", "break", "redirect", "permanent"}
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteEnableRegexAnnotation = "enable-regex"
	rewriteFlagAnnotation        = "rewrite-flag"
)

type rewrite struct {
	ingress *v1.Ingress
}

type Config struct {
	RewriteTarget string `json:"rewrite-target"`
	RewriteFlag   string `json:"rewrite-flag"`
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
	rewriteFlagAnnotation: {
		Doc: "",
		Validator: func(s string, ing *v1.Ingress) error {
			if s != "" {
				for _, v := range Flags {
					if s == v {
						return nil
					}
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
		return config, err
	}

	config.RewriteFlag, err = parser.GetStringAnnotation(rewriteFlagAnnotation, ing, rewriteAnnotations)
	if err != nil {
		return config, err
	}

	config.EnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, ing, rewriteAnnotations)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (r *rewrite) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, rewriteAnnotations, r.ingress)
}
