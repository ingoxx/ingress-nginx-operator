package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/file"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"sync"
	"time"
)

var (
	lock   sync.Mutex
	fileCh = make(chan NginxCfgParams, 10)
)

//var lock sync.Mutex
//var fileCh = make(chan NginxCfgParams, 10)

func main() {
	go func() {
		if err := file.IsNginxRunning(); err != nil {
			klog.ErrorS(err, err.Error())
		}
	}()

	StartHttp()
}

func StartHttp() {
	klog.Info("httpserver start")
	httpServer()
}

func httpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthCheck)
	mux.HandleFunc("/api/v1/nginx/config/update", updateNginxCfg)
	mux.HandleFunc("/api/v1/nginx/config/delete", deleteNginxCfg)

	listen := &http.Server{
		Addr:              ":9092",
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
	}

	if err := listen.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		klog.Fatalf("fail to start httpserver, error '%v'\n", err.Error())
	}
}

type NginxCfgParams struct {
	resp      http.ResponseWriter
	FileBytes []byte `json:"file_bytes"`
	FileName  string `json:"file_name"`
}

type RespData struct {
	Msg    string `json:"msg"`
	Code   int    `json:"code"`
	Status int    `json:"status"`
}

func (nc NginxCfgParams) H(rd RespData) {
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

func deleteNginxCfg(resp http.ResponseWriter, req *http.Request) {
	var ncp = NginxCfgParams{resp: resp}
	if req.Header.Get("X-Auth-Token") != constants.AuthToken {
		ncp.H(RespData{
			Msg:    "request unauthorized",
			Code:   1001,
			Status: http.StatusUnauthorized,
		})
		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		ncp.H(RespData{
			Code:   1002,
			Msg:    "bad request header",
			Status: http.StatusBadRequest,
		})
		return
	}

	if req.Method != http.MethodPost {
		ncp.H(RespData{
			Code:   1003,
			Msg:    "bad request method",
			Status: http.StatusBadRequest,
		})
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		ncp.H(RespData{
			Code:   1004,
			Msg:    "invalid body",
			Status: http.StatusBadRequest,
		})
		return
	}

	if err := json.Unmarshal(body, &ncp); err != nil {
		ncp.H(RespData{
			Code:   1005,
			Msg:    "json parsing failed",
			Status: http.StatusBadRequest,
		})
		return
	}

	if ncp.FileName == "" {
		ncp.H(RespData{
			Code:   1006,
			Msg:    "missing or empty 'file_name' parameter",
			Status: http.StatusBadRequest,
		})
		return
	}

	if len(ncp.FileBytes) == 0 {
		if err := file.HandleDeleteNgxConfig(ncp.FileName); err != nil {
			ncp.H(RespData{
				Code:   1007,
				Msg:    err.Error(),
				Status: http.StatusBadRequest,
			})
			return
		}
	}

	ncp.H(RespData{
		Code:   1000,
		Msg:    "update nginx config ok",
		Status: http.StatusOK,
	})

}

func updateNginxCfg(resp http.ResponseWriter, req *http.Request) {
	var ncp = NginxCfgParams{resp: resp}
	if req.Header.Get("X-Auth-Token") != constants.AuthToken {
		ncp.H(RespData{
			Msg:    "request unauthorized",
			Code:   1001,
			Status: http.StatusUnauthorized,
		})
		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		ncp.H(RespData{
			Code:   1002,
			Msg:    "bad request header",
			Status: http.StatusBadRequest,
		})
		return
	}

	if req.Method != http.MethodPost {
		ncp.H(RespData{
			Code:   1003,
			Msg:    "bad request method",
			Status: http.StatusBadRequest,
		})
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		ncp.H(RespData{
			Code:   1004,
			Msg:    "invalid body",
			Status: http.StatusBadRequest,
		})
		return
	}

	if err := json.Unmarshal(body, &ncp); err != nil {
		ncp.H(RespData{
			Code:   1005,
			Msg:    "json parsing failed",
			Status: http.StatusBadRequest,
		})
		return
	}

	if len(ncp.FileBytes) == 0 {
		ncp.H(RespData{
			Code:   1005,
			Msg:    "missing or empty 'file_bytes' parameter",
			Status: http.StatusBadRequest,
		})
		return
	}

	if ncp.FileName == "" {
		ncp.H(RespData{
			Code:   1006,
			Msg:    "missing or empty 'file_name' parameter",
			Status: http.StatusBadRequest,
		})
		return
	}

	if err := file.HandleConfigUpdate(ncp.FileName, ncp.FileBytes); err != nil {
		ncp.H(RespData{
			Code:   1007,
			Msg:    err.Error(),
			Status: http.StatusBadRequest,
		})

		return
	}

	select {
	case fileCh <- ncp:
		ncp.H(RespData{
			Code:   1000,
			Msg:    "update nginx config ok",
			Status: http.StatusOK,
		})
	default:
		ncp.H(RespData{
			Code:   1008,
			Msg:    "handling timeouts",
			Status: http.StatusOK,
		})
	}

	//ncp.H(RespData{
	//	Code:   1000,
	//	Msg:    "update nginx config ok",
	//	Status: http.StatusOK,
	//})
}

func healthCheck(resp http.ResponseWriter, req *http.Request) {
	var ncp = NginxCfgParams{resp: resp}

	if req.Method != http.MethodGet {
		http.Error(resp, "Method not allowed", http.StatusBadRequest)
		return
	}

	ncp.H(RespData{
		Code:   1000,
		Msg:    "health check ok",
		Status: http.StatusOK,
	})

}
