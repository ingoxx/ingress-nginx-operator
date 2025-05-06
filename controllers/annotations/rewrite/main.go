package rewrite

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"strings"
)

var (
	Flags = []string{"last", "break", "redirect", "permanent"}
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteEnableRegexAnnotation = "enable-regex"
	rewriteFlagAnnotation        = "rewrite-flag"
)

type rewriteIng struct {
	ingress service.K8sResourcesIngress
}

type Config struct {
	RewriteTarget string `json:"rewrite-target"`
	RewriteFlag   string `json:"rewrite-flag"`
	EnableRegex   bool   `json:"enable-regex"`
}

var rewriteAnnotations = parser.AnnotationsContents{
	rewriteTargetAnnotation: {
		Doc: "rewrite target path, like: /$1, /$2, /api. required",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" && !parser.IsRegex(s) {
				return cerr.NewInvalidIngressAnnotationsError(s, ing.GetName(), ing.GetNameSpace())
			}
			return nil
		},
	},
	rewriteEnableRegexAnnotation: {
		Doc: "optional",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if s == "false" || s == "true" {
					return nil
				}

				return cerr.NewInvalidIngressAnnotationsError(s, ing.GetName(), ing.GetNameSpace())
			}

			return nil
		},
	},
	rewriteFlagAnnotation: {
		Doc: fmt.Sprintf("rewrite flag, the value of the flag must be selected from here, '%v'. required", strings.Join(Flags, ",")),
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				for _, v := range Flags {
					if s == v {
						return nil
					}
				}

				return cerr.NewInvalidIngressAnnotationsError(s, ing.GetName(), ing.GetNameSpace())
			}

			return nil
		},
	},
}

func NewRewrite(ingress service.K8sResourcesIngress) parser.IngressAnnotationsParser {
	return &rewriteIng{
		ingress: ingress,
	}
}

func (r *rewriteIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}
	config.RewriteTarget, err = parser.GetStringAnnotation(rewriteTargetAnnotation, r.ingress, rewriteAnnotations)
	if err != nil {
		return config, err
	}

	config.RewriteFlag, err = parser.GetStringAnnotation(rewriteFlagAnnotation, r.ingress, rewriteAnnotations)
	if err != nil {
		return config, err
	}

	config.EnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, r.ingress, rewriteAnnotations)
	if err != nil {
		return config, err
	}

	if err = r.validate(config); err != nil {
		return config, err
	}

	return config, nil
}

func (r *rewriteIng) validate(config *Config) error {
	if (config.RewriteTarget == "") != (config.RewriteFlag == "") {
		return cerr.NewInvalidIngressAnnotationsError(rewriteFlagAnnotation+"/"+rewriteTargetAnnotation,
			r.ingress.GetName(),
			r.ingress.GetNameSpace())
	}

	return nil
}

func (r *rewriteIng) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, rewriteAnnotations, r.ingress)
}
