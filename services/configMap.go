package services

import (
	"encoding/json"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitconn"
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
	"sync"
)

var cmLocks = sync.Map{}

//var comLock = sync.Mutex{}

// ConfigMapServiceImpl 实现 ConfigMapService 接口
type ConfigMapServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	comLock sync.Mutex
}

// NewConfigMapServiceImpl 创建 Service 实例
func NewConfigMapServiceImpl(ctx context.Context, clientSet common.Generic) *ConfigMapServiceImpl {
	return &ConfigMapServiceImpl{ctx: ctx, generic: clientSet, comLock: sync.Mutex{}}
}

func (c *ConfigMapServiceImpl) getCmLock(cm string) *sync.Mutex {
	svcName := fmt.Sprintf("%s/%s", cm, c.generic.GetNameSpace())
	val, _ := cmLocks.LoadOrStore(svcName, &sync.Mutex{})
	return val.(*sync.Mutex)
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

func (c *ConfigMapServiceImpl) getConfigMap() (*v1.ConfigMap, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: c.GetCmName(), Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err == nil {
		return cm, err
	}

	return cm, nil
}

func (c *ConfigMapServiceImpl) DeleteConfigMap() error {
	cm, err := c.getConfigMap()
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if err := c.generic.GetClient().Delete(context.Background(), cm); err != nil {
		return err
	}

	return nil
}

func (c *ConfigMapServiceImpl) UpdateConfigMap(name, ns, key string, data []byte) (string, error) {
	//c.comLock.Lock()
	//defer c.comLock.Unlock()

	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: name, Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err == nil {
		cm.Data[key] = string(data)
		if err := c.generic.GetClient().Update(context.Background(), cm); err != nil {
			return cm.Data[key], err
		}

		cms, err := c.GetNgxConfigMap(ns)
		if err != nil {
			return cms[key], err
		}

		return cms[key], nil
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

func (c *ConfigMapServiceImpl) GetCm() (*v1.ConfigMap, error) {
	var cm = new(v1.ConfigMap)
	req := types.NamespacedName{Name: c.GetCmName(), Namespace: c.generic.GetNameSpace()}
	if err := c.generic.GetClient().Get(context.Background(), req, cm); err != nil {
		return cm, err
	}

	return cm, nil
}

func (c *ConfigMapServiceImpl) ClearCmData(key string) error {
	cm, err := c.GetCm()
	if err != nil {
		return err
	}

	if cm.Data[key] != "" {
		cm.Data[key] = ""
		if err := c.generic.GetClient().Update(context.Background(), cm); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigMapServiceImpl) GetNgxConfigMap(ns string) (map[string]string, error) {
	var data = make(map[string]string)
	var cms = new(v1.ConfigMapList)
	if err := c.generic.GetClient().List(context.Background(), cms, client.InNamespace(ns)); err != nil {
		return data, err
	}

	var tnb []*stream.Backend
	var tlb []*limitreq.ZoneRepConfig
	var tlc []*limitconn.ZoneConnConfig

	for _, v := range cms.Items {
		var nb []*stream.Backend
		var lb []*limitreq.ZoneRepConfig
		var lc []*limitconn.ZoneConnConfig
		s1 := v.Data[constants.StreamKey]
		s2 := v.Data[constants.LimitReqKey]
		s3 := v.Data[constants.LimitConnKey]

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

		if s3 != "" {
			if err := json.Unmarshal([]byte(s3), &lc); err != nil {
				return v.Data, err
			}
			tlc = append(tlc, lc...)
		}

	}

	if len(tnb) > 0 {
		nsb := c.removeDup(tnb)
		streamBck := nsb.([]*stream.Backend)
		b1, err := json.Marshal(&streamBck)
		if err != nil {
			return data, err
		}

		data[constants.StreamKey] = string(b1)
	}

	if len(tlb) > 0 {
		lzr := c.removeDup(tlb)
		lz := lzr.([]*limitreq.ZoneRepConfig)
		b2, err := json.Marshal(&lz)
		if err != nil {
			return data, err
		}

		data[constants.LimitReqKey] = string(b2)
	}

	if len(tlc) > 0 {
		lzc := c.removeDup(tlc)
		lc := lzc.([]*limitconn.ZoneConnConfig)
		b3, err := json.Marshal(&lc)
		if err != nil {
			return data, err
		}

		data[constants.LimitConnKey] = string(b3)
	}

	return data, nil
}

func (c *ConfigMapServiceImpl) GetCmName() string {
	return fmt.Sprintf("%s-%s-ngx-cm", c.generic.GetName(), c.generic.GetNameSpace())
}

func (c *ConfigMapServiceImpl) GetLatestStreamPorts(name string) ([]*stream.Backend, error) {
	var tnb []*stream.Backend
	configMap, err := c.GetNgxConfigMap(name)
	if err != nil {
		return tnb, err
	}

	data, ok := configMap[constants.StreamKey]
	if !ok {
		return tnb, nil
	}

	if err := json.Unmarshal([]byte(data), &tnb); err != nil {
		return tnb, err
	}

	return tnb, nil
}

func (c *ConfigMapServiceImpl) removeDup(data interface{}) interface{} {
	var dp = make(map[string]struct{})

	switch data.(type) {
	case []*stream.Backend:
		var sb = data.([]*stream.Backend)
		var nd []*stream.Backend
		for _, v := range sb {
			if _, ok := dp[v.StreamBackendName]; !ok {
				nd = append(nd, v)
				dp[v.StreamBackendName] = struct{}{}
			}
		}

		return nd

	case []*limitreq.ZoneRepConfig:
		var lzr = data.([]*limitreq.ZoneRepConfig)
		var nd []*limitreq.ZoneRepConfig
		for _, zone := range lzr {
			for _, zc := range zone.LimitZone {
				if _, ok := dp[zc.ZoneName]; !ok {
					nd = append(nd, zone)
					dp[zc.ZoneName] = struct{}{}
				}
			}
		}

		return nd
	case []*limitconn.ZoneConnConfig:
		var lzc = data.([]*limitconn.ZoneConnConfig)
		var nd []*limitconn.ZoneConnConfig

		for _, zone := range lzc {
			for _, zc := range zone.LimitZone {
				if _, ok := dp[zc.ZoneName]; !ok {
					nd = append(nd, zone)
					dp[zc.ZoneName] = struct{}{}
				}
			}
		}

		return nd
	}

	return nil
}
