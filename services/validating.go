package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"golang.org/x/net/context"
)

type ValidatingServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	cert    service.K8sResourcesCert
}
