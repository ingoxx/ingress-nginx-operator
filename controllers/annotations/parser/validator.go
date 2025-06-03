package parser

import (
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"k8s.io/klog/v2"
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
		if cfg, ok := config[annKey]; ok {
			if err := cfg.Validator(annVal, ing); err != nil {
				klog.Info(cfg.Doc)
				err = errors.Join(cerr.NewAnnotationValidationFailedError(annKey, err.Error(), ing.GetName(), ing.GetNameSpace()))
			}
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
	fmt.Println("annotationFullName >>> ", annotationFullName, ing.GetAnnotations()[annotationFullName], annotationValue)
	if !ok {
		return "", cerr.NewMissIngressAnnotationsError(annotationFullName, ing.GetName(), ing.GetNameSpace())
	}

	if ok {
		if annotationValue == "" {
			return "", cerr.NewInvalidIngressAnnotationsError(name, ing.GetName(), ing.GetNameSpace())
		}

		if config[name].Validator(annotationValue, ing) != nil {
			return "", cerr.NewInvalidIngressAnnotationsError(name, ing.GetName(), ing.GetNameSpace())
		}
	}

	return annotationFullName, nil
}
