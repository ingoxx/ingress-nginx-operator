package internal

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/adapter"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/services"
	"golang.org/x/net/context"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
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

	ingress, err := ing.GetIngress(nc.ctx, req.NamespacedName)
	if cerr.IsIngressNotFoundError(err) {
		ingRef := &v1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: req.Namespace,
				Name:      req.Name,
			},
		}
		nc.recorder.Event(ingRef, "Warning", "IngressNotExist", err.Error())

		return err
	}

	if err := ing.CheckController(); err != nil {
		nc.recorder.Event(ingress, "Normal", "NoCustomControllerSelected", err.Error())
		return err
	}

	if err := ing.CheckService(); err != nil {
		nc.recorder.Event(ingress, "Warning", "BackendsNoServiceAvailable", err.Error())
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
		nc.recorder.Event(ingress, "Warning", "NoCertAvailable", err.Error())
		return err
	}

	extract, err := annotations.NewExtractor(ing, ar).Extract()
	if err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToExtractAnnotations", err.Error())
		return err
	}

	deployment := services.NewDeploymentServiceImpl(nc.ctx, ing, extract)
	if err := deployment.CheckDeploy(); err != nil {
		nc.recorder.Event(ingress, "Warning", "DeployDetectionFailed", err.Error())
		return err
	}

	svc := services.NewSvcServiceImpl(nc.ctx, ing, extract)
	if err := svc.CheckSvc(); err != nil {
		nc.recorder.Event(ingress, "Warning", "ServiceDetectionFailed", err.Error())
		return err
	}

	if err := NewNginxController(ar, extract).Run(); err != nil {
		nc.recorder.Event(ingress, "Warning", "FailToGenerateNgxConfig", err.Error())
		return err
	}

	nc.recorder.Event(ingress, "Normal", "RunSuccessfully", "Ingress created successfully")

	return nil
}
