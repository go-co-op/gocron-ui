package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

//go:embed static/*
var staticFiles embed.FS

// Server is the main server struct which contains the scheduler, router, webSocket clients, webSocket mutex, upgrader, and config
type Server struct {
	Scheduler gocron.Scheduler
	Router    http.Handler
	wsClients map[*websocket.Conn]bool
	wsMutex   sync.RWMutex
	upgrader  websocket.Upgrader
	config    Config
}

// Config is the server configuration in which user can set the title of the UI
type Config struct {
	Title            string `json:"title"`
	APIEnabled       bool   `json:"api_enabled"`
	WebSocketEnabled bool   `json:"websocket_enabled"`
}

func defaultConfig() Config {
	return Config{
		Title:            "GoCron UI",
		APIEnabled:       true,
		WebSocketEnabled: true,
	}
}

// NewServer creates a new server instance
func NewServer(scheduler gocron.Scheduler, _ int, opts ...Option) *Server {
	s := &Server{
		Scheduler: scheduler,
		wsClients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true // allow all origins for development
			},
		},
		config: defaultConfig(),
	}

	// apply options
	for _, opt := range opts {
		opt(s)
	}

	router := mux.NewRouter()

	if s.config.APIEnabled {
		// api routes
		api := router.PathPrefix("/api").Subrouter()
		api.HandleFunc("/config", s.GetConfig).Methods("GET")
		api.HandleFunc("/jobs", s.GetJobs).Methods("GET")
		api.HandleFunc("/jobs", s.CreateJob).Methods("POST")
		api.HandleFunc("/jobs/{id}", s.GetJob).Methods("GET")
		api.HandleFunc("/jobs/{id}", s.DeleteJob).Methods("DELETE")
		api.HandleFunc("/jobs/{id}/run", s.RunJob).Methods("POST")
		api.HandleFunc("/scheduler/stop", s.StopScheduler).Methods("POST")
		api.HandleFunc("/scheduler/start", s.StartScheduler).Methods("POST")
	}

	if s.config.WebSocketEnabled {
		// webSocket route
		router.HandleFunc("/ws", s.HandleWebSocket)
	}

	// serve embedded static files (frontend)
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to load static files: %v", err)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(staticFS)))

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	s.Router = c.Handler(router)

	// start broadcasting job updates
	go s.broadcastJobUpdates()

	return s
}

// Option is a functional option for configuring the server
type Option func(*Server)

// WithTitle sets a custom title for the UI
func WithTitle(title string) Option {
	return func(s *Server) {
		s.config.Title = title
	}
}

// WithAPIDisabled disables API
func WithAPIDisabled() Option {
	return func(s *Server) {
		s.config.APIEnabled = false
	}
}

// WithWebSocketDisabled disables webSocket
func WithWebSocketDisabled() Option {
	return func(s *Server) {
		s.config.WebSocketEnabled = false
	}
}

// GetConfig gets server configuration
func (s *Server) GetConfig(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, s.config)
}

// HandleWebSocket is a webSocket handler
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.wsMutex.Lock()
	s.wsClients[conn] = true
	s.wsMutex.Unlock()

	log.Printf("WebSocket client connected. Total clients: %d", len(s.wsClients))

	// send initial job list
	jobs := s.getJobsData()
	if err := conn.WriteJSON(map[string]interface{}{
		"type": "jobs",
		"data": jobs,
	}); err != nil {
		log.Printf("Error sending initial jobs: %v", err)
	}

	// keep connection alive and handle client disconnection
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.wsMutex.Lock()
			delete(s.wsClients, conn)
			s.wsMutex.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(s.wsClients))
			break
		}
	}
}

