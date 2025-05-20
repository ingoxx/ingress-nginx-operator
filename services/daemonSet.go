package services

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"golang.org/x/net/context"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type DaemonSetServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	config  *annotations.IngressAnnotationsConfig
}

func NewDaemonSetServiceImpl(ctx context.Context, clientSet common.Generic, config *annotations.IngressAnnotationsConfig) *DaemonSetServiceImpl {
	return &DaemonSetServiceImpl{ctx: ctx, generic: clientSet, config: config}
}

func (ds *DaemonSetServiceImpl) daemonSetName() string {
	return ds.generic.GetName() + "-" + ds.generic.GetNameSpace() + "-" + "deploySet"
}

func (ds *DaemonSetServiceImpl) daemonSetLabelsName() string {
	return ds.generic.GetName() + "-" + ds.generic.GetNameSpace() + "-" + "nginx-stream"
}

func (ds *DaemonSetServiceImpl) GetDaemonSetKey() types.NamespacedName {
	return types.NamespacedName{Name: ds.daemonSetName(), Namespace: ds.generic.GetNameSpace()}
}

func (ds *DaemonSetServiceImpl) GetDaemonSet() (*v1.DaemonSet, error) {
	var dp = new(v1.DaemonSet)

	if err := ds.generic.GetClient().Get(ds.ctx, ds.GetDaemonSetKey(), dp); err != nil {
		return dp, err
	}

	return dp, nil
}

func (ds *DaemonSetServiceImpl) CreateDaemonSet() error {
	return nil
}

func (ds *DaemonSetServiceImpl) UpdateDaemonSet() error {
	return nil
}

func (ds *DaemonSetServiceImpl) DeleteDaemonSet() error {
	return nil
}

func (ds *DaemonSetServiceImpl) buildDaemonSet() *v1.DaemonSet {
	var dp = &v1.DaemonSet{
		ObjectMeta: ds.daemonSetMeta(),
		Spec:       ds.daemonSetSpec(),
	}

	return dp
}

func (ds *DaemonSetServiceImpl) daemonSetMeta() v12.ObjectMeta {
	labels := map[string]string{"app": ds.daemonSetLabelsName()}
	om := v12.ObjectMeta{
		Name:      ds.daemonSetName(),
		Namespace: ds.generic.GetNameSpace(),
		Labels:    labels,
	}

	return om
}

func (ds *DaemonSetServiceImpl) daemonSetSpec() v1.DaemonSetSpec {
	labels := map[string]string{"app": ds.daemonSetLabelsName()}
	dss := v1.DaemonSetSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: labels,
		},
		Template: ds.daemonSetTemplate(),
		UpdateStrategy: v1.DaemonSetUpdateStrategy{
			Type: v1.RollingUpdateDaemonSetStrategyType,
		},
	}

	return dss
}

func (ds *DaemonSetServiceImpl) daemonSetTemplate() v13.PodTemplateSpec {
	labels := map[string]string{"app": ds.daemonSetLabelsName()}
	dc := v13.PodTemplateSpec{
		ObjectMeta: v12.ObjectMeta{
			Labels: labels,
		},
		Spec: v13.PodSpec{
			Containers: ds.daemonSetPodContainer(),
		},
	}

	return dc
}

func (ds *DaemonSetServiceImpl) daemonSetPodContainer() []v13.Container {
	cs := make([]v13.Container, 0, 3)
	//cps := make([]v13.ContainerPort, 0, 10)
	//streamData := ds.config.EnableStream.StreamBackendList
	//
	//for _, v := range streamData {
	//	cp := v13.ContainerPort{
	//		ContainerPort: ,
	//	}
	//	cps = append(cps, cp)
	//}

	return cs
}

func (ds *DaemonSetServiceImpl) CheckDaemonSet() error {
	if ds.config.EnableStream.EnableStream {

	}

	return nil
}
