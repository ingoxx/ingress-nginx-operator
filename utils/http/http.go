package main

import (
	"encoding/json"
	"errors"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/file"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

func main() {
	if err := file.StartWatch(); err != nil {
		klog.Fatalf("fail to start file watch process, error msg '%s'", err.Error())
		return
	}

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
	FileName  string `json:"file_name"`
	FileBytes []byte `json:"file_bytes"`
}

func updateNginxCfg(resp http.ResponseWriter, req *http.Request) {
	var ncp NginxCfgParams
	if req.Header.Get("X-Auth-Token") != "your-shared-secret-token" {
		http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(resp, "invalid body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &ncp); err != nil {
		http.Error(resp, "json parsing failed", http.StatusBadRequest)
		return
	}

	if len(ncp.FileBytes) == 0 {
		http.Error(resp, "missing or empty 'file_bytes' parameter", http.StatusBadRequest)
		return
	}

	if ncp.FileName == "" {
		http.Error(resp, "missing or empty 'file_name' parameter", http.StatusBadRequest)
		return
	}

}
