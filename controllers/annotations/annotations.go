package annotations

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/allowcos"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/allowiplist"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/denyiplist"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/limitreq"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/loadBalance"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/parser"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/rewrite"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/ssl"
	"github.com/ingoxx/ingress-nginx-operator/controllers/annotations/stream"
	cerr "github.com/ingoxx/ingress-nginx-operator/pkg/error"
	"github.com/ingoxx/ingress-nginx-operator/pkg/service"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
)

type IngressAnnotationsConfig struct {
	Rewrite           rewrite.Config
	LoadBalance       loadBalance.Config
	SSLStapling       ssl.Config
	EnableCos         allowcos.Config
	EnableReqLimit    limitreq.Config
	EnableStream      stream.Config
	EnableIpWhileList allowiplist.Config
	EnableIpBlackList denyiplist.Config
}

func (iac *IngressAnnotationsConfig) GetIngAnnConfig() {}

type Extractor struct {
	annotations map[string]parser.IngressAnnotationsParser
	ingress     service.K8sResourcesIngress // 将会移除, 使用ResourcesMth接口调各个资源的方法
	resources   service.ResourcesMth
}

func NewExtractor(ing service.K8sResourcesIngress, resources service.ResourcesMth) *Extractor {
	return &Extractor{
		annotations: map[string]parser.IngressAnnotationsParser{
			"Rewrite":           rewrite.NewRewrite(ing, resources),
			"LoadBalance":       loadBalance.NewLoadBalanceIng(ing, resources),
			"SSLStapling":       ssl.NewSSL(ing, resources),
			"EnableCos":         allowcos.NewEnableCosIng(ing, resources),
			"EnableReqLimit":    limitreq.NewRequestLimitIng(ing, resources),
			"EnableStream":      stream.NewEnableStreamIng(ing, resources),
			"EnableIpWhileList": allowiplist.NewEnableIpWhiteListIng(ing, resources),
			"EnableIpBlackList": denyiplist.NewEnableIpBlackListIng(ing, resources),
		},
		ingress:   ing,
		resources: resources,
	}
}

func (e *Extractor) Extract() (*IngressAnnotationsConfig, error) {
	iak := new(IngressAnnotationsConfig)

	ia := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		if err := annotationParser.Validate(e.ingress.GetAnnotations()); err != nil {
			klog.Error(err)
			// 验证不通过就不让写入配置
			return iak, err
		}

		val, err := annotationParser.Parse()
		if err != nil {
			if cerr.IsMissIngressAnnotationsError(err) {
				continue
			}

			return nil, err
		}

		if val != nil {
			ia[name] = val
		}
	}

	if err := mapstructure.Decode(ia, iak); err != nil {
		klog.ErrorS(err, fmt.Sprintf("unexpected error merging extracted annotations, ingress '%s', namespace '%s'", e.ingress.GetName(), e.ingress.GetNameSpace()))
		return nil, err
	}

	//if err := mergo.MapWithOverwrite(iak, ia); err != nil {
	//	klog.ErrorS(err, fmt.Sprintf("unexpected error merging extracted annotations, ingress '%s', namespace '%s'", e.ingress.GetName(), e.ingress.GetNameSpace()))
	//	return nil, err
	//}

	return iak, nil
}
