package services

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
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
		Selector: data.labels,
		Ports:    s.svcServicePort(data.sbp),
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

	extraPort := int32(svcPort)
	sps = append(sps, v13.ServicePort{
		Name: fmt.Sprintf("port-%d", extraPort),
		Port: extraPort,
		TargetPort: intstr.IntOrString{
			IntVal: extraPort,
		},
		Protocol: v13.ProtocolTCP,
	})

	return sps
}

func (s *SvcServiceImpl) streamSvc() error {
	if s.config.EnableStream.EnableStream {
		var bk = make([]*v1.ServiceBackendPort, 0, 10)
		for _, s1 := range s.config.EnableStream.StreamBackendList {
			sp := &v1.ServiceBackendPort{
				Name:   s1.Name,
				Number: s1.Port,
			}
			bk = append(bk, sp)
		}

		ctlSvcKey := types.NamespacedName{Name: s.generic.GetDaemonSvcName(), Namespace: s.generic.GetNameSpace()}
		data := &buildSvcData{
			key:    ctlSvcKey,
			sbp:    bk,
			labels: map[string]string{"app": s.generic.GetDaemonSetLabel()},
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
	}

	return nil
}

func (s *SvcServiceImpl) ingressSvc() error {
	var bks = make([]*v1.ServiceBackendPort, 0, 6)
	config, err := s.generic.GetUpstreamConfig()
	if err != nil {
		return err
	}

	// 获取ingress配置文件中的所有svc
	for _, b1 := range config {
		for _, b2 := range b1.ServiceBackend {
			bks = append(bks, b2.Services)
		}
	}

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
	//var err error
	//if err1, err2 := s.streamSvc(), s.ingressSvc(); err1 != nil || err2 != nil {
	//	err = cuerr.Join(err1, err2)
	//}
	//
	//if err != nil {
	//	return err
	//}

	if err := s.ingressSvc(); err != nil {
		return err
	}

	return nil
}
