package services

import (
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/ingress"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// IngressServiceImpl 实现 IngressService 接口
type IngressServiceImpl struct {
	k8sCli      common.K8sClientSet
	operatorCli common.OperatorClientSet
	ctx         context.Context
	ingress     *v1.Ingress
}

// NewIngressServiceImpl 创建 Service 实例
func NewIngressServiceImpl(ctx context.Context, k8sCli common.K8sClientSet, operatorCli common.OperatorClientSet) *IngressServiceImpl {
	return &IngressServiceImpl{ctx: ctx, k8sCli: k8sCli, operatorCli: operatorCli}
}

func (i *IngressServiceImpl) GetIngress(ctx context.Context, req client.ObjectKey) (*v1.Ingress, error) {
	var ing = new(v1.Ingress)
	if err := i.operatorCli.GetClient().Get(ctx, req, ing); err != nil {
		return ing, cerr.NewIngressNotFoundError(fmt.Sprintf("ingress '%s' not found in namespace '%s'", req.Name, req.Namespace))
	}

	i.ingress = ing
	i.ctx = ctx

	return ing, nil
}

func (i *IngressServiceImpl) GetName() string {
	return i.ingress.Name
}

func (i *IngressServiceImpl) GetNameSpace() string {
	return i.ingress.Namespace
}

func (i *IngressServiceImpl) GetRules() []v1.IngressRule {
	var rs = make([]v1.IngressRule, len(i.ingress.Spec.Rules))

	for k, s := range i.ingress.Spec.Rules {
		rs = append(rs[:k], s)
	}

	return rs
}

func (i *IngressServiceImpl) GetHosts() []string {
	rs := i.GetRules()
	var hosts = make([]string, len(rs))

	for k, r := range rs {
		hosts = append(hosts[:k], r.Host)
	}

	return hosts
}

func (i *IngressServiceImpl) GetPaths() []string {
	rs := i.GetRules()
	var paths = make([]string, 0, 6)

	for _, r := range rs {
		for _, p := range r.HTTP.Paths {
			paths = append(paths, p.Path)
		}

	}

	return paths
}

func (i *IngressServiceImpl) GetPathType(name string) (string, error) {
	rs := i.GetRules()
	var pt string

	for _, r := range rs {
		for _, p := range r.HTTP.Paths {
			if _, err := i.GetBackend(name); err != nil {
				return pt, err
			}
			pt = string(*p.PathType)
		}
	}

	if pt == "" {
		return pt, cerr.NewMissIngressFieldValueError("pathType", i.GetName(), i.GetNameSpace())
	}

	return pt, nil
}

func (i *IngressServiceImpl) GetTlsHosts() []string {
	rs := i.GetTls()
	var hosts = make([]string, len(rs))

	for _, r := range rs {
		for k2, r2 := range r.Hosts {
			hosts = append(hosts[:k2], r2)
		}
	}

	return hosts
}

func (i *IngressServiceImpl) CheckTlsHosts() bool {
	tlsHost := i.GetHosts()
	ingHost := i.GetTlsHosts()

	set := make(map[string]struct{}, len(tlsHost))
	for _, item := range tlsHost {
		set[item] = struct{}{}
	}

	for _, item := range ingHost {
		if _, ok := set[item]; !ok {
			return false
		}
	}

	return true
}

func (i *IngressServiceImpl) GetBackend(name string) (*v1.ServiceBackendPort, error) {
	var bk = new(v1.ServiceBackendPort)
	var rs = i.GetRules()
	key := types.NamespacedName{Name: name, Namespace: i.GetNameSpace()}
	svc, err := i.GetService(key)
	if err != nil {
		return bk, err
	}

	for _, r := range rs {
		for _, p := range r.HTTP.Paths {
			if p.Backend.Service.Name == svc.Name {
				port := i.GetBackendPort(svc)
				if port == 0 {
					return bk, cerr.NewInvalidSvcPortError(svc.Name, i.GetName(), i.GetNameSpace())
				}

				bk.Number = port
				bk.Name = svc.Name

				return bk, nil
			}
		}
	}

	return bk, nil
}

func (i *IngressServiceImpl) GetAnnotations() map[string]string {
	return i.ingress.GetAnnotations()
}

func (i *IngressServiceImpl) GetIngressObjectMate() metav1.ObjectMeta {
	return i.ingress.ObjectMeta
}

func (i *IngressServiceImpl) GetDefaultBackend() (*v1.ServiceBackendPort, error) {
	var bk = new(v1.ServiceBackendPort)
	if i.ingress.Spec.DefaultBackend != nil {
		key := types.NamespacedName{Name: i.ingress.Spec.DefaultBackend.Service.Name, Namespace: i.GetNameSpace()}
		svc, err := i.GetService(key)
		if err != nil {
			return bk, err
		}

		port := i.GetDefaultBackendPort(svc)
		if port == 0 {
			return bk, cerr.NewInvalidSvcPortError(svc.Name, i.GetName(), i.GetNameSpace())
		}

		bk.Number = port
		bk.Name = svc.Name

	}

	return bk, nil
}

func (i *IngressServiceImpl) GetService(key client.ObjectKey) (*corev1.Service, error) {
	var svc = new(corev1.Service)
	//key := types.NamespacedName{Name: name, Namespace: i.GetNameSpace()}
	if err := i.operatorCli.GetClient().Get(i.ctx, key, svc); err != nil {
		return svc, err
	}

	return svc, nil
}

func (i *IngressServiceImpl) GetBackendPort(svc *corev1.Service) int32 {
	var port int32
	for _, r := range i.ingress.Spec.Rules {
		for _, p := range r.HTTP.Paths {
			if p.Backend.Service.Name == svc.Name {
				for _, sp := range i.GetSvcPort(svc) {
					if p.Backend.Service.Port.Number == sp {
						return sp
					}
				}
			}
		}
	}

	return port
}

func (i *IngressServiceImpl) GetBackendPorts(key client.ObjectKey) ([]*v1.ServiceBackendPort, error) {
	var ports = make([]*v1.ServiceBackendPort, 0, 5)
	service, err := i.GetService(key)
	if err != nil {
		return ports, err
	}

	for _, p := range service.Spec.Ports {
		bk := &v1.ServiceBackendPort{
			Name:   key.Name,
			Number: p.Port,
		}
		ports = append(ports, bk)
	}

	return ports, nil
}

func (i *IngressServiceImpl) GetDefaultBackendPort(svc *corev1.Service) int32 {
	var port int32
	for _, p := range i.GetSvcPort(svc) {
		if i.ingress.Spec.DefaultBackend.Service.Port.Number == p {
			return p
		}
	}

	return port
}

func (i *IngressServiceImpl) GetSvcPort(svc *corev1.Service) []int32 {
	var ports = make([]int32, 2)
	for _, p := range svc.Spec.Ports {
		ports = append(ports, p.Port)
	}

	return ports
}

func (i *IngressServiceImpl) GetUpstreamConfig() ([]*ingress.Backends, error) {
	var rs = i.GetRules()
	var upStreamConfigList = make([]*ingress.Backends, 0, len(rs))

	for _, r := range rs {
		var backends = make([]*ingress.IngBackends, 0, len(r.HTTP.Paths))

		for _, p := range r.HTTP.Paths {
			backend, err := i.GetBackend(p.Backend.Service.Name)
			if err != nil {
				return nil, err
			}
			// 使用正则表达式时pathType字段必须为：ImplementationSpecific
			imp := v1.PathTypeImplementationSpecific
			if (parser.IsRegex(p.Path) && *p.PathType != imp) || (*p.PathType == imp && !parser.IsRegex(p.Path)) {
				return upStreamConfigList, cerr.NewSetPathTypeError(i.GetName(), i.GetNameSpace())
			}

			bk := &ingress.IngBackends{
				Services:      backend,
				Path:          p.Path,
				PathType:      string(*p.PathType),
				SvcName:       backend.Name,
				IsPathIsRegex: parser.IsRegex(p.Path),
			}
			backends = append(backends, bk)
		}

		uc := &ingress.Backends{
			Host:           r.Host,
			Upstream:       i.getUpstreamName(r.Host),
			ServiceBackend: backends,
		}
		upStreamConfigList = append(upStreamConfigList, uc)
	}

	return upStreamConfigList, nil
}

func (i *IngressServiceImpl) getUpstreamName(data string) string {
	return fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(data, ".", "_"), i.GetName(), i.GetNameSpace())
}

