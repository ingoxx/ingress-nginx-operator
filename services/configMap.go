package services

import (
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ConfigMapServiceImpl 实现 ConfigMapService 接口
type ConfigMapServiceImpl struct {
	ing common.Generic
	ctx context.Context
}

// NewConfigMapServiceImpl 创建 Service 实例
func NewConfigMapServiceImpl(ctx context.Context, clientSet common.Generic) *ConfigMapServiceImpl {
	return &ConfigMapServiceImpl{ctx: ctx, ing: clientSet}
}

func (c *ConfigMapServiceImpl) GetConfigMapData(name string) ([]byte, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.ing.GetNameSpace()}
	if err := c.ing.GetClient().Get(context.Background(), req, cm); err != nil {
		return nil, cerr.NewKubernetesResourcesNotFoundError("ConfigMap", "name", c.ing.GetNameSpace())
	}

	data, ok := cm.Data[constants.NginxFullChain]
	if !ok || data == "" {
		return nil, cerr.NewKubernetesResourcesNotFoundError("ConfigMap", "configMap.Data", c.ing.GetNameSpace())
	}

	return []byte(data), nil
}

func (c *ConfigMapServiceImpl) CreateConfigMap(name string, data map[string]string) (map[string]string, error) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.ing.GetNameSpace(),
		},
		Data: data,
	}

	if err := c.ing.GetClient().Create(context.Background(), cm); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *ConfigMapServiceImpl) UpdateConfigMap(name string, data map[string]string) (map[string]string, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.ing.GetNameSpace()}
	if err := c.ing.GetClient().Get(context.Background(), req, cm); err == nil {
		cm.Data = data
		if err := c.ing.GetClient().Update(context.Background(), cm); err != nil {
			return nil, err
		}
	}

	return data, nil
}
