package services

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"golang.org/x/net/context"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	svcPort = 9092
)

type buildSvcData struct {
	sbp    []*v1.ServiceBackendPort
	labels map[string]string
	key    client.ObjectKey
}

type SvcServiceImpl struct {
	generic common.Generic
	ctx     context.Context
	config  *annotations.IngressAnnotationsConfig
}

func NewSvcServiceImpl(ctx context.Context, clientSet common.Generic, config *annotations.IngressAnnotationsConfig) *SvcServiceImpl {
	return &SvcServiceImpl{ctx: ctx, generic: clientSet, config: config}
}

func (s *SvcServiceImpl) GetSvc(key client.ObjectKey) (*v13.Service, error) {
	var svc = new(v13.Service)
	if err := s.generic.GetClient().Get(s.ctx, key, svc); err != nil {
		return svc, err
	}

	return svc, nil
}

func (s *SvcServiceImpl) UpdateSvc(svc *v13.Service, data *buildSvcData) error {
	svc.Spec.Ports = s.svcServicePort(data.sbp)
	return s.generic.GetClient().Update(s.ctx, svc)
}

func (s *SvcServiceImpl) DeleteSvc(svc *v13.Service) error {
	return s.generic.GetClient().Delete(s.ctx, svc)
}

func (s *SvcServiceImpl) CreateSvc(data *buildSvcData) error {
	if err := s.generic.GetClient().Create(s.ctx, s.buildSvcData(data)); err != nil {
		return err
	}
	return nil
}

func (s *SvcServiceImpl) buildSvcData(data *buildSvcData) *v13.Service {

	sd := &v13.Service{
		ObjectMeta: s.svcObjectMeta(data),
		Spec:       s.svcServiceSpec(data),
	}

	return sd
}

func (s *SvcServiceImpl) svcObjectMeta(data *buildSvcData) v12.ObjectMeta {
	om := v12.ObjectMeta{
		Name:      data.key.Name,
		Namespace: data.key.Namespace,
		Labels:    data.labels,
	}

	return om
}

func (s *SvcServiceImpl) svcServiceSpec(data *buildSvcData) v13.ServiceSpec {
	ss := v13.ServiceSpec{
		Selector:              data.labels,
		Ports:                 s.svcServicePort(data.sbp),
		Type:                  v13.ServiceTypeLoadBalancer,
		ExternalTrafficPolicy: v13.ServiceExternalTrafficPolicyTypeLocal,
	}

	return ss
}

func (s *SvcServiceImpl) svcServicePort(sbp []*v1.ServiceBackendPort) []v13.ServicePort {
	var sps = make([]v13.ServicePort, 0, len(sbp))

	var name string
	for _, v := range sbp {
		switch v.Number {
		case 80:
			name = "http"
		case 443:
			name = "https"
		default:
			name = fmt.Sprintf("port-%d", v.Number)
		}

		if v.Number == 0 {
			continue
		}

		sp := v13.ServicePort{
			Name: name,
			Port: v.Number,
			TargetPort: intstr.IntOrString{
				IntVal: v.Number,
			},
			Protocol: v13.ProtocolTCP,
		}

		sps = append(sps, sp)
	}

	//extraPort := int32(svcPort)
	//sps = append(sps, v13.ServicePort{
	//	Name: fmt.Sprintf("port-%d", extraPort),
	//	Port: extraPort,
	//	TargetPort: intstr.IntOrString{
	//		IntVal: extraPort,
	//	},
	//	Protocol: v13.ProtocolTCP,
	//})

	return sps
}

func (s *SvcServiceImpl) ingressSvc() error {
	var bks = make([]*v1.ServiceBackendPort, 0, 10)
	//config, err := s.generic.GetUpstreamConfig()
	//if err != nil {
	//	return err
	//}

	// 获取ingress配置文件中的所有svc
	for _, p := range constants.HttpPorts {
		sp := &v1.ServiceBackendPort{
			Name:   fmt.Sprintf("%s-%d", s.generic.GetDeployNameLabel(), p),
			Number: p,
		}
		bks = append(bks, sp)
	}
	//for _, b1 := range config {
	//	for _, b2 := range b1.ServiceBackend {
	//		bks = append(bks, b2.Services)
	//	}
	//}

	if s.config.EnableStream.EnableStream {
		for _, s1 := range s.config.EnableStream.StreamBackendList {
			sp := &v1.ServiceBackendPort{
				Name:   s1.Name,
				Number: s1.Port,
			}
			bks = append(bks, sp)
		}
	}

	//backend, err := s.generic.GetDefaultBackend()
	//if err != nil {
	//	return err
	//}
	//
	//if backend.Name != "" && backend.Number > 0 {
	//	bks = append(bks, backend)
	//}

	// controller的data plane
	ctlSvcKey := types.NamespacedName{Name: s.generic.GetDeploySvcName(), Namespace: s.generic.GetNameSpace()}
	data := &buildSvcData{
		key:    ctlSvcKey,
		sbp:    bks,
		labels: map[string]string{"app": s.generic.GetDeployLabel()},
	}
	svc, err := s.generic.GetService(ctlSvcKey)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := s.CreateSvc(data); err != nil {
				return err
			}
		}
		return err
	}

	if err := s.UpdateSvc(svc, data); err != nil {
		return err
	}

	return nil
}

func (s *SvcServiceImpl) CheckSvc() error {
	if err := s.ingressSvc(); err != nil {
		return err
	}

	return nil
}
