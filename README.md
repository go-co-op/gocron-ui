# gocron-ui: A Web UI for [gocron](https://github.com/go-co-op/gocron)

[![CI State](https://github.com/go-co-op/gocron-ui/actions/workflows/go_test.yml/badge.svg?branch=main&event=push)](https://github.com/go-co-op/gocron-ui/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-co-op/gocron-ui)](https://goreportcard.com/report/github.com/go-co-op/gocron-ui)
[![Go Doc](https://godoc.org/github.com/go-co-op/gocron-ui?status.svg)](https://pkg.go.dev/github.com/go-co-op/gocron-ui)
[![Slack](https://img.shields.io/badge/gophers-gocron-brightgreen?logo=slack)](https://gophers.slack.com/archives/CQ7T0T1FW)

A lightweight, real-time web interface for monitoring and controlling [gocron](https://github.com/go-co-op/gocron) scheduled jobs.

If you want to chat, you can find us on Slack at
[<img src="https://img.shields.io/badge/gophers-gocron-brightgreen?logo=slack">](https://gophers.slack.com/archives/CQ7T0T1FW)


## Features

- **Real-time Monitoring** - WebSocket-based live job status updates
- **Job Control** - Trigger jobs manually or remove them from the scheduler
- **Schedule Preview** - View upcoming executions for each job
- **Tagging System** - Organize and filter jobs by tags
- **Embedded UI** - Static files compiled into binary, zero external dependencies
- **Portable** - Single self-contained binary deployment
- **Modern UI** - Responsive design with vanilla JavaScript (no build step)

## Installation

```bash
go get github.com/go-co-op/gocron-ui
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/go-co-op/gocron/v2"
    "github.com/go-co-op/gocron-ui/server"
)

func main() {
    // create a scheduler
    scheduler, err := gocron.NewScheduler()
    if err != nil {
        log.Fatal(err)
    }

    // add a job to the scheduler
    _, err = scheduler.NewJob(
        gocron.DurationJob(10*time.Second),
        gocron.NewTask(func() {
            log.Println("Job executed")
        }),
        gocron.WithName("example-job"),
        gocron.WithTags("important"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // start the scheduler
    scheduler.Start()

    // start the web UI server
    srv := server.NewServer(scheduler, 8080)
    log.Println("GoCron UI available at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", srv.Router))
}
```

Open your browser to `http://localhost:8080` to view the dashboard.

## API Reference

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/jobs` | List all jobs |
| `GET` | `/api/jobs/{id}` | Get job details |
| `POST` | `/api/jobs/{id}/run` | Execute job immediately |
| `DELETE` | `/api/jobs/{id}` | Remove job from scheduler |
| `POST` | `/api/scheduler/start` | Start the scheduler |
| `POST` | `/api/scheduler/stop` | Stop the scheduler |

### WebSocket

Connect to `ws://localhost:8080/ws` for real-time job updates.

**Message Format:**
```json
{
  "type": "jobs",
  "data": [
    {
      "id": "uuid",
      "name": "job-name",
      "tags": ["tag1", "tag2"],
      "nextRun": "2025-10-07T15:30:00Z",
      "lastRun": "2025-10-07T15:29:50Z",
      "nextRuns": ["...", "..."],
      "schedule": "Every 10 seconds",
      "scheduleDetail": "Duration: 10s"
    }
  ]
}
```

## Examples

### Comprehensive Example

A full-featured example demonstrating 14 different job types is available in the [exmaples](./exmaples/) directory:

```bash
cd exmaples
go run main.go
```

**Demonstrates:**
- Interval-based jobs (duration, random intervals)
- Cron expression scheduling
- Daily and weekly jobs
- Parameterized jobs with custom arguments
- Context-aware jobs
- Singleton mode (prevent overlapping executions)
- Limited run jobs
- Event listeners (before/after job runs)
- One-time scheduled jobs
- Batch processing patterns
- Health check monitoring

Visit `http://localhost:8080` to see the UI in action.

## Deployment

### Binary Distribution

```bash
# Build
go build -o gocron-ui

# Run
./gocron-ui
```

The binary is self-contained and requires no external files or dependencies.

### Docker

```bash
docker build -t gocron-ui .
docker run -p 8080:8080 gocron-ui
```

See [Dockerfile](./Dockerfile) for details.

## Configuration

The server accepts the following configuration through the `NewServer` function:

```go
server.NewServer(scheduler gocron.Scheduler, port int) *Server
```

**Parameters:**
- `scheduler` - Your configured gocron scheduler instance
- `port` - HTTP port to listen on

## Important Notes

### Job Creation Limitation

GoCron UI is a **monitoring and control interface** for jobs defined in your Go code. Jobs cannot be created from the UI because they require compiled Go functions to execute. The UI provides:

- ✅ Real-time monitoring
- ✅ Manual job triggering
- ✅ Job deletion
- ✅ Schedule viewing
- Job creation (must be done in code)

## Production Considerations

- **Authentication**: This package does not include authentication. Implement your own auth middleware if deploying publicly.
- **CORS**: Default CORS settings allow all origins. Restrict this in production environments.
- **Error Handling**: Implement proper error logging and monitoring for production use.

## Maintainer

<div align="left">
    <div style="display: flex; justify-content: left; align-items: center; gap: 6px;">
        <a href="https://github.com/iyashjayesh" style="display: flex; align-items: center; text-decoration: none;">
            <img src="https://avatars.githubusercontent.com/u/53042582" width="20" style="border-radius:50%;" />
            <span style="margin-left: 5px; font-weight: bold;">Yash</span>
        </a>
        <span>&lt;&gt; Maintainer</span>
    </div>
</div>

## Star History

<a href="https://www.star-history.com/#go-co-op/gocron-ui&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=go-co-op/gocron-ui&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=go-co-op/gocron-ui&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=go-co-op/gocron-ui&type=Date" />
 </picture>
</a>