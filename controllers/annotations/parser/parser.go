package parser

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	v1 "k8s.io/api/networking/v1"
	"regexp"
	"strconv"
	"strings"
)

type IngressAnnotationsParser interface {
	Parse(*v1.Ingress) (interface{}, error)
	Validate(map[string]string) error
}

type AnnotationValidator func(s string, ing *v1.Ingress) error
type AnnotationsContents map[string]AnnotationConfig

// AnnotationConfig 使用说明配置
type AnnotationConfig struct {
	Doc       string
	Validator AnnotationValidator
}

type GetAnnotationVal map[string]string

func (g GetAnnotationVal) parseString(name string, ing *v1.Ingress) (string, error) {
	val, ok := g[name]
	if ok {
		if val == "" {
			return "", cerr.NewInvalidIngressAnnotationsError(name, ing.Name, ing.Namespace)
		}

		return val, nil
	}

	return "", cerr.NewMissIngressAnnotationsError(name, ing.Name, ing.Namespace)
}

func (g GetAnnotationVal) parseBool(name string, ing *v1.Ingress) (bool, error) {
	val, ok := g[name]
	if ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, cerr.NewInvalidIngressAnnotationsError(name, ing.Name, ing.Namespace)
		}

		return b, nil
	}

	return false, cerr.NewMissIngressAnnotationsError(name, ing.Name, ing.Namespace)
}

func (g GetAnnotationVal) parseSlice(name string, ing *v1.Ingress) ([]string, error) {
	var data = make([]string, 5)
	val, ok := g[name]
	if ok {
		if val == "" {
			return data, cerr.NewInvalidIngressAnnotationsError(name, ing.Name, ing.Namespace)
		}

		return data, nil
	}

	return data, cerr.NewMissIngressAnnotationsError(name, ing.Name, ing.Namespace)
}

func GetDnsRegex(str string) string {
	p := `([a0-z9]+\.)+([a-z]+)`
	matched := regexp.MustCompile(p)
	dns := matched.FindStringSubmatch(str)
	if len(dns) == 0 {
		return ""
	}

	return dns[0]
}

func GetStringAnnotation(name string, ing *v1.Ingress, config AnnotationsContents) (string, error) {
	key, err := CheckAnnotationsKeyVal(name, ing, config)
	if err != nil {
		return "", err
	}
	return GetAnnotationVal(ing.GetAnnotations()).parseString(key, ing)
}

func GetBoolAnnotations(name string, ing *v1.Ingress, config AnnotationsContents) (bool, error) {
	key, err := CheckAnnotationsKeyVal(name, ing, config)
	if err != nil {
		return false, err
	}
	return GetAnnotationVal(ing.GetAnnotations()).parseBool(key, ing)
}

func GetAnnotationSuffix(annotation string) string {
	return strings.TrimPrefix(annotation, constants.AnnotationsPrefix+"/")
}

func GetAnnotationKey(suffix string) string {
	return fmt.Sprintf("%v/%v", constants.AnnotationsPrefix, suffix)
}
