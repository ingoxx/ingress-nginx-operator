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
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	RewriteTarget string `json:"rewrite-target"`
	RewriteFlag   string `json:"rewrite-flag"`
	EnableRegex   bool   `json:"enable-regex"`
}

var rewriteAnnotations = parser.AnnotationsContents{
	rewriteTargetAnnotation: {
		Doc: "rewrite target path, like: /$1, /$2, /api.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" && !parser.IsRegex(s) {
				return cerr.NewInvalidIngressAnnotationsError(s, ing.GetName(), ing.GetNameSpace())
			}
			return nil
		},
	},
	rewriteEnableRegexAnnotation: {
		Doc: "set true or false to use Regex.",
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
		Doc: fmt.Sprintf("rewrite flag, the value of the flag must be selected from here, '%v'.", strings.Join(Flags, ",")),
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

func NewRewrite(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &rewriteIng{
		ingress:   ingress,
		resources: resources,
	}
}

func (r *rewriteIng) Parse() (interface{}, error) {
	var err error
	config := &Config{}
	config.RewriteTarget, err = parser.GetStringAnnotation(rewriteTargetAnnotation, r.ingress, rewriteAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.RewriteFlag, err = parser.GetStringAnnotation(rewriteFlagAnnotation, r.ingress, rewriteAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	config.EnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, r.ingress, rewriteAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	if verr := r.validate(config); verr != nil {
		return config, verr
	}

	return config, err
}

func (r *rewriteIng) validate(config *Config) error {
	if config.EnableRegex {
		var isValidPath bool
		for _, v := range r.ingress.GetPaths() {
			if parser.IsRegex(v) {
				isValidPath = true
			}
		}

		if !isValidPath {
			return cerr.NewInvalidIngressValueError(r.ingress.GetName(), r.ingress.GetNameSpace())
		}
	}

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
