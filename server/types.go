package server

// JobData represents the job information sent to clients
type JobData struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Tags           []string `json:"tags"`
	NextRun        string   `json:"nextRun"`
	LastRun        string   `json:"lastRun"`
	NextRuns       []string `json:"nextRuns"`
	Schedule       string   `json:"schedule"`       // human-readable schedule description
	ScheduleDetail string   `json:"scheduleDetail"` // technical schedule details (cron expression, interval, etc.)
	SchedulerName  string   `json:"schedulerName"`
}

// CreateJobRequest represents the request to create a new job
type CreateJobRequest struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"` // duration, cron, daily, weekly, monthly
	Interval       int64    `json:"interval,omitempty"`
	CronExpression string   `json:"cronExpression,omitempty"`
	AtTime         string   `json:"atTime,omitempty"` // Format: HH:MM:SS
	Tags           []string `json:"tags,omitempty"`
}
