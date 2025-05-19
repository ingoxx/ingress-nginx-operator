package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"golang.org/x/net/context"
)

type SvcServiceImpl struct {
	generic common.Generic
	ctx     context.Context
}

func NewSvcServiceImpl(ctx context.Context, clientSet common.Generic) *SvcServiceImpl {
	return &SvcServiceImpl{ctx: ctx, generic: clientSet}
}

func (d *SvcServiceImpl) GetSvc(name string) error {
	return nil
}

func (d *SvcServiceImpl) CreateSvc() error {
	return nil
}

func (d *SvcServiceImpl) UpdateSvc() error {
	return nil
}

func (d *SvcServiceImpl) DeleteSvc() error {
	return nil
}
