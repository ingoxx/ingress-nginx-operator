package parser

import v1 "k8s.io/api/networking/v1"

type IngressAnnotationsParser interface {
	Parse(*v1.Ingress) (interface{}, error)
	Validate(map[string]string) error
}
