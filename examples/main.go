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

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	title := flag.String("title", "GoCron Scheduler", "Custom title for the UI")
	flag.Parse()

	// create the gocron scheduler
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// example 1: Simple interval job - runs every 10 seconds
	_, err = scheduler.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(func() {
			log.Println("Running 10-second interval job")
		}),
		gocron.WithName("simple-10s-interval"),
		gocron.WithTags("interval", "simple"),
	)
	if err != nil {
		log.Printf("Error creating simple interval job: %v", err)
	}

	// example 2: Fast job - runs every 5 seconds
	_, err = scheduler.NewJob(
		gocron.DurationJob(5*time.Second),
		gocron.NewTask(func() {
			log.Println("Fast 5-second job executed")
		}),
		gocron.WithName("fast-5s-job"),
		gocron.WithTags("interval", "fast"),
	)
	if err != nil {
		log.Printf("Error creating fast job: %v", err)
	}

	// example 3: Cron job - runs every minute
	_, err = scheduler.NewJob(
		gocron.CronJob("* * * * *", false),
		gocron.NewTask(func() {
			log.Println("Cron job executed (every minute)")
		}),
		gocron.WithName("cron-every-minute"),
		gocron.WithTags("cron", "periodic"),
	)
	if err != nil {
		log.Printf("Error creating cron job: %v", err)
	}

	// example 4: Daily job at specific time
	_, err = scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(14, 30, 0), // 2:30 PM
		)),
		gocron.NewTask(func() {
			log.Println("Daily job executed at 2:30 PM")
		}),
		gocron.WithName("daily-afternoon-report"),
		gocron.WithTags("daily", "report"),
	)
	if err != nil {
		log.Printf("Error creating daily job: %v", err)
	}

	// example 5: Weekly job - runs on specific days
	_, err = scheduler.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday, time.Wednesday, time.Friday),
			gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
		gocron.NewTask(func() {
			log.Println("Weekly job executed (Mon, Wed, Fri at 9:00 AM)")
		}),
		gocron.WithName("weekly-mwf-morning"),
		gocron.WithTags("weekly", "morning", "report"),
	)
	if err != nil {
		log.Printf("Error creating weekly job: %v", err)
	}

	// example 6: Job with parameters
	_, err = scheduler.NewJob(
		gocron.DurationJob(12*time.Second),
		gocron.NewTask(func(name string, count int) {
			log.Printf("Job with parameters: name=%s, count=%d", name, count)
		}, "example-job", 42),
		gocron.WithName("parameterized-job"),
		gocron.WithTags("parameters", "demo"),
	)
	if err != nil {
		log.Printf("Error creating parameterized job: %v", err)
	}

	// example 7: Job with context
	_, err = scheduler.NewJob(
		gocron.DurationJob(8*time.Second),
		gocron.NewTask(func(ctx context.Context) {
			log.Printf("Job with context executed, context: %v", ctx)
		}),
		gocron.WithName("context-aware-job"),
		gocron.WithTags("context", "advanced"),
	)
	if err != nil {
		log.Printf("Error creating context job: %v", err)
	}

	// example 8: Random duration job
	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(5*time.Second, 15*time.Second),
		gocron.NewTask(func() {
			log.Println("Random interval job executed (5-15 seconds)")
		}),
		gocron.WithName("random-interval-job"),
		gocron.WithTags("random", "variable"),
	)
	if err != nil {
		log.Printf("Error creating random job: %v", err)
	}

	// example 9: Job with singleton mode (prevents overlapping runs)
	_, err = scheduler.NewJob(
		gocron.DurationJob(5*time.Second),
		gocron.NewTask(func() {
			log.Println("Singleton job started")
			time.Sleep(8 * time.Second) // Simulate long-running task
			log.Println("Singleton job completed")
		}),
		gocron.WithName("singleton-mode-job"),
		gocron.WithTags("singleton", "long-running"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		log.Printf("Error creating singleton job: %v", err)
	}

	// example 10: Limited run job (runs only 3 times)
	_, err = scheduler.NewJob(
		gocron.DurationJob(7*time.Second),
		gocron.NewTask(func() {
			log.Println("Limited run job executed")
		}),
		gocron.WithName("limited-run-job"),
		gocron.WithTags("limited", "demo"),
		gocron.WithLimitedRuns(3),
	)
	if err != nil {
		log.Printf("Error creating limited run job: %v", err)
	}

	// example 11: Job with event listeners
	_, err = scheduler.NewJob(
		gocron.DurationJob(15*time.Second),
		gocron.NewTask(func() {
			log.Println("Job with listeners executed")
			// Simulate some work
			time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
		}),
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
	)
	if err != nil {
		log.Printf("Error creating event listener job: %v", err)
	}

	// example 12: One-time job (runs once at a specific time)
	oneTimeAt := time.Now().Add(30 * time.Second)
	_, err = scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(oneTimeAt)),
		gocron.NewTask(func() {
			log.Println("One-time job executed!")
		}),
		gocron.WithName("one-time-job"),
		gocron.WithTags("onetime", "scheduled"),
	)
	if err != nil {
		log.Printf("Error creating one-time job: %v", err)
	}

	// example 13: Job that simulates data processing
	_, err = scheduler.NewJob(
		gocron.DurationJob(20*time.Second),
		gocron.NewTask(func() {
			items := rand.Intn(100) + 1
			log.Printf("Processing %d items...", items)
			time.Sleep(2 * time.Second)
			log.Printf("Successfully processed %d items", items)
		}),
		gocron.WithName("data-processor-job"),
		gocron.WithTags("processing", "batch"),
	)
	if err != nil {
		log.Printf("Error creating data processor job: %v", err)
	}

	// example 14: Health check job
	_, err = scheduler.NewJob(
		gocron.DurationJob(30*time.Second),
		gocron.NewTask(func() {
			status := "healthy"
			if rand.Float32() < 0.1 { // 10% chance of unhealthy
				status = "degraded"
			}
			log.Printf("Health check: System is %s", status)
		}),
		gocron.WithName("health-check-job"),
		gocron.WithTags("monitoring", "health"),
	)
	if err != nil {
		log.Printf("Error creating health check job: %v", err)
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
