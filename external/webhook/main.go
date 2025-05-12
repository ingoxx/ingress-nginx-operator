package webhook

import (
	"golang.org/x/net/context"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type IngressValidator struct {
	decoder *admission.Decoder
}

func (v *IngressValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ing := &v1.Ingress{}
	if err := v.decoder.Decode(req, ing); err != nil {
		return admission.Errored(400, err)
	}

	// 示例校验逻辑
	//if ing.Spec.IngressClassName == nil || *ing.Spec.IngressClassName != "my-ingress-class" {
	//	return admission.Denied("ingressClassName must be 'my-ingress-class'")
	//}

	klog.Info("Ingress validated", "name", ing.Name)

	return admission.Allowed("ok")
}

func (v *IngressValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
