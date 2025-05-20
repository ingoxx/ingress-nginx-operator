package http

import (
	"encoding/json"
	"errors"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

func startHttp() {
	klog.Info("httpserver start")

	go func() {
		httpServer()
	}()
}

func httpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthCheck)
	mux.HandleFunc("/api/v1/nginx/config", updateNginxCfg)

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
	Type string `json:"type"`
	File []byte `json:"file"`
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

	if ncp.Type == "" {
		http.Error(resp, "missing or empty 'type' parameter", http.StatusBadRequest)
		return
	}

	if len(ncp.File) == 0 {
		http.Error(resp, "missing or empty 'file' parameter", http.StatusBadRequest)
		return
	}

}