func (i *IngressServiceImpl) GetBackendName(name *v1.ServiceBackendPort) string {
	return fmt.Sprintf("%s.%s.svc:%d", name.Name, i.GetNameSpace(), name.Number)
}

func (i *IngressServiceImpl) GetAnyBackendName(name *v1.ServiceBackendPort, ns string) string {
	return fmt.Sprintf("%s.%s.svc:%d", name.Name, ns, name.Number)
}

func (i *IngressServiceImpl) GetClientSet() *kubernetes.Clientset {
	return i.k8sCli.GetClientSet()
}

func (i *IngressServiceImpl) GetDynamicClientSet() dynamic.Interface {
	return i.k8sCli.GetDynamicClientSet()
}

func (i *IngressServiceImpl) GetClient() client.Client {
	return i.operatorCli.GetClient()
}

func (i *IngressServiceImpl) GetTls() []v1.IngressTLS {
	return i.ingress.Spec.TLS
}

func (i *IngressServiceImpl) GetDaemonSetNameLabel() string {
	return "daemonset-manager"
}

func (i *IngressServiceImpl) GetDeployNameLabel() string {
	return "deploy-manager"
}

func (i *IngressServiceImpl) GetDaemonSetLabel() string {
	return "daemonset-manager-app"
}

func (i *IngressServiceImpl) GetDeployLabel() string {
	return "deploy-manager-app"
}

