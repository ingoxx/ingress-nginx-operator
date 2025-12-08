package service

import (
	"encoding/json"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/internal/domain"
	"k8s.io/klog/v2"
	"net/http"
)

type NginxCfgService struct {
	resp      http.ResponseWriter
	rp        domain.ReqParamsImp
	FileBytes []byte `json:"file_bytes"`
	FileName  string `json:"file_name"`
}

func NewNginxCfgParams() *NginxCfgService {
	return &NginxCfgService{}
}

func (nc NginxCfgService) H(rd domain.RespData) {
	nc.resp.Header().Set("Content-Type", "application/json")
	nc.resp.WriteHeader(rd.Status)

	b, err := json.Marshal(&rd)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("json marshal failed, esg '%s'", err.Error()))
	}

	if _, err := nc.resp.Write(b); err != nil {
		klog.Fatalf(fmt.Sprintf("respone failed, esg '%s'", err.Error()))
	}
}