// broadcastJobUpdates broadcasts job updates to all connected webSocket clients
func (s *Server) broadcastJobUpdates() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.wsMutex.RLock()
		if len(s.wsClients) == 0 {
			s.wsMutex.RUnlock()
			continue
		}
		s.wsMutex.RUnlock()

		jobs := s.getJobsData()
		message := map[string]interface{}{
			"type": "jobs",
			"data": jobs,
		}

		s.wsMutex.RLock()
		for client := range s.wsClients {
			err := client.WriteJSON(message)
			if err != nil {
				log.Printf("Error broadcasting to client: %v", err)
				s.wsMutex.RUnlock()
				s.wsMutex.Lock()
				delete(s.wsClients, client)
				client.Close()
				s.wsMutex.Unlock()
				s.wsMutex.RLock()
			}
		}
		s.wsMutex.RUnlock()
	}
}

// GetJobs gets all jobs
func (s *Server) GetJobs(w http.ResponseWriter, _ *http.Request) {
	jobs := s.getJobsData()
	respondJSON(w, http.StatusOK, jobs)
}

// GetJob gets a single job
func (s *Server) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	jobs := s.Scheduler.Jobs()
	for _, job := range jobs {
		if job.ID() == id {
			jobData := s.convertJobToData(job)
			respondJSON(w, http.StatusOK, jobData)
			return
		}
	}

	respondError(w, http.StatusNotFound, "Job not found")
}

// CreateJob creates a new job
func (s *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Job name is required")
		return
	}

	// create job definition based on type
	var jobDef gocron.JobDefinition
	var err error

	switch req.Type {
	case "duration":
		if req.Interval <= 0 {
			respondError(w, http.StatusBadRequest, "Interval must be positive for duration jobs")
			return
		}
		jobDef = gocron.DurationJob(time.Duration(req.Interval) * time.Second)

	case "cron":
		if req.CronExpression == "" {
			respondError(w, http.StatusBadRequest, "Cron expression is required for cron jobs")
			return
		}
		jobDef = gocron.CronJob(req.CronExpression, false)

	case "daily":
		if req.Interval <= 0 {
			respondError(w, http.StatusBadRequest, "Interval must be positive for daily jobs")
			return
		}
		if req.AtTime == "" {
			respondError(w, http.StatusBadRequest, "AtTime is required for daily jobs")
			return
		}
		atTime, err := parseTime(req.AtTime)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid time format. Use HH:MM:SS")
			return
		}
		jobDef = gocron.DailyJob(uint(req.Interval), gocron.NewAtTimes(atTime))

	default:
		respondError(w, http.StatusBadRequest, "Invalid job type. Supported: duration, cron, daily")
		return
	}

	// create the task (example: just logs the job name)
	task := gocron.NewTask(func() {
		log.Printf("Executing job: %s", req.Name)
	})

	// create job options
	options := []gocron.JobOption{
		gocron.WithName(req.Name),
	}
	if len(req.Tags) > 0 {
		options = append(options, gocron.WithTags(req.Tags...))
	}

	// add job to scheduler
	job, err := s.Scheduler.NewJob(jobDef, task, options...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jobData := s.convertJobToData(job)
	respondJSON(w, http.StatusCreated, jobData)
}

// DeleteJob deletes a job
func (s *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	if err := s.Scheduler.RemoveJob(id); err != nil { // remove job from scheduler using the job ID & RemoveJob is a method of the Scheduler interface
		respondError(w, http.StatusNotFound, "Job not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Job deleted successfully"})
}

// RunJob runs a job immediately
func (s *Server) RunJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	jobs := s.Scheduler.Jobs()
	for _, job := range jobs {
		if job.ID() == id {
			if err := job.RunNow(); err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"message": "Job executed"})
			return
		}
	}

	respondError(w, http.StatusNotFound, "Job not found")
}

// StopScheduler stops the scheduler
func (s *Server) StopScheduler(w http.ResponseWriter, _ *http.Request) {
	if err := s.Scheduler.StopJobs(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "Scheduler stopped"})
}

// StartScheduler starts the scheduler
func (s *Server) StartScheduler(w http.ResponseWriter, _ *http.Request) {
	s.Scheduler.Start()
	respondJSON(w, http.StatusOK, map[string]string{"message": "Scheduler started"})
}

