# GoCron UI - Examples

Get up and running with GoCron UI in **under 60 seconds** and explore 14 different job types demonstrating the full capabilities of gocron.

## Quick Start

```bash
# run with default settings (port 8080)
go run main.go

# or specify a custom port
go run main.go -port 3000
```

**Open your browser:** http://localhost:8080

### Build Binary (Optional)

```bash
go build -o example main.go
./example
```

The binary is self-contained and can run from anywhere!

---

## What You'll See

A real-time dashboard with **14 different jobs** demonstrating:

| # | Job Type | Description | Interval |
|---|----------|-------------|----------|
| 1 | Simple Interval | Basic duration-based execution | 10s |
| 2 | Fast Job | High-frequency execution | 5s |
| 3 | Cron Job | Cron expression scheduling | Every minute |
| 4 | Daily Job | Daily at specific time | 2:30 PM |
| 5 | Weekly Job | Weekly on specific days | Mon, Wed, Fri at 9:00 AM |
| 6 | Parameterized Job | Job with custom parameters | 12s |
| 7 | Context-Aware Job | Job using context | 8s |
| 8 | Random Interval Job | Variable timing | 5-15s (random) |
| 9 | Singleton Mode Job | Prevents overlapping executions | 5s (runs for 8s) |
| 10 | Limited Run Job | Executes limited times | 7s (3 times only) |
| 11 | Event Listener Job | Monitors job lifecycle | 15s |
| 12 | One-Time Job | Runs once at scheduled time | 30s after start |
| 13 | Data Processor Job | Simulates batch processing | 20s |
| 14 | Health Check Job | Monitors system health | 30s |

### UI Features

| Feature | Description |
|---------|-------------|
| **Real-time Updates** | Live WebSocket-based status updates |
| **Manual Triggers** | Execute any job on-demand |
| **Job Deletion** | Remove jobs from scheduler |
| **Schedule Preview** | View next 5 scheduled runs |
| **Tag Filtering** | Organize jobs with tags |

---

## Key Concepts Demonstrated

### Job Definitions

```go
// duration-based
gocron.DurationJob(10*time.Second)

// cron expression
gocron.CronJob("* * * * *", false)

// daily at specific time
gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(14, 30, 0)))

// weekly on specific days
gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday, time.Wednesday, time.Friday),
    gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0)))

// random interval
gocron.DurationRandomJob(5*time.Second, 15*time.Second)

// one-time execution
gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(30*time.Second)))
```

### Job Options

```go
// set job name
gocron.WithName("my-job")

// add tags for organization
gocron.WithTags("production", "critical")

// prevent overlapping executions
gocron.WithSingletonMode(gocron.LimitModeReschedule)

// limit total executions
gocron.WithLimitedRuns(3)

// add lifecycle event listeners
gocron.WithEventListeners(
    gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
        log.Printf("Starting: %s", jobName)
    }),
    gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
        log.Printf("Completed: %s", jobName)
    }),
)
```

### Task Functions

```go
// simple task
gocron.NewTask(func() {
    log.Println("Task executed")
})

// task with parameters
gocron.NewTask(func(name string, count int) {
    log.Printf("Processing: %s (%d)", name, count)
}, "data", 42)

// task with context
gocron.NewTask(func(ctx context.Context) {
    select {
    case <-ctx.Done():
        log.Println("Task cancelled")
    default:
        log.Println("Task running")
    }
})
```

---

## API Usage

### REST Endpoints

```bash
# list all jobs
curl http://localhost:8080/api/jobs

# get specific job
curl http://localhost:8080/api/jobs/{job-id}

# run job immediately
curl -X POST http://localhost:8080/api/jobs/{job-id}/run

# delete job
curl -X DELETE http://localhost:8080/api/jobs/{job-id}

# control scheduler
curl -X POST http://localhost:8080/api/scheduler/stop
curl -X POST http://localhost:8080/api/scheduler/start
```

### WebSocket Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Jobs update:', data);
};

ws.onopen = () => {
  console.log('Connected to GoCron UI');
};
```

---

## Advanced Features

### Singleton Mode

Prevents concurrent executions (critical for long-running jobs):

```go
gocron.WithSingletonMode(gocron.LimitModeReschedule)
```

### Event Listeners

Monitor job lifecycle events:

```go
gocron.WithEventListeners(
    gocron.BeforeJobRuns(beforeFunc),
    gocron.AfterJobRuns(afterFunc),
    gocron.AfterJobRunsWithError(errorFunc),
    gocron.AfterJobRunsWithPanic(panicFunc),
)
```

### Graceful Shutdown

```go
// wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// shutdown scheduler
scheduler.Shutdown()
```

Press `Ctrl+C` to trigger graceful shutdown.

---

## Troubleshooting

### Port Already in Use

```bash
go run main.go -port 3000
```

### Jobs Not Showing

1. Ensure scheduler is started: `scheduler.Start()`
2. Check WebSocket connection in browser console
3. Review server logs for errors

---

## Learn More

| Resource | Link |
|----------|------|
| **GoCron Docs** | https://github.com/go-co-op/gocron |
| **GoCron UI** | https://github.com/go-co-op/gocron-ui |
| **Go Package** | https://pkg.go.dev/github.com/go-co-op/gocron/v2 |
| **Slack** | [#gocron channel](https://gophers.slack.com/archives/CQ7T0T1FW) |
| **Issues** | [GitHub Issues](https://github.com/go-co-op/gocron-ui/issues) |

---

**Happy Scheduling!**
