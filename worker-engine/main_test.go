package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler(t *testing.T) {
	w := NewWorker()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	w.healthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", resp["status"])
	}
	if resp["service"] != "worker-engine" {
		t.Errorf("expected worker-engine service, got %v", resp["service"])
	}
}

func TestSubmitHandler(t *testing.T) {
	w := NewWorker()
	go w.processJobs()

	body := `{"id":"test-1","task":"send_email","priority":"high","payload":{"to":"user@test.com"}}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	w.submitHandler(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rr.Code)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["job_id"] != "test-1" {
		t.Errorf("expected job_id test-1, got %v", resp["job_id"])
	}
}

func TestSubmitHandlerInvalidMethod(t *testing.T) {
	w := NewWorker()
	req := httptest.NewRequest(http.MethodGet, "/submit", nil)
	rr := httptest.NewRecorder()

	w.submitHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestSubmitHandlerMissingFields(t *testing.T) {
	w := NewWorker()
	body := `{"task":"no_id"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	w.submitHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStatusHandler(t *testing.T) {
	w := NewWorker()
	go w.processJobs()

	body := `{"id":"status-test","task":"test_task","priority":"normal"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	w.submitHandler(rr, req)

	time.Sleep(200 * time.Millisecond)

	req = httptest.NewRequest(http.MethodGet, "/status?id=status-test", nil)
	rr = httptest.NewRecorder()
	w.statusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var job Job
	json.NewDecoder(rr.Body).Decode(&job)
	if job.Status != "completed" {
		t.Errorf("expected completed, got %s", job.Status)
	}
}

func TestStatusHandlerNotFound(t *testing.T) {
	w := NewWorker()
	req := httptest.NewRequest(http.MethodGet, "/status?id=missing", nil)
	rr := httptest.NewRecorder()

	w.statusHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestStatsHandler(t *testing.T) {
	w := NewWorker()
	go w.processJobs()

	body := `{"id":"stats-1","task":"t1","priority":"low"}`
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	w.submitHandler(rr, req)

	time.Sleep(200 * time.Millisecond)

	req = httptest.NewRequest(http.MethodGet, "/stats", nil)
	rr = httptest.NewRecorder()
	w.statsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var stats map[string]int
	json.NewDecoder(rr.Body).Decode(&stats)
	if stats["total"] < 1 {
		t.Errorf("expected at least 1 total job, got %d", stats["total"])
	}
}
