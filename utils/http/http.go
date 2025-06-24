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

func (nc NginxCfgParams) H(respData map[string]interface{}) {
	nc.resp.Header().Set("Content-Type", "application/json")
	nc.resp.WriteHeader((respData["status"]).(int))

	b, err := json.Marshal(&respData)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("json marshal failed, esg '%s'", err.Error()))
	}

	if _, err := nc.resp.Write(b); err != nil {
		klog.Fatalf(fmt.Sprintf("respone failed, esg '%s'", err.Error()))
	}
}

func updateNginxCfg(resp http.ResponseWriter, req *http.Request) {
	var ncp = NginxCfgParams{resp: resp}
	if req.Header.Get("X-Auth-Token") != "your-shared-secret-token" {
		ncp.H(map[string]interface{}{
			"code":   1001,
			"msg":    "request unauthorized",
			"status": http.StatusUnauthorized,
		})
		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		ncp.H(map[string]interface{}{
			"code":   1002,
			"msg":    "bad request header",
			"status": http.StatusBadRequest,
		})
		return
	}

	if req.Method != http.MethodPost {
		ncp.H(map[string]interface{}{
			"code":   1003,
			"msg":    "bad request method",
			"status": http.StatusBadRequest,
		})
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		ncp.H(map[string]interface{}{
			"code":   1004,
			"msg":    "invalid body",
			"status": http.StatusBadRequest,
		})
		return
	}

	if err := json.Unmarshal(body, &ncp); err != nil {
		ncp.H(map[string]interface{}{
			"code":   1005,
			"msg":    "json parsing failed",
			"status": http.StatusBadRequest,
		})
		return
	}

	if len(ncp.FileBytes) == 0 {
		ncp.H(map[string]interface{}{
			"code":   1005,
			"msg":    "missing or empty 'file_bytes' parameter",
			"status": http.StatusBadRequest,
		})
		return
	}

	if ncp.FileName == "" {
		ncp.H(map[string]interface{}{
			"code":   1006,
			"msg":    "missing or empty 'file_name' parameter",
			"status": http.StatusBadRequest,
		})
		return
	}

	fmt.Println("file content >>> ", string(ncp.FileBytes))
	fmt.Println("file name >>> ", ncp.FileName)

	if err := file.SaveToFile(ncp.FileName, ncp.FileBytes); err != nil {
		klog.Error(fmt.Sprintf("save to file failed, file name '%s'", ncp.FileName))
	}

	ncp.H(map[string]interface{}{
		"code":   1000,
		"msg":    "update nginx config ok",
		"status": http.StatusOK,
	})
}
