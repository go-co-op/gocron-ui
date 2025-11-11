package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-co-op/gocron-ui/server"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var jobs = []struct {
	definition gocron.JobDefinition
	task       gocron.Task
	options    []gocron.JobOption
}{
	{
		gocron.DurationJob(10 * time.Second),
		gocron.NewTask(func() { log.Println("Running 10-second interval job") }),
		[]gocron.JobOption{gocron.WithName("simple-10s-interval"), gocron.WithTags("interval", "simple")},
	},
	{
		gocron.DurationJob(5 * time.Second),
		gocron.NewTask(func() { log.Println("Running 5-second interval job") }),
		[]gocron.JobOption{gocron.WithName("simple-5s-interval"), gocron.WithTags("interval", "simple")},
	},

	{
		gocron.CronJob("* * * * *", false),
		gocron.NewTask(func() { log.Println("Cron job executed (every minute)") }),
		[]gocron.JobOption{gocron.WithName("cron-every-minute"), gocron.WithTags("cron", "periodic")},
	},
	{
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(14, 30, 0))),
		gocron.NewTask(func() { log.Println("Daily job executed at 2:30 PM") }),
		[]gocron.JobOption{gocron.WithName("daily-afternoon-report"), gocron.WithTags("daily", "report")},
	},
	{
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday, time.Wednesday, time.Friday), gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
		gocron.NewTask(func() { log.Println("Weekly job executed (Mon, Wed, Fri at 9:00 AM)") }),
		[]gocron.JobOption{gocron.WithName("weekly-mwf-morning"), gocron.WithTags("weekly", "morning", "report")},
	},
	{
		gocron.DurationJob(12 * time.Second),
		gocron.NewTask(func(name string, count int) { log.Printf("Job with parameters: name=%s, count=%d", name, count) }, "example-job", 42),
		[]gocron.JobOption{gocron.WithName("parameterized-job"), gocron.WithTags("parameters", "demo")},
	},
	{
		gocron.DurationJob(8 * time.Second),
		gocron.NewTask(func(ctx context.Context) { log.Printf("Job with context executed, context: %v", ctx) }),
		[]gocron.JobOption{gocron.WithName("context-aware-job"), gocron.WithTags("context", "advanced")},
	},
	{
		gocron.DurationRandomJob(5*time.Second, 15*time.Second),
		gocron.NewTask(func() { log.Println("Random interval job executed (5-15 seconds)") }),
		[]gocron.JobOption{gocron.WithName("random-interval-job"), gocron.WithTags("random", "variable")},
	},
	{
		gocron.DurationJob(5 * time.Second),
		gocron.NewTask(func() {
			log.Println("Singleton job started")
			time.Sleep(8 * time.Second)
			log.Println("Singleton job completed")
		}),
		[]gocron.JobOption{gocron.WithName("singleton-mode-job"), gocron.WithTags("singleton", "long-running"), gocron.WithSingletonMode(gocron.LimitModeReschedule)},
	},
	{
		gocron.DurationJob(7 * time.Second),
		gocron.NewTask(func() { log.Println("Limited run job executed") }),
		[]gocron.JobOption{gocron.WithName("limited-run-job"), gocron.WithTags("limited", "demo"), gocron.WithLimitedRuns(3)},
	},
	{
		gocron.DurationJob(15 * time.Second),
		gocron.NewTask(func() {
			log.Println("Job with listeners executed")
			time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
		}),
		[]gocron.JobOption{
			gocron.WithName("event-listener-job"),
			gocron.WithTags("events", "monitoring"),
			gocron.WithEventListeners(
				gocron.AfterJobRuns(func(_ uuid.UUID, jobName string) {
					log.Printf("   → AfterJobRuns: %s completed", jobName)
				}),
				gocron.BeforeJobRuns(func(_ uuid.UUID, jobName string) {
					log.Printf("   → BeforeJobRuns: %s starting", jobName)
				}),
			),
		},
	},
	{
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(30 * time.Second))),
		gocron.NewTask(func() { log.Println("One-time job executed!") }),
		[]gocron.JobOption{gocron.WithName("one-time-job"), gocron.WithTags("onetime", "scheduled")},
	},
	{
		gocron.DurationJob(20 * time.Second),
		gocron.NewTask(func() {
			items := rand.Intn(100) + 1
			log.Printf("Processing %d items...", items)
			time.Sleep(2 * time.Second)
			log.Printf("Successfully processed %d items", items)
		}),
		[]gocron.JobOption{gocron.WithName("data-processor-job"), gocron.WithTags("processing", "batch")},
	},
	{
		gocron.DurationJob(30 * time.Second),
		gocron.NewTask(func() {
			status := "healthy"
			if rand.Float32() < 0.1 {
				status = "degraded"
			}
			log.Printf("Health check: System is %s", status)
		}),
		[]gocron.JobOption{gocron.WithName("health-check-job"), gocron.WithTags("monitoring", "health")},
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

	// create and start the API server with custom title
	srv := server.NewServer(scheduler, *port, server.WithTitle(*title))

	// start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		log.Println("\n" + strings.Repeat("=", 70))
		log.Printf("GoCron UI Server Started")
		log.Println(strings.Repeat("=", 70))
		log.Printf("Web UI:       http://localhost%s", addr)
		log.Printf("API:          http://localhost%s/api", addr)
		log.Printf("WebSocket:    ws://localhost%s/ws", addr)
		log.Printf("Total Jobs:   %d", len(scheduler.Jobs()))
		log.Println(strings.Repeat("=", 70) + "\n")

		if err := http.ListenAndServe(addr, srv.Router); err != nil {
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
