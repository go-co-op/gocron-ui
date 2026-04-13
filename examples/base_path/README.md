# GoCronUI Base Path Example

This is a simple example demonstrating using gocron-ui behind a base path (e.g., `/admin/cron/`) and using the standard http.Handler interface to serve the application.

```go
go run main.go
```

You can access the web UI at [http://localhost:8080/admin/cron/](http://localhost:8080/admin/cron/).

To access the API:

```bash
$ curl http://127.0.0.1:8080/admin/cron/api/jobs -v | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080
> GET /admin/cron/api/jobs HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 200 OK
< Content-Type: application/json
< Vary: Origin
< Date: Thu, 15 Jan 2026 21:56:15 GMT
< Content-Length: 1187
<
{ [1187 bytes data]

100  1187  100  1187    0     0  1074k      0 --:--:-- --:--:-- --:--:-- 1159k
* Connection #0 to host 127.0.0.1 left intact
[
  {
    "id": "15fde994-9cde-4c82-a9e5-6fea65da1bc6",
    "name": "simple-5s-interval",
    "tags": [
      "interval",
      "simple"
    ],
    "nextRun": "2026-01-15T22:56:19+01:00",
    "lastRun": "2026-01-15T22:56:14+01:00",
    "nextRuns": [
      "2026-01-15T22:56:19+01:00",
      "2026-01-15T22:56:24+01:00",
      "2026-01-15T22:56:29+01:00",
      "2026-01-15T22:56:34+01:00",
      "2026-01-15T22:56:39+01:00"
    ],
    "schedule": "Every 5 seconds",
    "scheduleDetail": "Duration: 5s"
  },
  {
    "id": "6996c9b5-8444-4534-a0d4-7f2c7dc09717",
    "name": "simple-10s-interval",
    "tags": [
      "interval",
      "simple"
    ],
    "nextRun": "2026-01-15T22:56:24+01:00",
    "lastRun": "2026-01-15T22:56:14+01:00",
    "nextRuns": [
      "2026-01-15T22:56:24+01:00",
      "2026-01-15T22:56:34+01:00",
      "2026-01-15T22:56:44+01:00",
      "2026-01-15T22:56:54+01:00",
      "2026-01-15T22:57:04+01:00"
    ],
    "schedule": "Every 10 seconds",
    "scheduleDetail": "Duration: 10s"
  },
  {
    "id": "7a5a5965-f56a-4cfe-9abd-5dc91aff5cdd",
    "name": "simple-20s-interval",
    "tags": [
      "interval",
      "simple"
    ],
    "nextRun": "2026-01-15T22:56:24+01:00",
    "lastRun": "2026-01-15T22:56:04+01:00",
    "nextRuns": [
      "2026-01-15T22:56:24+01:00",
      "2026-01-15T22:56:44+01:00",
      "2026-01-15T22:57:04+01:00",
      "2026-01-15T22:57:24+01:00",
      "2026-01-15T22:57:44+01:00"
    ],
    "schedule": "Every 20 seconds",
    "scheduleDetail": "Duration: 20s"
  }
]
