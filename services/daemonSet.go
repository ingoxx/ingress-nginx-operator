package services

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"golang.org/x/net/context"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v14 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DaemonSetServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	config  *annotations.IngressAnnotationsConfig
}

type buildDaemonSetData struct {
	sbp []*v14.ServiceBackendPort
	key client.ObjectKey
}

func NewDaemonSetServiceImpl(ctx context.Context, clientSet common.Generic, config *annotations.IngressAnnotationsConfig) *DaemonSetServiceImpl {
	return &DaemonSetServiceImpl{ctx: ctx, generic: clientSet, config: config}
}

func (ds *DaemonSetServiceImpl) GetDaemonSetKey(key client.ObjectKey) types.NamespacedName {
	return types.NamespacedName{Name: ds.generic.GetDaemonSetNameLabel(), Namespace: key.Namespace}
}

func (ds *DaemonSetServiceImpl) GetDaemonSet(data *buildDaemonSetData) (*v1.DaemonSet, error) {
	var dp = new(v1.DaemonSet)
	if err := ds.generic.GetClient().Get(ds.ctx, ds.GetDaemonSetKey(data.key), dp); err != nil {
		return dp, err
	}

	return dp, nil
}

func (ds *DaemonSetServiceImpl) UpdateDaemonSet(data *v1.DaemonSet) error {
	return ds.generic.GetClient().Update(ds.ctx, data)
}

func (ds *DaemonSetServiceImpl) DeleteDaemonSet(data *v1.DaemonSet) error {
	return ds.generic.GetClient().Delete(ds.ctx, data)
}

func (ds *DaemonSetServiceImpl) CreateDaemonSet(data *buildDaemonSetData) error {
	if err := ds.generic.GetClient().Create(ds.ctx, ds.buildDaemonSet(data)); err != nil {
		return err
	}

	return nil
}

func (ds *DaemonSetServiceImpl) buildDaemonSet(data *buildDaemonSetData) *v1.DaemonSet {
	var dp = &v1.DaemonSet{
		ObjectMeta: ds.daemonSetMeta(),
		Spec:       ds.daemonSetSpec(data),
	}

	return dp
}

func (ds *DaemonSetServiceImpl) daemonSetMeta() v12.ObjectMeta {
	labels := map[string]string{"app": ds.generic.GetDaemonSetNameLabel()}
	om := v12.ObjectMeta{
		Name:      ds.generic.GetDaemonSetNameLabel(),
		Namespace: ds.generic.GetNameSpace(),
		Labels:    labels,
	}

	return om
}

func (ds *DaemonSetServiceImpl) daemonSetSpec(data *buildDaemonSetData) v1.DaemonSetSpec {
	labels := map[string]string{"app": ds.generic.GetDaemonSetNameLabel()}
	dss := v1.DaemonSetSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: labels,
		},
		Template: ds.daemonSetTemplate(data),
		UpdateStrategy: v1.DaemonSetUpdateStrategy{
			Type: v1.RollingUpdateDaemonSetStrategyType,
		},
	}

	return dss
}

func (ds *DaemonSetServiceImpl) daemonSetTemplate(data *buildDaemonSetData) v13.PodTemplateSpec {
	labels := map[string]string{"app": ds.generic.GetDaemonSetNameLabel()}
	dc := v13.PodTemplateSpec{
		ObjectMeta: v12.ObjectMeta{
			Labels: labels,
		},
		Spec: v13.PodSpec{
			Containers: ds.daemonSetPodContainer(data),
		},
	}

	return dc
}

func (ds *DaemonSetServiceImpl) daemonSetPodContainer(data *buildDaemonSetData) []v13.Container {
	cs := make([]v13.Container, 0, 3)
	cps := make([]v13.ContainerPort, 0, 10)

	for _, v := range data.sbp {
		cp := v13.ContainerPort{
			ContainerPort: v.Number,
		}
		cps = append(cps, cp)
	}

	cp := v13.ContainerPort{
		ContainerPort: 9092,
	}
	cps = append(cps, cp)

	return cs
}

func (ds *DaemonSetServiceImpl) CheckDaemonSet() error {
	if ds.config.EnableStream.EnableStream {
		var svcPorts = make([]*v14.ServiceBackendPort, 0, 4)
		streamData := ds.config.EnableStream.StreamBackendList
		for _, v := range streamData {
			key := types.NamespacedName{Name: v.Name, Namespace: v.Namespace}
			svc, err := ds.generic.GetService(key)
			if err != nil {
				return err
			}
			for _, vd2 := range svc.Spec.Ports {
				svcPort := &v14.ServiceBackendPort{
					Name:   svc.Name,
					Number: vd2.Port,
				}
				svcPorts = append(svcPorts, svcPort)
			}

			b := &buildDaemonSetData{
				sbp: svcPorts,
				key: key,
			}
			set, err := ds.GetDaemonSet(b)
			if err != nil {
				if errors.IsNotFound(err) {
					if err := ds.CreateDaemonSet(b); err != nil {
						return err
					}
					continue
				}

				return err
			}

			if err := ds.UpdateDaemonSet(set); err != nil {
				return err
			}
		}

	}

	return nil
}
