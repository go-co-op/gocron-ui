package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-co-op/gocron-ui/server"
	"github.com/go-co-op/gocron/v2"
)

const (
	usernameEnv = "GOCRON_UI_USERNAME"
	passwordEnv = "GOCRON_UI_PASSWORD"
)

var jobs = []struct {
	name       string
	definition gocron.JobDefinition
	task       gocron.Task
	options    []gocron.JobOption
}{
	{
		"simple-10s-interval", gocron.DurationJob(10 * time.Second), gocron.NewTask(func() { log.Println("Running 10-second interval job") }), []gocron.JobOption{gocron.WithName("simple-10s-interval"), gocron.WithTags("interval", "simple")},
	},
	{
		"simple-5s-interval", gocron.DurationJob(5 * time.Second), gocron.NewTask(func() { log.Println("Running 5-second interval job") }), []gocron.JobOption{gocron.WithName("simple-5s-interval"), gocron.WithTags("interval", "simple")},
	},
	{
		"simple-20s-interval", gocron.DurationJob(20 * time.Second), gocron.NewTask(func() { log.Println("Running 20-second interval job") }), []gocron.JobOption{gocron.WithName("simple-20s-interval"), gocron.WithTags("interval", "simple")},
	},
}

func main() {
	username, usernameOK := os.LookupEnv(usernameEnv)
	password, passwordOK := os.LookupEnv(passwordEnv)

	if (!usernameOK || !passwordOK) || username == "" || password == "" {
		log.Fatalf("Environment variables %s and %s must be set for basic authentication", usernameEnv, passwordEnv)
	}

	port := flag.Int("port", 8080, "Port to run the server on")
	title := flag.String("title", "GoCron Scheduler", "Custom title for the UI")
	flag.Parse()

	// create the gocron scheduler
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// add jobs to the scheduler
	for _, job := range jobs {
		if _, err := scheduler.NewJob(job.definition, job.task, job.options...); err != nil {
			log.Printf("Error creating job: %v", err)
		}
	}

	// start the scheduler
	scheduler.Start()
	log.Println("Scheduler started with", len(scheduler.Jobs()), "jobs")

	// create and start the API server with custom title
	srv := server.NewServer(scheduler, *port, server.WithTitle(*title))

	// start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		log.Println("\n" + strings.Repeat("=", 70))
		log.Printf("GoCron UI Server Started with Basic Authentication")
		log.Println(strings.Repeat("=", 70))
		log.Printf("Web UI:       http://localhost%s", addr)
		log.Printf("API:          http://localhost%s/api", addr)
		log.Printf("WebSocket:    ws://localhost%s/ws", addr)
		log.Printf("Total Jobs:   %d", len(scheduler.Jobs()))
		log.Println(strings.Repeat("=", 70) + "\n")

		if err := http.ListenAndServe(addr, basicAuthMiddleware(srv.Router, username, password)); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\nShutting down server...")

	// shutdown scheduler
	if err := scheduler.Shutdown(); err != nil {
		log.Printf("Error shutting down scheduler: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// https://www.alexedwards.net/blog/basic-authentication-in-go
func basicAuthMiddleware(next http.Handler, expectedUsername, expectedPassword string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			expectedUsernameHash := sha256.Sum256([]byte(expectedUsername))
			expectedPasswordHash := sha256.Sum256([]byte(expectedPassword))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
