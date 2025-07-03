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
	HealthUrl      = "/api/v1/health"
	NginxConfUpUrl = "/api/v1/nginx/config/update"
	HealthPort     = 9092
	Command        = []string{"/httpserver"}
	Images         = "gotec007/manager-nginx"
	Replicas       = 1
	AuthToken      = "k8s"
	HttpStatusOk   = 1000
	HttpPorts      = []int32{80, 443, 9092}
	DefaultPort    = 80
)
