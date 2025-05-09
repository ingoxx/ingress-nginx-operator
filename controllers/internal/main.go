package internal

import (
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations"
	"github.com/ingoxx/ingress-nginx-operator/pkg/adapter"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/services"
	"golang.org/x/net/context"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CrdNginxController struct {
	k8sCli      common.K8sClientSet
	operatorCli common.OperatorClientSet
	ctx         context.Context
}

func NewCrdNginxController(ctx context.Context, k8sCli common.K8sClientSet, operatorCli common.OperatorClientSet) *CrdNginxController {
	return &CrdNginxController{
		ctx:         ctx,
		k8sCli:      k8sCli,
		operatorCli: operatorCli,
	}
}

func (nc *CrdNginxController) Start(req ctrl.Request) error {
	ing := services.NewIngressServiceImpl(nc.ctx, nc.k8sCli, nc.operatorCli)

	if _, err := ing.GetIngress(nc.ctx, req.NamespacedName); cerr.IsIngressNotFoundError(err) {
		return err
	}

	if err := ing.CheckController(); err != nil {
		return err
	}

	if err := ing.CheckService(); err != nil {
		return err
	}

	cert := services.NewCertServiceImpl(nc.ctx, ing)
	secret := services.NewSecretServiceImpl(nc.ctx, ing, cert)
	issuer := services.NewIssuerServiceImpl(nc.ctx, ing, cert)

	ar := adapter.ResourceAdapter{
		Ingress: ing,
		Secret:  secret,
		Cert:    cert,
		Issuer:  issuer,
	}

	if err := ar.CheckCert(); err != nil {
		return err
	}

	extract, err := annotations.NewExtractor(ing).Extract()
	if err != nil {
		return err
	}

	if err := NewNginxController(ar, extract).Run(); err != nil {
		return err
	}
	return nil
}
