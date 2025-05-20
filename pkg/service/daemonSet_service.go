package service

type K8sResourcesDaemonSet interface {
	GetDaemonSet() error
	DeleteDaemonSet() error
	CheckDaemonSet() error
}
