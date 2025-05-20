package service

import v1 "k8s.io/api/apps/v1"

type K8sResourcesDeploy interface {
	GetDeploy() (*v1.Deployment, error)
	CreateDeploy() error
	UpdateDeploy(*v1.Deployment) error
	DeleteDeploy() error
	CheckDeploy() error
}
