package services

import (
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

func (s *SvcServiceImpl) UpdateSvc(svc *v13.Service) error {
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
	var sps = make([]v13.ServicePort, 0, 2)
	for _, v := range sbp {
		sp := v13.ServicePort{
			Port: v.Number,
			TargetPort: intstr.IntOrString{
				IntVal: v.Number,
			},
		}

		sps = append(sps, sp)
	}
	return sps
}

func (s *SvcServiceImpl) streamSvc() error {
	if s.config.EnableStream.EnableStream {
		var streamSvc = make([]*v1.ServiceBackendPort, 0, len(s.config.EnableStream.StreamBackendList))
		for _, s1 := range s.config.EnableStream.StreamBackendList {
			daemonSetKey := types.NamespacedName{Name: s1.Name, Namespace: s1.Namespace}
			svc, err := s.GetSvc(daemonSetKey)
			if err != nil {
				if errors.IsNotFound(err) {
					port := s.generic.GetBackendPort(svc)
					svcBackendPort := &v1.ServiceBackendPort{
						Name:   svc.Name,
						Number: port,
					}
					streamSvc = append(streamSvc, svcBackendPort)
					data := &buildSvcData{
						key:    daemonSetKey,
						sbp:    streamSvc,
						labels: map[string]string{"app": s.generic.GetDaemonSetNameLabel()},
					}
					if err := s.CreateSvc(data); err != nil {
						return err
					}

					continue
				}

				return err
			}

			if err := s.UpdateSvc(svc); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SvcServiceImpl) ingressSvc() error {
	config, err := s.generic.GetUpstreamConfig()
	if err != nil {
		return err
	}

	for _, b1 := range config {
		for _, b2 := range b1.Services {
			svcKey := types.NamespacedName{Name: b2.Name, Namespace: s.generic.GetNameSpace()}
			svc, err := s.generic.GetService(svcKey)
			if err != nil {
				if errors.IsNotFound(err) {
					//deployKey := types.NamespacedName{Name: s.generic.GetName(), Namespace: s.generic.GetNameSpace()}
					data := &buildSvcData{
						key:    svcKey,
						sbp:    b1.Services,
						labels: map[string]string{"app": s.generic.GetDeployNameLabel()},
					}
					if err := s.CreateSvc(data); err != nil {
						return err
					}

					continue
				}

				return err
			}

			if err := s.UpdateSvc(svc); err != nil {
				return err
			}

		}
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