func (i *IngressServiceImpl) GetDaemonSvcName() string {
	return "daemonset-manager-svc"
}

func (i *IngressServiceImpl) GetDeploySvcName() string {
	return "deploy-manager-svc"
}

func (i *IngressServiceImpl) CheckService() error {
	var err error
	if err1, err2 := i.checkDefaultBackend(), i.checkBackend(); err1 != nil && err2 != nil {
		err = errors.Join(err1, err2)
		return err
	}

	return nil
}

func (i *IngressServiceImpl) CheckController() error {
	ic := new(v1.IngressClass)

	if i.ingress.Spec.IngressClassName == nil && i.ingress.GetAnnotations()[constants.IngAnnotationKey] == "" {
		return fmt.Errorf("select available ingress nginx controller")
	}

	if i.ingress.Annotations[constants.IngAnnotationKey] == constants.IngAnnotationVal {
		return nil
	}

	if err := i.operatorCli.GetClient().Get(i.ctx, types.NamespacedName{Name: *i.ingress.Spec.IngressClassName}, ic); err != nil {
		return err
	}

	if ic.Spec.Controller != constants.IngController {
		return fmt.Errorf("pls select available ingress nginx controller")
	}

	return nil
}

func (i *IngressServiceImpl) CheckHost() error {
	var recordExistsHost string
	var rs = i.GetRules()

	for _, r := range rs {
		if r.Host == "" {
			return cerr.NewMissIngressFieldValueError("host", i.GetName(), i.GetNameSpace())
		}

		if recordExistsHost == "" {
			recordExistsHost = r.Host
		} else if recordExistsHost == r.Host {
			return cerr.NewDuplicateHostError(i.GetName(), i.GetNameSpace())
		}
	}

	return nil
}

func (i *IngressServiceImpl) CheckPath(path []v1.HTTPIngressPath) error {
	pattern := `^/`
	var recordExistsPath string
	for _, p := range path {
		if recordExistsPath == "" {
			recordExistsPath = p.Path
		} else if recordExistsPath == p.Path {
			return cerr.NewDuplicatePathError(i.GetName(), i.GetNameSpace())
		}

		matched, err := regexp.MatchString(pattern, p.Path)
		if err != nil {
			return err
		}

		if !matched {
			return cerr.NewInvalidIngressPathError(p.Path, i.GetName(), i.GetNameSpace())
		}

		if err := i.CheckPathType(p); err != nil {
			return err
		}

		if p.Backend.Service == nil {
			return cerr.NewMissIngressFieldValueError("Service", i.GetName(), i.GetNameSpace())
		}

		key := types.NamespacedName{Name: p.Backend.Service.Name, Namespace: i.GetNameSpace()}
		svc, err := i.GetService(key)
		if err != nil {
			return err
		}

		if port := i.GetBackendPort(svc); port == 0 {
			return cerr.NewInvalidSvcPortError(svc.Name, i.GetName(), i.GetNameSpace())
		}
	}

	return nil
}

func (i *IngressServiceImpl) CheckPathType(path v1.HTTPIngressPath) error {
	if path.PathType == nil {
		return cerr.NewMissIngressFieldValueError("PathType", i.GetName(), i.GetNameSpace())
	}

	switch *path.PathType {
	case v1.PathTypePrefix, v1.PathTypeExact, v1.PathTypeImplementationSpecific:

	default:
		return cerr.NewMissIngressFieldValueError("PathType", i.GetName(), i.GetNameSpace())
	}

	return nil
}

func (i *IngressServiceImpl) checkDefaultBackend() error {
	if i.ingress.Spec.DefaultBackend == nil {
		return cerr.NewMissIngressFieldValueError("defaultBackend", i.GetName(), i.GetNameSpace())
	}

	if i.ingress.Spec.DefaultBackend != nil {
		key := types.NamespacedName{Namespace: i.GetNameSpace(), Name: i.ingress.Spec.DefaultBackend.Service.Name}
		svc, err := i.GetService(key)
		if err != nil {
			return err
		}

		if port := i.GetDefaultBackendPort(svc); port == 0 {
			return cerr.NewInvalidSvcPortError(svc.Name, i.GetName(), i.GetNameSpace())
		}

	}
	return nil
}

func (i *IngressServiceImpl) checkBackend() error {
	rules := i.GetRules()
	if len(rules) == 0 {
		return cerr.NewMissIngressFieldValueError("rules", i.GetName(), i.GetNameSpace())
	}

	for _, r := range rules {
		if err := i.CheckHost(); err != nil {
			return err
		}

		if r.HTTP == nil {
			return cerr.NewMissIngressFieldValueError("HTTP", i.GetName(), i.GetNameSpace())
		}

		if err := i.CheckPath(r.HTTP.Paths); err != nil {
			return err
		}
	}

	return nil
}
