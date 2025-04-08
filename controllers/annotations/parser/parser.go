package parser

import (
	"fmt"
	v1 "k8s.io/api/networking/v1"
	"strings"
)

const (
	AnnotationsPrefix = "ingress.nginx.k8s.io"
)

type IngressAnnotationsParser interface {
	Parse(*v1.Ingress) (interface{}, error)
	Validate(map[string]string) error
}

type AnnotationsContents map[string]AnnotationConfig

// AnnotationConfig 使用说明配置
type AnnotationConfig struct {
	Doc string
}

type GetAnnotations map[string]string

func (g GetAnnotations) parseString() (string, error) {
	return "", nil
}

func (g GetAnnotations) parseBool() (bool, error) {
	return false, nil
}

func (g GetAnnotations) parseSlice() ([]string, error) {
	return []string{}, nil
}

func GetAnnotationSuffix(annotation string) string {
	return strings.TrimPrefix(annotation, AnnotationsPrefix+"/")
}

func GetAnnotationKey(suffix string) string {
	return fmt.Sprintf("%v/%v", AnnotationsPrefix, suffix)
}
