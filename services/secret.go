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
	var data map[string][]byte

	secret, err := s.GetSecret(key)
	if err != nil {
		return data, err
	}

	data, err = s.extractTlsData(secret.Data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (s *SecretServiceImpl) GetSecret(key client.ObjectKey) (*corev1.Secret, error) {
	sc := new(corev1.Secret)
	if err := s.clientSet.GetClient().Get(s.ctx, key, sc); err != nil {
		return sc, err
	}

	return sc, nil
}

func (s *SecretServiceImpl) extractTlsData(data map[string][]byte) (map[string][]byte, error) {
	var parsed = make(map[string][]byte)
	for k, v := range data {
		parsed[k] = v
	}

	return parsed, nil
}
