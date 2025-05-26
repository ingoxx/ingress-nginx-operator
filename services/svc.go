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

func (s *SvcServiceImpl) UpdateSvc(data *buildSvcData) error {
	return s.generic.GetClient().Update(s.ctx, s.buildSvcData(data))
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
	var sps = make([]v13.ServicePort, 0, len(sbp)+1)

	usedNames := make(map[string]bool)

	for _, v := range sbp {
		var name string
		switch v.Number {
		case 80:
			name = "http"
		case 443:
			name = "https"
		default:
			name = fmt.Sprintf("port-%d", v.Number)
		}

		// 确保 name 唯一
		original := name
		i := 1
		for usedNames[name] {
			name = fmt.Sprintf("%s-%d", original, i)
			i++
		}
		usedNames[name] = true

		sp := v13.ServicePort{
			Name: name,
			Port: v.Number,
			TargetPort: intstr.IntOrString{
				IntVal: v.Number,
			},
			Protocol: v13.ProtocolTCP, // 建议显式指定
		}

		sps = append(sps, sp)
	}

	// 额外添加固定端口（如 9092）
	extraPort := 9092
	name := "http"
	if usedNames[name] {
		name = fmt.Sprintf("port-%d", extraPort)
	}
	sps = append(sps, v13.ServicePort{
		Name: name,
		Port: int32(extraPort),
		TargetPort: intstr.IntOrString{
			IntVal: int32(extraPort),
		},
		Protocol: v13.ProtocolTCP,
	})

	return sps
}

func (s *SvcServiceImpl) streamSvc() error {
	if s.config.EnableStream.EnableStream {
		for _, s1 := range s.config.EnableStream.StreamBackendList {
			ingressSvcKey := types.NamespacedName{Name: s1.Name, Namespace: s1.Namespace}
			ports, err := s.generic.GetBackendPorts(ingressSvcKey)
			if err != nil {
				return err
			}

			// controller的data plane
			ctlSvcKey := types.NamespacedName{Name: s.generic.GetDaemonSvcName(), Namespace: s1.Namespace}

			data := &buildSvcData{
				key:    ctlSvcKey,
				sbp:    ports,
				labels: map[string]string{"app": s.generic.GetDaemonSetLabel()},
			}
			_, err = s.generic.GetService(ctlSvcKey)
			if err != nil {
				if errors.IsNotFound(err) {
					if err := s.CreateSvc(data); err != nil {
						return err
					}

					continue
				}

				return err
			}

			if err := s.UpdateSvc(data); err != nil {
				return err
			}

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
		for _, b2 := range b1.Services {
			bks = append(bks, b2)
		}
	}

	// controller的data plane
	ctlSvcKey := types.NamespacedName{Name: s.generic.GetDeploySvcName(), Namespace: s.generic.GetNameSpace()}
	data := &buildSvcData{
		key:    ctlSvcKey,
		sbp:    bks,
		labels: map[string]string{"app": s.generic.GetDeployLabel()},
	}
	_, err = s.generic.GetService(ctlSvcKey)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := s.CreateSvc(data); err != nil {
				return err
			}
		}
		return err
	}

	if err := s.UpdateSvc(data); err != nil {
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
