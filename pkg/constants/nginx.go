package constants

const (
	NginxConfDir        = "/etc/nginx/conf.d"
	NginxBin            = "/usr/sbin/nginx"
	NginxSSLDir         = "/etc/nginx/ssl"
	NginxTlsCrt         = "tls.crt"
	NginxTlsKey         = "tls.key"
	NginxTlsCa          = "ca.crt"
	NginxFullChain      = "fullchain.pem"
	NginxPid            = "/var/run/nginx.pid"
	NginxMainConf       = "/etc/nginx/nginx.conf"
	NginxTmpl           = "/rootfs/etc/nginx/template/nginx.tmpl"
	NginxServerTmpl     = "/rootfs/etc/nginx/template/server.tmpl"
	NginxMainServerTmpl = "/rootfs/etc/nginx/template/mainServer.tmpl"
	NginxDefaultTmpl    = "/rootfs/etc/nginx/template/defaultBackend.tmpl"
)
