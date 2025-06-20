package constants

const (
	DeployLabel      = "deploy-manager-app"
	DeployName       = "deploy-manager"
	DeploySvcName    = "deploy-manager-svc"
	DaemonSetLabel   = "daemonset-manager-app"
	DaemonSetName    = "daemonset-manager"
	DaemonSetSvcName = "daemonset-manager-svc"
)

var (
	HealthUrl  = "/api/v1/health"
	HealthPort = 9092
	Command    = []string{"/httpserver"}
	Images     = "gotec007/ingress-nginx"
)
