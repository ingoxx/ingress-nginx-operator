package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"golang.org/x/net/context"
)

type DaemonSetServiceImpl struct {
	generic common.Generic
	ctx     context.Context
}

func NewDaemonSetServiceImpl(ctx context.Context, clientSet common.Generic) *DaemonSetServiceImpl {
	return &DaemonSetServiceImpl{ctx: ctx, generic: clientSet}
}

func (d *DaemonSetServiceImpl) GetDaemonSet() error {
	return nil
}

func (d *DaemonSetServiceImpl) CreateDaemonSet() error {
	return nil
}

func (d *DaemonSetServiceImpl) UpdateDaemonSet() error {
	return nil
}

func (d *DaemonSetServiceImpl) DeleteDaemonSet() error {
	return nil
}
