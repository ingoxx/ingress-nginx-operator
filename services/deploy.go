package services

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"golang.org/x/net/context"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

type DeploymentServiceImpl struct {
	ctx     context.Context
	generic common.Generic
	config  *annotations.IngressAnnotationsConfig
}

func NewDeploymentServiceImpl(ctx context.Context, clientSet common.Generic, config *annotations.IngressAnnotationsConfig) *DeploymentServiceImpl {
	return &DeploymentServiceImpl{ctx: ctx, generic: clientSet, config: config}
}

func (d *DeploymentServiceImpl) GetDeployKey() types.NamespacedName {
	return types.NamespacedName{Name: d.generic.GetDeployNameLabel(), Namespace: d.generic.GetNameSpace()}
}

func (d *DeploymentServiceImpl) deployLabels() map[string]string {
	return map[string]string{"app": d.generic.GetDeployLabel()}
}

func (d *DeploymentServiceImpl) GetDeploy() (*v1.Deployment, error) {
	var dp = new(v1.Deployment)
	if err := d.generic.GetClient().Get(d.ctx, d.GetDeployKey(), dp); err != nil {
		return dp, err
	}

	return dp, nil
}

func (d *DeploymentServiceImpl) UpdateDeploy(deploy *v1.Deployment) error {
	deploy.Spec.Template.Spec.Containers = d.deployPodContainer()
	if err := d.generic.GetClient().Update(d.ctx, deploy); err != nil {
		return err
	}

	return nil
}

func (d *DeploymentServiceImpl) DeleteDeploy() error {
	deploy, err := d.GetDeploy()
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	if err := d.generic.GetClient().Delete(d.ctx, deploy); err != nil {
		return err
	}

	return nil
}

func (d *DeploymentServiceImpl) CreateDeploy() error {
	if err := d.generic.GetClient().Create(d.ctx, d.buildDeployData()); err != nil {
		return err
	}
	return nil
}

func (d *DeploymentServiceImpl) buildDeployData() *v1.Deployment {
	var dp = &v1.Deployment{
		ObjectMeta: d.deployMeta(),
		Spec:       d.deploySpec(),
	}

	return dp
}

func (d *DeploymentServiceImpl) deployMeta() v12.ObjectMeta {
	om := v12.ObjectMeta{
		Name:      d.generic.GetDeployNameLabel(),
		Namespace: d.generic.GetNameSpace(),
		Labels:    d.deployLabels(),
	}

	return om
}

func (d *DeploymentServiceImpl) deploySpec() v1.DeploymentSpec {
	var replicas = new(int32)
	var revisionHistoryLimit = new(int32)
	*replicas = 2
	*revisionHistoryLimit = 10

	ds := v1.DeploymentSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: d.deployLabels(),
		},
		Replicas:             replicas,
		Strategy:             d.deployStrategy(),
		Template:             d.deployPodTemplate(),
		RevisionHistoryLimit: revisionHistoryLimit,
	}

	return ds
}

func (d *DeploymentServiceImpl) deployPodTemplate() v13.PodTemplateSpec {
	dc := v13.PodTemplateSpec{
		ObjectMeta: v12.ObjectMeta{
			Labels: d.deployLabels(),
		},
		Spec: v13.PodSpec{
			Containers:                    d.deployPodContainer(),
			TerminationGracePeriodSeconds: pointer.Int64(30),
			DNSPolicy:                     v13.DNSClusterFirst,
			RestartPolicy:                 v13.RestartPolicyAlways,
			Affinity:                      d.nodeAffinity(),
		},
	}

	return dc
}

func (d *DeploymentServiceImpl) nodeAffinity() *v13.Affinity {
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

func (d *DeploymentServiceImpl) deployPodContainer() []v13.Container {
	cs := make([]v13.Container, 0, 3)
	cps := make([]v13.ContainerPort, 0, 10)

	bks, err := d.generic.GetUpstreamConfig()
	if err != nil {
		return cs
	}
	for _, b := range bks {
		for _, b2 := range b.ServiceBackend {
			cp := v13.ContainerPort{
				ContainerPort: b2.Services.Number,
			}
			cps = append(cps, cp)
		}
	}

	cp := v13.ContainerPort{
		ContainerPort: int32(constants.HealthPort),
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
		Name:    d.generic.GetDeployNameLabel(),
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

func (d *DeploymentServiceImpl) deployStrategy() v1.DeploymentStrategy {
	var strategy = v1.DeploymentStrategy{
		Type: v1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &v1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 0,
			},
			MaxSurge: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 1,
			},
		},
	}

	return strategy
}

func (d *DeploymentServiceImpl) CheckDeploy() error {
	deploy, err := d.GetDeploy()
	if err != nil && errors.IsNotFound(err) {
		if err := d.CreateDeploy(); err != nil {
			return err
		}

		return err
	}

	if err := d.UpdateDeploy(deploy); err != nil {
		return err
	}

	return nil
}
