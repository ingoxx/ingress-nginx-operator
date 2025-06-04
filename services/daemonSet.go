package services

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"golang.org/x/net/context"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v14 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
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

func (ds *DaemonSetServiceImpl) daemonSetLabels() map[string]string {
	return map[string]string{"app": constants.DaemonSetLabel}
}

func (ds *DaemonSetServiceImpl) GetDaemonSet(data *buildDaemonSetData) (*v1.DaemonSet, error) {
	var dp = new(v1.DaemonSet)
	if err := ds.generic.GetClient().Get(ds.ctx, ds.GetDaemonSetKey(data.key), dp); err != nil {
		return dp, err
	}

	return dp, nil
}

func (ds *DaemonSetServiceImpl) UpdateDaemonSet(daemonSet *v1.DaemonSet, data *buildDaemonSetData) error {
	daemonSet.Spec.Template.Spec.Containers = ds.daemonSetPodContainer(data)
	return ds.generic.GetClient().Update(ds.ctx, daemonSet)
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
		ObjectMeta: ds.daemonSetMeta(data),
		Spec:       ds.daemonSetSpec(data),
	}

	return dp
}

func (ds *DaemonSetServiceImpl) daemonSetMeta(data *buildDaemonSetData) v12.ObjectMeta {
	om := v12.ObjectMeta{
		Name:      constants.DaemonSetName,
		Namespace: data.key.Namespace,
		Labels:    ds.daemonSetLabels(),
	}

	return om
}

func (ds *DaemonSetServiceImpl) daemonSetSpec(data *buildDaemonSetData) v1.DaemonSetSpec {
	var revisionHistoryLimit = new(int32)
	*revisionHistoryLimit = 10

	dss := v1.DaemonSetSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: ds.daemonSetLabels(),
		},
		Template: ds.daemonSetTemplate(data),
		UpdateStrategy: v1.DaemonSetUpdateStrategy{
			Type: v1.RollingUpdateDaemonSetStrategyType,
		},
		RevisionHistoryLimit: revisionHistoryLimit,
	}

	return dss
}

func (ds *DaemonSetServiceImpl) daemonSetTemplate(data *buildDaemonSetData) v13.PodTemplateSpec {
	dc := v13.PodTemplateSpec{
		ObjectMeta: v12.ObjectMeta{
			Labels: ds.daemonSetLabels(),
		},
		Spec: v13.PodSpec{
			Containers:                    ds.daemonSetPodContainer(data),
			TerminationGracePeriodSeconds: pointer.Int64(30),
			DNSPolicy:                     v13.DNSClusterFirst,
			RestartPolicy:                 v13.RestartPolicyAlways,
			Affinity:                      ds.nodeAffinity(),
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

	readinessProbe := &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			HTTPGet: &v13.HTTPGetAction{
				Path: constants.HealthUrl,
				Port: intstr.FromInt(constants.HealthPort),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
	}

	livenessProbe := &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			HTTPGet: &v13.HTTPGetAction{
				Path: constants.HealthUrl,
				Port: intstr.FromInt(constants.HealthPort),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       30,
	}

	c := v13.Container{
		Command: constants.Command,
		Name:    constants.DeployName,
		Image:   constants.Images,
		Ports:   cps,
		Resources: v13.ResourceRequirements{
			Requests: v13.ResourceList{
				v13.ResourceCPU:    resource.MustParse("100m"),
				v13.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: v13.ResourceList{
				v13.ResourceCPU:    resource.MustParse("500m"),
				v13.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		ReadinessProbe: readinessProbe,
		LivenessProbe:  livenessProbe,
	}

	cs = append(cs, c)

	return cs
}

func (ds *DaemonSetServiceImpl) nodeAffinity() *v13.Affinity {
	return &v13.Affinity{
		NodeAffinity: &v13.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v13.NodeSelector{
				NodeSelectorTerms: []v13.NodeSelectorTerm{
					{
						MatchExpressions: []v13.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/arch",
								Operator: v13.NodeSelectorOpIn,
								Values: []string{
									"amd64", "arm64", "ppc64le", "s390x",
								},
							},
							{
								Key:      "kubernetes.io/os",
								Operator: v13.NodeSelectorOpIn,
								Values:   []string{"linux"},
							},
						},
					},
				},
			},
		},
	}
}

func (ds *DaemonSetServiceImpl) CheckDaemonSet() error {
	if ds.config.EnableStream.EnableStream {
		streamData := ds.config.EnableStream.StreamBackendList
		for _, v := range streamData {
			key := types.NamespacedName{Name: v.Name, Namespace: v.Namespace}
			_, err := ds.generic.GetService(key)
			if err != nil {
				return err
			}

			ports, err := ds.generic.GetBackendPorts(key)
			if err != nil {
				return err
			}

			data := &buildDaemonSetData{
				sbp: ports,
				key: key,
			}

			daemonSet, err := ds.GetDaemonSet(data)
			if err != nil {
				if errors.IsNotFound(err) {
					if err := ds.CreateDaemonSet(data); err != nil {
						return err
					}
					continue
				}

				return err
			}

			if err := ds.UpdateDaemonSet(daemonSet, data); err != nil {
				return err
			}
		}

	}

	return nil
}
