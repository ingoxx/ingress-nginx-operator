package parser

import (
	"errors"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"regexp"
)

func IsRegex(str string) bool {
	pattern := `[.*+?^${}()|\[\]\\]`
	re := regexp.MustCompile(pattern)
	return re.MatchString(str)
}

func IsSpecificPrefix(annotation string) bool {
	pattern := `^` + constants.AnnotationsPrefix + "/"
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(annotation) != nil
}

func CheckAnnotations(annotations map[string]string, config AnnotationsContents, ing service.K8sResourcesIngress) error {
	var err error
	for annotation := range annotations {
		if !IsSpecificPrefix(annotation) {
			continue
		}

		annKey := GetAnnotationSuffix(annotation)
		annVal := ing.GetAnnotations()[annotation]
		if cfg, ok := config[annKey]; ok && cfg.Validator(annVal, ing) != nil {
			err = errors.Join(cerr.NewAnnotationValidationFailedError(annKey, ing.GetName(), ing.GetNameSpace()))
		}
	}

	return err
}

func CheckAnnotationsKeyVal(name string, ing service.K8sResourcesIngress, config AnnotationsContents) (string, error) {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return "", cerr.NewMissIngressAnnotationsError("", ing.GetName(), ing.GetNameSpace())
	}

	annotationFullName := GetAnnotationKey(name)
	annotationValue, ok := ing.GetAnnotations()[annotationFullName]
	if !ok {
		return "", cerr.NewMissIngressAnnotationsError(annotationFullName, ing.GetName(), ing.GetNameSpace())
	}

	if ok && (annotationValue == "" || config[name].Validator(annotationValue, ing) != nil) {
		return "", cerr.NewInvalidIngressAnnotationsError(name, ing.GetName(), ing.GetNameSpace())
	}

	return annotationFullName, nil
}
