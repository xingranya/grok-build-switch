package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAppQuitEndpointRequestsShutdown(t *testing.T) {
	quit := make(chan struct{}, 1)
	server := &Server{OnQuit: func() { quit <- struct{}{} }}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/app/quit", nil)
	server.handleAppQuit(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("POST /api/app/quit status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	select {
	case <-quit:
	case <-time.After(time.Second):
		t.Fatal("退出回调没有执行")
	}
}

func TestAppQuitEndpointRejectsUnavailableOrWrongMethod(t *testing.T) {
	server := &Server{}
	recorder := httptest.NewRecorder()
	server.handleAppQuit(recorder, httptest.NewRequest(http.MethodPost, "/api/app/quit", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("未初始化退出服务时 status = %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	server.handleAppQuit(recorder, httptest.NewRequest(http.MethodGet, "/api/app/quit", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /api/app/quit status = %d", recorder.Code)
	}
}
