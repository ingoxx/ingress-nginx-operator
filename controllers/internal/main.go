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

	if err := ing.CheckController(); err != nil {
		nc.recorder.Event(ingress, "Normal", "NoCustomControllerSelected", err.Error())
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

	ar.Svc = services.NewSvcServiceImpl(nc.ctx, ing, ar)
	ar.Deployment = services.NewDeploymentServiceImpl(nc.ctx, ing, ar)

	extract := annotations.NewExtractor(ing, ar)

	ngx := NewNginxController()

	if !controllerutil.ContainsFinalizer(ingress, constants.Finalizer) {
		controllerutil.AddFinalizer(ingress, constants.Finalizer)
		if err := ing.UpdateIngress(ingress); err != nil {
			return err
		}
	}

	// 删除ingress
	if !ingress.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := nc.delete(ingress, ing, ar, extract, ngx); err != nil {
			nc.recorder.Event(ingress, "Warning", "FailToDeleteIngress", err.Error())
			return err
		}

		return nil

	} else {
		if err := nc.run(ingress, ing, ar, extract, ngx); err != nil {
			nc.recorder.Event(ingress, "Warning", "IngressValidationFailed", err.Error())
			return err
		}

	}

	return nil
}

func (nc *CrdNginxController) run(ingress *v1.Ingress, ing common.Generic, ar service.ResourcesMth, extract *annotations.Extractor, ngx *NginxController) error {
	if err := ing.CheckService(); err != nil {
		nc.recorder.Event(ingress, "Warning", "BackendsNoServiceAvailable", err.Error())
		return err
	}

	if err := ar.CheckCert(); err != nil {
		nc.recorder.Event(ingress, "Warning", "NoCertAvailable", err.Error())
		return err
	}

	config, err := extract.Extract()
	if err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToExtractAnnotations", err.Error())
		return err
	}

	if err := ngx.Run(ar, config); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToGenerateNgxConfig", err.Error())
		return err
	}

	nc.recorder.Event(ingress, "Normal", "RunSuccessfully", fmt.Sprintf("'%s' ingress update successfully", ingress.Name))

	return nil
}

func (nc *CrdNginxController) delete(ingress *v1.Ingress, ing common.Generic, ar service.ResourcesMth, extract *annotations.Extractor, ngx *NginxController) error {
	ngx.IsDel = true

	config, err := extract.Extract()
	if err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToExtractAnnotations", err.Error())
		return err
	}

	if err := ngx.Run(ar, config); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToGenerateNgxConfig", err.Error())
		return err
	}

	if err := ar.DeleteConfigMap(); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToDeleteConfigMap", err.Error())
		return err
	}

	if err := ar.DeleteCert(); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToDeleteCert", err.Error())
		return err
	}

	if err := ar.DeleteIssuer(); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToDeleteIssuer", err.Error())
		return err
	}

	if err := ar.DeleteSecret(); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToDeleteSecret", err.Error())
		return err
	}

	if controllerutil.RemoveFinalizer(ingress, constants.Finalizer) {
		if err := ing.UpdateIngress(ingress); err != nil {
			return err
		}
	}

	klog.Infof("the ingress %s has been successfully deleted\n", ingress.Name)

	return nil
}
