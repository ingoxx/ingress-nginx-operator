package webhook

import (
	"context"
	"fmt"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type IngressValidator struct {
	Client  client.Client
	Decoder *admission.Decoder
}

// Handle 必须实现 Handle 接口
func (v *IngressValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &v1.Ingress{}

	// 1. 解析请求中的对象
	if err := v.Decoder.Decode(req, ingress); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// 只校验 Create 操作
	if req.Operation != "CREATE" {
		return admission.Allowed("not a CREATE operation")
	}

	// 2. 查询 namespace 中是否存在同名的 ingress
	key := types.NamespacedName{
		Namespace: ingress.Namespace,
		Name:      ingress.Name,
	}

	existing := &v1.Ingress{}
	err := v.Client.Get(ctx, key, existing)

	// 已存在则拒绝
	if err == nil {
		return admission.Denied(
			fmt.Sprintf("Ingress '%s' already exists in namespace '%s'",
				ingress.Name, ingress.Namespace),
		)
	}

	// 如果不是 NotFound 则说明是查询错误
	if err != nil && !errors.IsNotFound(err) {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.Allowed("Ingress name is unique")
}

// InjectDecoder decoder 注入
func (v *IngressValidator) InjectDecoder(d *admission.Decoder) error {
	v.Decoder = d
	return nil
}

func RegisterWebhook(mgr manager.Manager) error {
	mgr.GetLogger().Info("Registering Ingress validating webhook")

	hookServer := mgr.GetWebhookServer()

	hookServer.Register("/validate-ingress", &webhook.Admission{
		Handler: &IngressValidator{
			Client: mgr.GetClient(),
		},
	})

	return nil
}
