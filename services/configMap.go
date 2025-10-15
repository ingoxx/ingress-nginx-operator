package services

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMapServiceImpl 实现 ConfigMapService 接口
type ConfigMapServiceImpl struct {
	generic common.Generic
	ctx     context.Context
}

// NewConfigMapServiceImpl 创建 Service 实例
func NewConfigMapServiceImpl(ctx context.Context, clientSet common.Generic) *ConfigMapServiceImpl {
	return &ConfigMapServiceImpl{ctx: ctx, generic: clientSet}
}

func (c *ConfigMapServiceImpl) GetConfigMapData(name string) ([]byte, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err != nil {
		return nil, cerr.NewKubernetesResourcesNotFoundError("ConfigMap", "name", c.generic.GetNameSpace())
	}

	data, ok := cm.Data[constants.NginxFullChain]
	if !ok || data == "" {
		return nil, cerr.NewKubernetesResourcesNotFoundError("ConfigMap", "configMap.Data", c.generic.GetNameSpace())
	}

	return []byte(data), nil
}

func (c *ConfigMapServiceImpl) CreateConfigMap(name, key string, data []byte) (map[string]string, error) {
	var cd = map[string]string{
		key: string(data),
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.generic.GetNameSpace(),
		},
		Data: cd,
	}

	if err := c.generic.GetClient().Create(context.Background(), cm); err != nil {
		return nil, err
	}

	return cm.Data, nil
}

func (c *ConfigMapServiceImpl) UpdateConfigMap(name, key string, data []byte) (map[string]string, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err == nil {
		cm.Data[key] = string(data)
		if err := c.generic.GetClient().Update(context.Background(), cm); err != nil {
			return nil, err
		}

		return cm.Data, nil
	}

	cd, err := c.CreateConfigMap(name, key, data)
	if err != nil {
		return nil, err
	}

	return cd, nil
}

func (c *ConfigMapServiceImpl) GetNgxConfigMap(name string) (map[string]string, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err != nil {
		if errors.IsNotFound(err) {
			return cm.Data, nil
		}

		return cm.Data, err
	}

	var cms = new(v1.ConfigMapList)
	if err := c.generic.GetClient().List(context.Background(), cms, client.InNamespace(name)); err != nil {
		return cm.Data, err
	}

	var nb []*stream.Backend
	var lb []*limitreq.ZoneRepConfig

	fmt.Println(nb, lb)

	return cm.Data, nil
}
