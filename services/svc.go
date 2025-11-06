package services

import (
	"encoding/json"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"golang.org/x/net/context"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var svcLocks = sync.Map{}

type buildSvcData struct {
	sbp    []*v1.ServiceBackendPort
	labels map[string]string
	key    client.ObjectKey
}

type SvcServiceImpl struct {
	generic          common.Generic
	ctx              context.Context
	allResourcesData service.ResourcesMth
	config           *annotations.IngressAnnotationsConfig
}

func NewSvcServiceImpl(ctx context.Context, clientSet common.Generic, allRes service.ResourcesMth, config *annotations.IngressAnnotationsConfig) *SvcServiceImpl {
	return &SvcServiceImpl{ctx: ctx, generic: clientSet, allResourcesData: allRes, config: config}
}

func (s *SvcServiceImpl) getSvcLock(svc string) *sync.Mutex {
	svcName := fmt.Sprintf("%s/%s", svc, s.generic.GetNameSpace())
	val, _ := svcLocks.LoadOrStore(svcName, &sync.Mutex{})
	return val.(*sync.Mutex)
}

func (s *SvcServiceImpl) GetAllEndPoints() ([]string, error) {
	var podIPs []string
	endpoints, err := s.generic.GetClientSet().CoreV1().Endpoints(s.generic.GetNameSpace()).Get(s.ctx, constants.SvcHandlesName, v12.GetOptions{})
	if err != nil {
		return podIPs, err
	}

	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			podIPs = append(podIPs, addr.IP)
		}
	}

	return podIPs, nil
}

func (s *SvcServiceImpl) GetSvc(key client.ObjectKey) (*v13.Service, error) {
	var svc = new(v13.Service)
	if err := s.generic.GetClient().Get(s.ctx, key, svc); err != nil {
		return svc, err
	}

	return svc, nil
}

func (s *SvcServiceImpl) UpdateSvc(svc *v13.Service, data *buildSvcData) error {
	lock := s.getSvcLock(constants.DeploySvcName)
	lock.Lock()
	defer lock.Unlock()

	svc.Spec.Ports = s.svcServicePort(data.sbp)
	if err := s.generic.GetClient().Update(s.ctx, svc); err != nil {
		return err
	}

	if err := s.UpdateHandlesSvc(data); err != nil {
		return err
	}

	return nil
}

func (s *SvcServiceImpl) DeleteSvc(svc *v13.Service) error {
	return s.generic.GetClient().Delete(s.ctx, svc)
}

func (s *SvcServiceImpl) CreateSvc(data *buildSvcData) error {
	if err := s.generic.GetClient().Create(s.ctx, s.buildSvcData(data)); err != nil {
		return err
	}

	if err := s.CreateHandlesSvc(data); err != nil {
		return err
	}

	return nil
}

// CreateHandlesSvc 创建无头svc
func (s *SvcServiceImpl) CreateHandlesSvc(data *buildSvcData) error {
	svc := &v13.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:      constants.SvcHandlesName,
			Namespace: data.key.Namespace,
			Labels:    data.labels,
		},
		Spec: v13.ServiceSpec{
			ClusterIP: "None",
			Selector:  data.labels,
			Ports:     s.svcServicePort(data.sbp),
		},
	}
	if err := s.generic.GetClient().Create(s.ctx, svc); err != nil {
		return err
	}

	return nil
}

// UpdateHandlesSvc 更新无头svc
func (s *SvcServiceImpl) UpdateHandlesSvc(data *buildSvcData) error {
	lock := s.getSvcLock(constants.SvcHandlesName)

	lock.Lock()
	defer lock.Unlock()

	data.key.Name = constants.SvcHandlesName
	svc, err := s.generic.GetService(data.key)
	if err != nil {
		return err
	}

	svc.Spec.Ports = s.svcServicePort(data.sbp)

	if err := s.generic.GetClient().Update(s.ctx, svc); err != nil {
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

	return sps
}

func (s *SvcServiceImpl) getLatestStreamPorts() ([]*stream.Backend, error) {
	var sb []*stream.Backend
	configMap, err := s.allResourcesData.GetNgxConfigMap(s.generic.GetNameSpace())
	if err != nil {
		return sb, err
	}

	data, ok := configMap[constants.StreamKey]
	if !ok {
		return sb, nil
	}

	if data != "" {
		if err := json.Unmarshal([]byte(data), &sb); err != nil {
			return sb, err
		}
	}

	return sb, nil
}

func (s *SvcServiceImpl) ingressSvc() error {
	var bks = make([]*v1.ServiceBackendPort, 0, 10)

	// 获取ingress配置文件中的所有svc
	for _, p := range constants.HttpPorts {
		sp := &v1.ServiceBackendPort{
			Name:   fmt.Sprintf("%s-%d", s.generic.GetDeployNameLabel(), p),
			Number: p,
		}
		bks = append(bks, sp)
	}

	streamPorts, err := s.getLatestStreamPorts()
	if err != nil {
		return err
	}

	for _, s1 := range streamPorts {
		sp := &v1.ServiceBackendPort{
			Name:   s1.Name,
			Number: s1.Port,
		}
		bks = append(bks, sp)
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

			return nil
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
