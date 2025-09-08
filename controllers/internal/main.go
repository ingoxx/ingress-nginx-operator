package internal

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	"github.com/ingoxx/ingress-nginx-operator/pkg/adapter"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/services"
	"golang.org/x/net/context"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type AggregatedFeatures struct {
	EnableReqLimit map[string][]*limitreq.ReqBackendsConfig
	EnableStream   map[string][]*stream.Backend
}

type CrdNginxController struct {
	k8sCli      common.K8sClientSet
	operatorCli common.OperatorClientSet
	ctx         context.Context
	recorder    record.EventRecorder
}

func NewCrdNginxController(ctx context.Context, k8sCli common.K8sClientSet, operatorCli common.OperatorClientSet, recorder record.EventRecorder) *CrdNginxController {
	return &CrdNginxController{
		ctx:         ctx,
		k8sCli:      k8sCli,
		operatorCli: operatorCli,
		recorder:    recorder,
	}
}

func (nc *CrdNginxController) Start(req ctrl.Request) error {
	ing := services.NewIngressServiceImpl(nc.ctx, nc.k8sCli, nc.operatorCli)
	il, err := ing.GetIngressList(nc.ctx, req.NamespacedName)
	if err != nil {
		klog.Error(err)
		return err
	}

	af := &AggregatedFeatures{}

	for _, i := range il.Items {
		if _, err := ing.GetIngress(&i); err != nil {
			klog.Error(err)
			continue
		}

		if err := nc.check(ing, af); err != nil {
			nc.recorder.Eventf(&i, "Warning", "IngressDetectionFailed", err.Error())
			klog.Error(err)
			continue
		}
	}

	return nil
}

func (nc *CrdNginxController) check(ing common.Generic, af *AggregatedFeatures) error {
	if err := ing.CheckController(); err != nil {
		klog.Warning(err)
		return err
	}

	if err := ing.CheckService(); err != nil {
		klog.Error(err)
		return err
	}

	cert := services.NewCertServiceImpl(nc.ctx, ing)
	secret := services.NewSecretServiceImpl(nc.ctx, ing, cert)
	issuer := services.NewIssuerServiceImpl(nc.ctx, ing, cert)
	configMap := services.NewConfigMapServiceImpl(nc.ctx, ing)

	ar := adapter.ResourceAdapter{
		Ingress:   ing,
		Secret:    secret,
		Cert:      cert,
		Issuer:    issuer,
		ConfigMap: configMap,
	}

	if err := ar.CheckCert(); err != nil {
		klog.Error(err)
		return err
	}

	extract, err := annotations.NewExtractor(ing, ar).Extract()
	if err != nil {
		klog.Error(err)
		return err
	}

	if extract.EnableStream.EnableStream {
		af.EnableStream[ing.GetNameSpace()] = append(af.EnableStream[ing.GetNameSpace()], extract.EnableStream.StreamBackendList...)
	}

	if extract.EnableReqLimit.EnableRequestLimit {
		af.EnableReqLimit[ing.GetNameSpace()] = append(af.EnableReqLimit[ing.GetNameSpace()], &extract.EnableReqLimit.ReqBackendsConfig)
	}

	deployment := services.NewDeploymentServiceImpl(nc.ctx, ing, extract)
	if err := deployment.CheckDeploy(); err != nil {
		klog.Error(err)
		return err
	}

	svc := services.NewSvcServiceImpl(nc.ctx, ing, extract)
	if err := svc.CheckSvc(); err != nil {
		klog.Error(err)
		return err
	}

	if err := NewNginxController(ar, extract).GenerateServerTmpl(extract); err != nil {
		klog.Error(err)
		return err
	}

	return nil
}
