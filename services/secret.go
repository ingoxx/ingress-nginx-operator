package services

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/file"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	cert    service.K8sResourcesCert
}

func NewSecretServiceImpl(ctx context.Context, clientSet common.Generic, cert service.K8sResourcesCert) *SecretServiceImpl {
	return &SecretServiceImpl{ctx: ctx, generic: clientSet, cert: cert}
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
	if err := s.generic.GetClient().Get(s.ctx, key, sc); err != nil {
		return sc, err
	}

	return sc, nil
}

func (s *SecretServiceImpl) DeleteSecret() error {
	req := types.NamespacedName{Name: s.cert.SecretObjectKey(), Namespace: s.generic.GetNameSpace()}
	secret, err := s.GetSecret(req)
	if err != nil {
		return err
	}
	if err := s.generic.GetClient().Delete(s.ctx, secret); err != nil {
		return err
	}

	return nil
}

func (s *SecretServiceImpl) extractTlsData(data map[string][]byte) (map[string][]byte, error) {
	var parsed = make(map[string][]byte)
	for k, v := range data {
		parsed[k] = v
	}

	return parsed, nil
}

func (s *SecretServiceImpl) GetTlsFile() (map[string]ingress.Tls, error) {
	if len(s.generic.GetTls()) > 0 {
		cs, err := s.caSigned()
		if err != nil {
			return cs, err
		}

		return cs, nil
	}

	ss, err := s.selfSigned()
	if err != nil {
		return ss, err
	}

	return ss, nil
}

func (s *SecretServiceImpl) selfSigned() (map[string]ingress.Tls, error) {
	var ss ingress.Tls
	var ht = make(map[string]ingress.Tls)

	key := types.NamespacedName{Name: s.cert.SecretObjectKey(), Namespace: s.generic.GetNameSpace()}
	data, err := s.GetTlsData(key)
	if err != nil {
		return ht, err
	}

	for _, v := range s.generic.GetHosts() {
		for k, v2 := range data {
			fileName := filepath.Join(constants.NginxSSLDir, fmt.Sprintf("%s-%s", s.cert.SecretObjectKey(), k))
			if err := file.SaveToFile(fileName, v2); err != nil {
				return ht, err
			}

			if k == constants.NginxTlsCrt {
				ss.TlsCrt = fileName
			} else if k == constants.NginxTlsKey {
				ss.TlsKey = fileName
			}
		}
		ht[v] = ss
	}

	return ht, nil
}

func (s *SecretServiceImpl) caSigned() (map[string]ingress.Tls, error) {
	if !s.generic.CheckTlsHosts() {
		return nil, cerr.NewNotFoundTlsHostError(s.generic.GetName(), s.generic.GetNameSpace())
	}

	var ss ingress.Tls
	var ht = make(map[string]ingress.Tls)

	for _, v := range s.generic.GetTls() {
		key := types.NamespacedName{Name: v.SecretName, Namespace: s.generic.GetNameSpace()}
		_, err := s.GetSecret(key)
		if err != nil {
			return nil, err
		}

		data, err := s.GetTlsData(key)
		if err != nil {
			return nil, err
		}
		for _, h := range v.Hosts {
			for k, v2 := range data {
				fileName := filepath.Join(constants.NginxSSLDir, s.cert.SecretObjectKey())
				if err := file.SaveToFile(fileName, v2); err != nil {
					return ht, err
				}
				//if err := os.WriteFile(fileName, v2, 0644); err != nil {
				//	return nil, err
				//}

				if k == constants.NginxTlsCrt {
					ss.TlsCrt = fileName
				} else if k == constants.NginxTlsKey {
					ss.TlsKey = fileName
				}
			}
			ht[h] = ss
		}
	}

	return ht, nil
}
