package service

import (
	"encoding/json"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/internal/domain"
	"io"
	"k8s.io/klog/v2"
	"net/http"
)

type RespService struct {
	req  *http.Request
	resp http.ResponseWriter
}

func NewRespService(resp http.ResponseWriter, req *http.Request) *RespService {
	return &RespService{
		req:  req,
		resp: resp,
	}
}

func (nc *RespService) B() ([]byte, error) {
	body, err := io.ReadAll(nc.req.Body)
	if err != nil {
		return body, err
	}

	return body, nil
}

func (nc *RespService) H(rd domain.RespData) {
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
