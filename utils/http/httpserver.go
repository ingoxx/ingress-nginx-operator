package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/file"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

func main() {
	go func() {
		if err := file.StartWatch(); err != nil {
			klog.Fatalf("fail to start file watch process, error msg '%s'", err.Error())
			return
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

	listen := &http.Server{
		Addr:              ":9092",
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
	}

	if err := listen.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		klog.Fatalf("fail to start httpserver, error '%v'\n", err.Error())
	}
}

func healthCheck(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(resp, "Method not allowed", http.StatusBadRequest)
		return
	}

	resp.WriteHeader(http.StatusOK)
	if _, err := resp.Write([]byte("OK")); err != nil {
		klog.Fatalf("fail to response, error '%v'\n", err.Error())
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

func updateNginxCfg(resp http.ResponseWriter, req *http.Request) {
	var ncp = NginxCfgParams{resp: resp}
	if req.Header.Get("X-Auth-Token") != "k8s" {
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

	if err := file.SaveToFile(ncp.FileName, ncp.FileBytes); err != nil {
		klog.Error(fmt.Sprintf("save to file failed, file name '%s'", ncp.FileName))
	}

	ncp.H(RespData{
		Code:   1000,
		Msg:    "update nginx config ok",
		Status: http.StatusOK,
	})
}
