# GoCronUI Basic Auth Example

This is a simple example demonstrating securing gocron-ui with basic auth:

```go
> GOCRON_UI_USERNAME=admin GOCRON_UI_PASSWORD=password go run main.go
```

In another terminal:

**valid credentials**
```bash
curl -v -u "admin:password" 127.0.0.1:8080/api/jobs | jq .
\  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080
* Server auth using Basic with user 'admin'
> GET /api/jobs HTTP/1.1
> Host: 127.0.0.1:8080
> Authorization: Basic YWRtaW46cGFzc3dvcmQ=
> User-Agent: curl/8.7.1
> Accept: */*
> 
* Request completely sent off
< HTTP/1.1 200 OK
< Content-Type: application/json
< Vary: Origin
< Date: Sun, 09 Nov 2025 23:09:13 GMT
< Content-Length: 1187
< 
{ [1187 bytes data]
100  1187  100  1187    0     0  1837k      0 --:--:-- --:--:-- --:--:-- 1159k
* Connection #0 to host 127.0.0.1 left intact
[
  {
    "id": "3681db0c-fd30-48ed-af21-2082f6a9f3d7",
    "name": "simple-10s-interval",
    "tags": [
      "interval",
      "simple"
    ],
    "nextRun": "2025-11-10T10:09:07+11:00",
    "lastRun": "2025-11-10T10:09:07+11:00",
    "nextRuns": [
      "2025-11-10T10:09:07+11:00",
      "2025-11-10T10:09:17+11:00",
      "2025-11-10T10:09:27+11:00",
      "2025-11-10T10:09:37+11:00",
      "2025-11-10T10:09:47+11:00"
    ],
    "schedule": "Every 10 seconds",
    "scheduleDetail": "Duration: 10s"
  },
  {
    "id": "7ff3624c-248e-47f6-90bd-49da5451ddcf",
    "name": "simple-20s-interval",
    "tags": [
      "interval",
      "simple"
    ],
    "nextRun": "2025-11-10T10:09:07+11:00",
    "lastRun": "2025-11-10T10:09:07+11:00",
    "nextRuns": [
      "2025-11-10T10:09:07+11:00",
      "2025-11-10T10:09:27+11:00",
      "2025-11-10T10:09:47+11:00",
      "2025-11-10T10:10:07+11:00",
      "2025-11-10T10:10:27+11:00"
    ],
    "schedule": "Every 20 seconds",
    "scheduleDetail": "Duration: 20s"
  },
  {
    "id": "ad6ddb39-e1a3-4dc4-8539-cf1d9a4667eb",
    "name": "simple-5s-interval",
    "tags": [
      "interval",
      "simple"
    ],
    "nextRun": "2025-11-10T10:09:17+11:00",
    "lastRun": "2025-11-10T10:09:12+11:00",
    "nextRuns": [
      "2025-11-10T10:09:17+11:00",
      "2025-11-10T10:09:22+11:00",
      "2025-11-10T10:09:27+11:00",
      "2025-11-10T10:09:32+11:00",
      "2025-11-10T10:09:37+11:00"
    ],
    "schedule": "Every 5 seconds",
    "scheduleDetail": "Duration: 5s"
  }
]
```

**missing or invalid credentials**
```bash
curl -v 127.0.0.1:8080/api/jobs       
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080
> GET /api/jobs HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/8.7.1
> Accept: */*
> 
* Request completely sent off
< HTTP/1.1 401 Unauthorized
< Content-Type: text/plain; charset=utf-8
< Www-Authenticate: Basic realm="restricted", charset="UTF-8"
< X-Content-Type-Options: nosniff
< Date: Sun, 09 Nov 2025 23:10:02 GMT
< Content-Length: 13
< 
Unauthorized
* Connection #0 to host 127.0.0.1 left intact
```