package stream

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/jsonParser"
	"strconv"
)

const (
	enableStreamAnnotations    = "enable-stream"
	setStreamConfigAnnotations = "set-stream-config"
)

type enableStreamIng struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesData
}

type Config struct {
	EnableStream    bool   `json:"enable-stream"`
	SetStreamConfig string `json:"set-stream-config"`
}

type BackendList struct {
	Backends []Backend
}

type Backend struct {
	Name      string
	Namespace string
}

var enableStreamIngAnnotations = parser.AnnotationsContents{
	enableStreamAnnotations: {
		Doc: "set true or false.",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if _, err := strconv.ParseBool(s); err != nil {
					return cerr.NewInvalidIngressAnnotationsError(enableStreamAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}

			return nil
		},
	},
	setStreamConfigAnnotations: {
		Doc: "nginx stream, support cross namespace, example: \"backends\": [ {\"name\": \"svc-1\", \"namespace\": \"web\"}, {\"name\": \"svc-2\", \"namespace\": \"api\"} ]",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			var bks = new(BackendList)
			if s != "" {
				if err := jsonParser.JSONToStruct(s, bks); err != nil {
					return err
				}
			}

			return nil
		},
	},
}
