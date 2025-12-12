package internal

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/adapter"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/ingoxx/ingress-nginx-operator/services"
	"golang.org/x/net/context"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type CrdNginxController struct {
	k8sCli      common.K8sClientSet
	operatorCli common.OperatorClientSet
	recorder    record.EventRecorder
	ctx         context.Context
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

	ingList, err := ing.GetIngressList(nc.ctx, req.NamespacedName)
	if err != nil {
		return err
	}

	for _, ig := range ingList.Items {
		if err := nc.check(&ig, ing); err != nil {
			return err
		}
	}

	return nil
}

func (nc *CrdNginxController) check(ingress *v1.Ingress, ing common.Generic) error {
	ing.NewIngress(ingress)

	cert := services.NewCertServiceImpl(nc.ctx, ing)
	secret := services.NewSecretServiceImpl(nc.ctx, ing, cert)
	issuer := services.NewIssuerServiceImpl(nc.ctx, ing, cert)
	configMap := services.NewConfigMapServiceImpl(nc.ctx, ing)
	ngx := NewNginxController()

	if err := ing.CheckController(); err != nil {
		nc.recorder.Event(ingress, "Normal", "NoCustomControllerSelected", err.Error())
		return err
	}

	ar := adapter.ResourceAdapter{
		Ingress:   ing,
		Secret:    secret,
		Cert:      cert,
		Issuer:    issuer,
		ConfigMap: configMap,
	}

	if err := ing.CheckService(); err != nil {
		nc.recorder.Event(ingress, "Warning", "BackendsNoServiceAvailable", err.Error())
		return err
	}

	if err := ar.CheckCert(); err != nil {
		nc.recorder.Event(ingress, "Warning", "NoCertAvailable", err.Error())
		return err
	}

	extract, err := annotations.NewExtractor(ing, ar).Extract()
	if err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToExtractAnnotations", err.Error())
		return err
	}

	ar.Svc = services.NewSvcServiceImpl(nc.ctx, ing, ar, extract)

	ar.Deployment = services.NewDeploymentServiceImpl(nc.ctx, ing, ar, extract)

	if !controllerutil.ContainsFinalizer(ingress, constants.Finalizer) {
		controllerutil.AddFinalizer(ingress, constants.Finalizer)
		if err := ing.UpdateIngress(ingress); err != nil {
			return err
		}
	}

	if !ingress.ObjectMeta.DeletionTimestamp.IsZero() {
		// 删除ingress后的逻辑处理
		ngx.IsDel = true
		if err := ar.ConfigMap.DeleteConfigMap(); err != nil {
			nc.recorder.Event(ingress, "Warning", "FailToDeleteConfigMap", err.Error())
			return err
		}

		if err := ar.Cert.DeleteCert(); err != nil {
			nc.recorder.Event(ingress, "Warning", "FailToDeleteCert", err.Error())
			return err
		}

		if err := ar.Issuer.DeleteIssuer(); err != nil {
			nc.recorder.Event(ingress, "Warning", "FailToDeleteIssuer", err.Error())
			return err
		}

		if err := ar.Secret.DeleteSecret(); err != nil {
			nc.recorder.Event(ingress, "Warning", "FailToDeleteSecret", err.Error())
			return err
		}

		if err := ngx.Run(ar, extract); err != nil {
			nc.recorder.Event(ingress, "Warning", "FailToGenerateNgxConfig", err.Error())
			return err
		}

		if controllerutil.RemoveFinalizer(ingress, constants.Finalizer) {
			if err := ing.UpdateIngress(ingress); err != nil {
				return err
			}
		}

		klog.Infof("ingress %s deletion complete", ingress.Name)

		return nil
	}

	if err := ngx.Run(ar, extract); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToGenerateNgxConfig", err.Error())
		return err
	}

	nc.recorder.Event(ingress, "Normal", "RunSuccessfully", fmt.Sprintf("'%s' ingress update successfully", ingress.Name))

	return nil
}

func (nc *CrdNginxController) start(ingress *v1.Ingress, ing common.Generic, resources service.ResourcesMth) {

}

func (nc *CrdNginxController) stop(ingress *v1.Ingress, ing common.Generic, resources service.ResourcesMth) {
}
