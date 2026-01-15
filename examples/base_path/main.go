package main

import (
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

	// create and start the API server with custom title and a custom path
	basePath := "/admin/cron/"
	srv := server.NewServer(scheduler, *port, server.WithTitle(*title), server.WithBasePath(basePath))
	router := http.NewServeMux()
	router.Handle(basePath, srv.Router)

	// start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		log.Println("\n" + strings.Repeat("=", 70))
		log.Printf("GoCron UI Server Started")
		log.Println(strings.Repeat("=", 70))
		log.Printf("Web UI:       http://localhost%s/admin/cron/", addr)
		log.Printf("API:          http://localhost%s/admin/cron/api", addr)
		log.Printf("WebSocket:    ws://localhost%s/admin/cron/ws", addr)
		log.Printf("Total Jobs:   %d", len(scheduler.Jobs()))
		log.Println(strings.Repeat("=", 70) + "\n")

		if err := http.ListenAndServe(addr, router); err != nil {
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
