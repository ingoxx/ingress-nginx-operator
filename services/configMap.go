package services

import (
	"encoding/json"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
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

func (c *ConfigMapServiceImpl) UpdateConfigMap(name, ns, key string, data []byte) (string, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err == nil {
		cm.Data[key] = string(data)
		if err := c.generic.GetClient().Update(context.Background(), cm); err != nil {
			return "", err
		}

		return "", nil
	}

	if _, err := c.CreateConfigMap(name, key, data); err != nil {
		return "", err
	}

	tm, err := c.GetNgxConfigMap(ns)
	if err != nil {
		return "", err
	}

	return tm[key], nil
}

func (c *ConfigMapServiceImpl) GetNgxConfigMap(name string) (map[string]string, error) {
	var data = make(map[string]string)
	var cms = new(v1.ConfigMapList)
	if err := c.generic.GetClient().List(context.Background(), cms, client.InNamespace(name)); err != nil {
		return data, err
	}

	var tnb []*stream.Backend
	var tlb []*limitreq.ZoneRepConfig

	for _, v := range cms.Items {
		var nb []*stream.Backend
		var lb []*limitreq.ZoneRepConfig
		s1 := v.Data[constants.StreamKey]
		s2 := v.Data[constants.LimitReqKey]

		if s1 != "" {
			if err := json.Unmarshal([]byte(s1), &nb); err != nil {
				return v.Data, err
			}
			tnb = append(tnb, nb...)
		}

		if s2 != "" {
			if err := json.Unmarshal([]byte(s2), &lb); err != nil {
				return v.Data, err
			}
			tlb = append(tlb, lb...)
		}

	}

	if len(tnb) > 0 {
		b1, err := json.Marshal(&tnb)
		if err != nil {
			return data, err
		}

		data[constants.StreamKey] = string(b1)
	}

	if len(tlb) > 0 {
		b2, err := json.Marshal(&tlb)
		if err != nil {
			return data, err
		}

		data[constants.LimitReqKey] = string(b2)
	}

	return data, nil
}
