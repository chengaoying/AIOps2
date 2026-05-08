package model

import "time"

type JobMeta struct {
	JobID          string            `json:"job_id"`
	Platform       string            `json:"platform"`
	JobName        string            `json:"job_name"`
	Status         string            `json:"status"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        time.Time         `json:"end_time"`
	DurationMs     int64             `json:"duration_ms"`
	SubmitTime     time.Time         `json:"submit_time"`
	Priority       string            `json:"priority"`
	User           string            `json:"user"`
	Queue          string            `json:"queue"`
	ExitCode       int               `json:"exit_code"`
	ErrorMsg       string            `json:"error_msg"`
	Logs           []string          `json:"logs"`
	Metrics        map[string]float64 `json:"metrics"`
	DependencyJobIDs []string        `json:"dependency_job_ids"`
	RawData        map[string]any    `json:"raw_data"`
}

type PluginConfig struct {
	Enabled  bool                   `json:"enabled"`
	APIURL   string                `json:"api_url"`
	Interval time.Duration          `json:"interval"`
	Extra    map[string]any        `json:"extra,omitempty"`
}
