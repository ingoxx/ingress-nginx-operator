package services

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"golang.org/x/net/context"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type DeploymentServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	config  *annotations.IngressAnnotationsConfig
}

func NewDeploymentServiceImpl(ctx context.Context, clientSet common.Generic) *DeploymentServiceImpl {
	return &DeploymentServiceImpl{ctx: ctx, generic: clientSet}
}

func (d *DeploymentServiceImpl) deployName() string {
	return d.generic.GetName() + "-" + d.generic.GetNameSpace() + "-" + "deploy"
}

func (d *DeploymentServiceImpl) deployLabelsName() string {
	return d.generic.GetName() + "-" + d.generic.GetNameSpace() + "-" + "nginx"
}

func (d *DeploymentServiceImpl) GetDeployKey() types.NamespacedName {
	return types.NamespacedName{Name: d.deployName(), Namespace: d.generic.GetNameSpace()}
}

func (d *DeploymentServiceImpl) GetDeploy() (*v1.Deployment, error) {
	var dp = new(v1.Deployment)
	if err := d.generic.GetClient().Get(d.ctx, d.GetDeployKey(), dp); err != nil {
		return dp, err
	}

	return dp, nil
}

func (d *DeploymentServiceImpl) CreateDeploy() error {
	dp := d.buildDeployData()
	d.generic.GetClient().Create(d.ctx, dp)

	return nil
}

func (d *DeploymentServiceImpl) UpdateDeploy() error {
	return nil
}

func (d *DeploymentServiceImpl) DeleteDeploy() error {
	return nil
}

func (d *DeploymentServiceImpl) buildDeployData() *v1.Deployment {
	var dp = &v1.Deployment{}

	return dp
}

func (d *DeploymentServiceImpl) deployMeta() v12.ObjectMeta {
	labels := map[string]string{"app": d.deployLabelsName()}

	om := v12.ObjectMeta{
		Name:      d.deployLabelsName(),
		Namespace: d.generic.GetNameSpace(),
		Labels:    labels,
	}

	return om
}

func (d *DeploymentServiceImpl) deploySpec() *v1.DeploymentSpec {
	labels := map[string]string{"app": d.deployLabelsName()}

	ds := &v1.DeploymentSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: labels,
		},
	}

	return ds
}

func (d *DeploymentServiceImpl) deployPodTemplate() *v13.PodTemplateSpec {
	labels := map[string]string{"app": d.deployLabelsName()}

	ds := &v13.PodTemplateSpec{
		ObjectMeta: v12.ObjectMeta{
			Labels: labels,
		},
		Spec: v13.PodSpec{
			Containers: []v13.Container{},
		},
	}

	return ds
}

func (d *DeploymentServiceImpl) deployPodContainer() *v13.Container {

	c := &v13.Container{
		Name:  d.deployLabelsName(),
		Image: constants.NginxImages,
		Ports: []v13.ContainerPort{
			{ContainerPort: 80},
		},
	}

	return c
}
