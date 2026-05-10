package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Job struct {
	ID        string                 `json:"id"`
	Task      string                 `json:"task"`
	Priority  string                 `json:"priority"`
	Status    string                 `json:"status"`
	Payload   map[string]interface{} `json:"payload"`
	StartedAt string                 `json:"started_at,omitempty"`
	EndedAt   string                 `json:"ended_at,omitempty"`
	Result    string                 `json:"result,omitempty"`
}

type Worker struct {
	mu       sync.Mutex
	jobs     map[string]*Job
	queue    chan *Job
	logger   *log.Logger
}

func NewWorker() *Worker {
	return &Worker{
		jobs:   make(map[string]*Job),
		queue:  make(chan *Job, 100),
		logger: log.New(os.Stdout, "[worker-engine] ", log.LstdFlags),
	}
}

func (w *Worker) processJobs() {
	for job := range w.queue {
		w.mu.Lock()
		job.Status = "processing"
		job.StartedAt = time.Now().UTC().Format(time.RFC3339)
		w.mu.Unlock()

		w.logger.Printf("Processing job %s (task=%s)", job.ID, job.Task)
		time.Sleep(100 * time.Millisecond)

		w.mu.Lock()
		job.Status = "completed"
		job.EndedAt = time.Now().UTC().Format(time.RFC3339)
		job.Result = fmt.Sprintf("Task '%s' completed successfully", job.Task)
		w.mu.Unlock()

		w.logger.Printf("Job %s completed", job.ID)
	}
}

func (w *Worker) healthHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "worker-engine",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"queue_size": len(w.queue),
	})
}

func (w *Worker) submitHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		w.logger.Printf("Invalid job submission: %v", err)
		http.Error(rw, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if job.ID == "" || job.Task == "" {
		http.Error(rw, `{"error":"fields 'id' and 'task' are required"}`, http.StatusBadRequest)
		return
	}

	w.mu.Lock()
	job.Status = "queued"
	w.jobs[job.ID] = &job
	w.mu.Unlock()

	w.queue <- &job
	w.logger.Printf("Job %s submitted (task=%s, priority=%s)", job.ID, job.Task, job.Priority)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusAccepted)
	json.NewEncoder(rw).Encode(map[string]string{"status": "accepted", "job_id": job.ID})
}

func (w *Worker) statusHandler(rw http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(rw, `{"error":"query param 'id' is required"}`, http.StatusBadRequest)
		return
	}

	w.mu.Lock()
	job, exists := w.jobs[jobID]
	w.mu.Unlock()

	if !exists {
		http.Error(rw, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(job)
}

func (w *Worker) statsHandler(rw http.ResponseWriter, r *http.Request) {
	w.mu.Lock()
	total := len(w.jobs)
	var completed, processing, queued int
	for _, j := range w.jobs {
		switch j.Status {
		case "completed":
			completed++
		case "processing":
			processing++
		case "queued":
			queued++
		}
	}
	w.mu.Unlock()

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]int{
		"total":      total,
		"completed":  completed,
		"processing": processing,
		"queued":     queued,
	})
}

func main() {
	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8081"
	}

	worker := NewWorker()
	go worker.processJobs()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", worker.healthHandler)
	mux.HandleFunc("/submit", worker.submitHandler)
	mux.HandleFunc("/status", worker.statusHandler)
	mux.HandleFunc("/stats", worker.statsHandler)

	worker.logger.Printf("Starting PulseQueue Worker Engine on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		worker.logger.Fatalf("Server failed: %v", err)
	}
}
