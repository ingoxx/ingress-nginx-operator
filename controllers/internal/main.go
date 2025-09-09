package internal

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	"github.com/ingoxx/ingress-nginx-operator/pkg/adapter"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/services"
	"golang.org/x/net/context"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type aff map[string][]*AggregatedFeatures

func (a aff) Add(key string, feature *AggregatedFeatures) {
	a[key] = append(a[key], feature)
}

type AggregatedFeatures struct {
	Ingress        *v1.Ingress
	EnableReqLimit *limitreq.ReqBackendsConfig
	EnableStream   []*stream.Backend
	NameSpace      string
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

	af := make(aff)

	for _, i := range il.Items {
		if _, err := ing.GetIngress(&i); err != nil {
			klog.Error(err)
			continue
		}

		if err := nc.check(ing, af, false); err != nil {
			nc.recorder.Eventf(&i, "Warning", "IngressDetectionFailed", err.Error())
			klog.Error(err)
			continue
		}
	}

	//聚合公共配置nginx.conf
	var st = new(stream.Config)
	var lm = new(limitreq.Config)
	for m := range af {
		for _, v := range af[m] {

			if len(v.EnableStream) > 0 {
				st.StreamBackendList = v.EnableStream
				if !st.EnableStream {
					st.EnableStream = true
				}
			}
			if len(v.EnableReqLimit.Backends) > 0 && len(v.EnableReqLimit.LimitReqZone) > 0 {
				lm.Backends = v.EnableReqLimit.Backends
				lm.LimitReqZone = v.EnableReqLimit.LimitReqZone
				if !lm.EnableRequestLimit {
					lm.EnableRequestLimit = true
				}
			}
		}
	}

	return nil
}

func (nc *CrdNginxController) check(ing common.Generic, af aff, isMainConf bool) error {
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
		a := &AggregatedFeatures{
			Ingress:      ing,
			NameSpace:    ing.GetNameSpace(),
			EnableStream: extract.EnableStream.StreamBackendList,
		}
		af.Add(ing.GetNameSpace(), a)
	}

	if extract.EnableReqLimit.EnableRequestLimit {
		b := &AggregatedFeatures{
			Ingress:        ing,
			NameSpace:      ing.GetNameSpace(),
			EnableReqLimit: &extract.EnableReqLimit.ReqBackendsConfig,
		}
		af.Add(ing.GetNameSpace(), b)
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

	if isMainConf {
		if err := NewNginxController(ar).GenerateNgxConfTmpl(extract); err != nil {
			klog.Error(err)
			return err
		}
	} else {
		if err := NewNginxController(ar).GenerateServerTmpl(extract); err != nil {
			klog.Error(err)
			return err
		}
	}

	return nil
}
