package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretServiceImpl struct {
	ingress   service.K8sResourcesIngress
	clientSet common.Generic
	ctx       context.Context
}

func NewSecretServiceImpl(ctx context.Context, clientSet common.Generic) *SecretServiceImpl {
	return &SecretServiceImpl{ctx: ctx, clientSet: clientSet}
}

func (s *SecretServiceImpl) GetTlsData(key client.ObjectKey) (map[string][]byte, error) {
	return nil, nil
}

func (s *SecretServiceImpl) GetSecret() (*corev1.Secret, error) {
	sc := new(corev1.Secret)
	return sc, nil
}
