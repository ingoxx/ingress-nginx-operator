package parser

import (
	"errors"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	v1 "k8s.io/api/networking/v1"
	"regexp"
)

func IsRegex(str string) bool {
	pattern := `^\/(\w+)|^/\$([0-9])$|^\$([0-9])$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func IsSpecificPrefix(annotation string) bool {
	pattern := `^` + constants.AnnotationsPrefix + "/"
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(annotation) != nil
}

func CheckAnnotations(annotations map[string]string, config AnnotationsContents, ing *v1.Ingress) error {
	var err error
	for annotation := range annotations {
		if !IsSpecificPrefix(annotation) {
			continue
		}

		annKey := GetAnnotationSuffix(annotation)
		annVal := ing.GetAnnotations()[annotation]
		if cfg, ok := config[annKey]; ok && cfg.Validator(annVal, ing) != nil {
			err = errors.Join(err)
		}
	}

	return err
}

func CheckAnnotationsKeyVal(name string, ing *v1.Ingress, config AnnotationsContents) (string, error) {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return "", cerr.NewMissIngressAnnotationsError("", ing.Name, ing.Namespace)
	}

	annotationFullName := GetAnnotationKey(name)
	annotationValue := ing.GetAnnotations()[annotationFullName]

	if annotationValue == "" {
		return "", cerr.NewInvalidIngressAnnotationsError(name, ing.Name, ing.Namespace)
	}

	if err := config[name].Validator(annotationValue, ing); err != nil {
		return "", cerr.NewInvalidIngressAnnotationsError(name, ing.Name, ing.Namespace)
	}

	return annotationFullName, nil
}