// helper functions
func (s *Server) getJobsData() []JobData {
	jobs := s.Scheduler.Jobs()
	result := make([]JobData, 0, len(jobs))
	for _, job := range jobs {
		result = append(result, s.convertJobToData(job))
	}
	return result
}

func (s *Server) convertJobToData(job gocron.Job) JobData {
	nextRun, _ := job.NextRun()
	lastRun, _ := job.LastRun()

	// get next 5 runs
	nextRuns, _ := job.NextRuns(5)

	// determine schedule info based on job name patterns or intervals
	schedule, scheduleDetail := s.inferSchedule(job, nextRuns)

	return JobData{
		ID:             job.ID().String(),
		Name:           job.Name(),
		Tags:           job.Tags(),
		NextRun:        formatTime(nextRun),
		LastRun:        formatTime(lastRun),
		NextRuns:       formatTimes(nextRuns),
		Schedule:       schedule,
		ScheduleDetail: scheduleDetail,
	}
}

func (s *Server) inferSchedule(job gocron.Job, nextRuns []time.Time) (string, string) {
	name := job.Name()

	// try to infer from job name
	if len(name) > 0 {
		// check for common patterns in name
		if strings.Contains(name, "every") || strings.Contains(name, "interval") {
			if strings.Contains(name, "10s") || strings.Contains(name, "10-s") {
				return "Every 10 seconds", "Duration: 10s"
			}
			if strings.Contains(name, "5s") {
				return "Every 5 seconds", "Duration: 5s"
			}
			if strings.Contains(name, "minute") {
				return "Every minute", "Duration: 1m"
			}
		}

		if strings.Contains(name, "cron") {
			return "Cron schedule", "Cron: * * * * *"
		}

		if strings.Contains(name, "daily") {
			return "Daily", "Daily schedule"
		}

		if strings.Contains(name, "weekly") {
			return "Weekly", "Weekly: Mon, Wed, Fri"
		}

		if strings.Contains(name, "random") {
			return "Random interval", "Random: 5-15s"
		}

		if strings.Contains(name, "singleton") {
			return "Every 5 seconds (singleton)", "Duration: 5s, Mode: Singleton"
		}

		if strings.Contains(name, "limited") {
			return "Every 7 seconds (limited)", "Duration: 7s, Max runs: 3"
		}

		if strings.Contains(name, "parameter") {
			return "Every 12 seconds", "Duration: 12s"
		}

		if strings.Contains(name, "context") {
			return "Every 8 seconds", "Duration: 8s"
		}

		if strings.Contains(name, "one-time") || strings.Contains(name, "onetime") {
			return "One time only", "OneTime job"
		}
	}

	// try to infer from next runs interval
	if len(nextRuns) >= 2 {
		interval := nextRuns[1].Sub(nextRuns[0])

		if interval < time.Minute {
			seconds := int(interval.Seconds())
			return fmt.Sprintf("Every %d seconds", seconds), fmt.Sprintf("Duration: %ds", seconds)
		}
		if interval < time.Hour {
			minutes := int(interval.Minutes())
			return fmt.Sprintf("Every %d minutes", minutes), fmt.Sprintf("Duration: %dm", minutes)
		}
		if interval < 24*time.Hour {
			hours := int(interval.Hours())
			return fmt.Sprintf("Every %d hours", hours), fmt.Sprintf("Duration: %dh", hours)
		}
		days := int(interval.Hours() / 24)
		return fmt.Sprintf("Every %d days", days), fmt.Sprintf("Duration: %dd", days)
	}

	return "Scheduled", "Custom schedule"
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatTimes(times []time.Time) []string {
	result := make([]string, 0, len(times))
	for _, t := range times {
		result = append(result, formatTime(t))
	}
	return result
}

func parseTime(timeStr string) (gocron.AtTime, error) {
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		// try without seconds
		t, err = time.Parse("15:04", timeStr)
		if err != nil {
			return nil, err
		}
	}
	return gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second())), nil
}

// respondJSON responds with JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
