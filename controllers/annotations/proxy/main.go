package proxy

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
)

const (
	httpUpgradeAnnotations = "http-upgrade"
)

type UpgradePoxy struct {
	ingress   service.K8sResourcesIngress
	resources service.ResourcesMth
}

type Config struct {
	HttpUpgrade string `json:"http-upgrade"`
}

var upgradePoxyAnnotations = parser.AnnotationsContents{
	httpUpgradeAnnotations: {
		Doc: "optional, $http_upgrade",
		Validator: func(s string, ing service.K8sResourcesIngress) error {
			if s != "" {
				if s != "$http_upgrade" {
					return cerr.NewInvalidIngressAnnotationsError(httpUpgradeAnnotations, ing.GetName(), ing.GetNameSpace())
				}
			}
			return nil
		},
	},
}

func NewUpgradePoxy(ingress service.K8sResourcesIngress, resources service.ResourcesMth) parser.IngressAnnotationsParser {
	return &UpgradePoxy{
		ingress:   ingress,
		resources: resources,
	}
}

func (u *UpgradePoxy) Parse() (interface{}, error) {
	var err error
	config := &Config{}

	config.HttpUpgrade, err = parser.GetStringAnnotation(httpUpgradeAnnotations, u.ingress, upgradePoxyAnnotations)
	if err != nil && !cerr.IsMissIngressAnnotationsError(err) {
		return config, err
	}

	return config, nil
}

func (u *UpgradePoxy) Validate(ing map[string]string) error {
	return parser.CheckAnnotations(ing, upgradePoxyAnnotations, u.ingress)
}
