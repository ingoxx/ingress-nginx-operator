package services

import (
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (i *IngressServiceImpl) GetHosts(ctx context.Context, namespace, name string) []string {
	var hosts = make([]string, 2)
	if len(i.ingress.Spec.Rules) > 0 {
		for _, r := range i.ingress.Spec.Rules {
			hosts = append(hosts, r.Host)
		}
	}

	return hosts
}

func (i *IngressServiceImpl) GetBackend(name string) (*v1.ServiceBackendPort, error) {
	var bk = new(v1.ServiceBackendPort)
	if len(i.ingress.Spec.Rules) > 0 {
		svc, err := i.GetService(name)
		if err != nil {
			return bk, err
		}

		for _, r := range i.ingress.Spec.Rules {
			for _, p := range r.HTTP.Paths {
				if p.Backend.Service.Name == svc.Name {
					port := i.GetBackendPort(svc)
					if port == 0 {
						return bk, cerr.NewInvalidSvcPortError(svc.Name, i.ingress.Name, i.ingress.Namespace)
					}

					bk.Number = port
					bk.Name = svc.Name

					return bk, nil
				}
			}
		}
	}

	return bk, nil
}

func (i *IngressServiceImpl) GetDefaultBackend() (*v1.ServiceBackendPort, error) {
	var bk = new(v1.ServiceBackendPort)
	if i.ingress.Spec.DefaultBackend != nil {
		svc, err := i.GetService(i.ingress.Spec.DefaultBackend.Service.Name)
		if err != nil {
			return bk, err
		}

		port := i.GetDefaultBackendPort(svc)
		if port == 0 {
			return bk, cerr.NewInvalidSvcPortError(svc.Name, i.ingress.Name, i.ingress.Namespace)
		}

		bk.Number = port
		bk.Name = svc.Name

	}

	return bk, nil
}

func (i *IngressServiceImpl) GetService(name string) (*corev1.Service, error) {
	var svc *corev1.Service
	key := types.NamespacedName{Name: name, Namespace: i.ingress.Namespace}
	if err := i.operatorCli.GetClient().Get(i.ctx, key, svc); err != nil {
		return svc, err
	}

	return svc, nil
}

func (i *IngressServiceImpl) GetBackendPort(svc *corev1.Service) int32 {
	var port int32
	if len(i.ingress.Spec.Rules) > 0 {
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
	}

	return port
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

func (i *IngressServiceImpl) GetUpstreamName(paths []v1.HTTPIngressPath, ing interface{}) string {
	return ""
}

func (i *IngressServiceImpl) getUpstreamBackend(paths []v1.HTTPIngressPath) string {
	return ""
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
	if *i.ingress.Spec.IngressClassName == "" && i.ingress.GetAnnotations()[constants.IngAnnotationKey] == "" {
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

	return nil
}

func (i *IngressServiceImpl) CheckPath() error {

	return nil
}

func (i *IngressServiceImpl) checkDefaultBackend() error {
	if i.ingress.Spec.DefaultBackend != nil {
		svc, err := i.GetService(i.ingress.Spec.DefaultBackend.Service.Name)
		if err != nil {
			return err
		}

		if port := i.GetDefaultBackendPort(svc); port == 0 {
			return cerr.NewInvalidSvcPortError(svc.Name, i.ingress.Name, i.ingress.Namespace)
		}

	}
	return nil
}

func (i *IngressServiceImpl) checkBackend() error {
	if len(i.ingress.Spec.Rules) > 0 {
		for _, r := range i.ingress.Spec.Rules {
			for _, p := range r.HTTP.Paths {
				svc, err := i.GetService(p.Backend.Service.Name)
				if err != nil {
					return err
				}
				if port := i.GetBackendPort(svc); port == 0 {
					return cerr.NewInvalidSvcPortError(svc.Name, i.ingress.Name, i.ingress.Namespace)
				}
			}
		}
	}

	return nil
}
