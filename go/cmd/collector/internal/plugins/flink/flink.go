package flink

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aiops2/collector/internal/model"
)

type FlinkPlugin struct {
	restURL        string
	metricsEnabled bool
	interval       time.Duration
	client         *http.Client
}

type flinkJob struct {
	JobID        string `json:"id"`
	JobName      string `json:"name"`
	JobType      string `json:"type"`
	Status       string `json:"status"`
	StartTime    int64  `json:"start-time"`
	EndTime      int64  `json:"end-time"`
	Duration     int64  `json:"duration"`
	LastModTime  int64  `json:"last-modification-time"`
}

type flinkJobsResponse struct {
	Jobs []flinkJob `json:"jobs"`
}

type flinkCheckpoint struct {
	CheckpointID         int64  `json:"id"`
	JobID                string `json:"job_id"`
	Status               string `json:"status"`
	Type                 string `json:"type"`
	TriggerTime          int64  `json:"trigger_time"`
	LatestAckTimestamp    int64  `json:"latest_ack_timestamp"`
	StateSize            int64  `json:"state_size"`
	EndToEndDuration     int64  `json:"end_to_end_duration"`
	AlignmentBuffered    int64  `json:"alignment_buffered"`
	ProcessedData        int64  `json:"processed_data"`
	PersistedData        int64  `json:"persisted_data"`
	NumSubtasks          int    `json:"num_subtasks"`
	NumAcknowledgedSubtasks int `json:"num_acknowledged_subtasks"`
}

type flinkTaskManager struct {
	TaskManagerID string `json:"id"`
	TaskManagerHost string `json:"taskManagerHost"`
	NumberOfSlots  int    `json:"numberOfSlots"`
	AvailableSlots  int    `json:"availableSlots"`
	MemoryUsed     int64  `json:"heapMemoryUsed"`
	MemoryTotal    int64  `json:"heapMemory"`
}

func New(restURL string, metricsEnabled bool, interval time.Duration) *FlinkPlugin {
	return &FlinkPlugin{
		restURL:        restURL,
		metricsEnabled:  metricsEnabled,
		interval:       interval,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *FlinkPlugin) Name() string {
	return "flink"
}

func (p *FlinkPlugin) Init(ctx context.Context, cfg model.PluginConfig) error {
	p.restURL = cfg.APIURL
	if enabled, ok := cfg.Extra["metrics_enabled"].(bool); ok {
		p.metricsEnabled = enabled
	}
	return nil
}

func (p *FlinkPlugin) Collect(ctx context.Context, jobID string) (*model.JobMeta, error) {
	url := fmt.Sprintf("%s/jobs/%s", p.restURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var job flinkJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.toJobMeta(&job), nil
}

func (p *FlinkPlugin) CollectAll(ctx context.Context) ([]*model.JobMeta, error) {
	url := fmt.Sprintf("%s/jobs", p.restURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var jobsResp flinkJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	jobs := make([]*model.JobMeta, 0, len(jobsResp.Jobs))
	for _, job := range jobsResp.Jobs {
		jobs = append(jobs, p.toJobMeta(&job))
	}

	return jobs, nil
}

func (p *FlinkPlugin) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/overview", p.restURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (p *FlinkPlugin) toJobMeta(job *flinkJob) *model.JobMeta {
	status := p.mapState(job.Status)
	startTime := time.UnixMilli(job.StartTime)
	endTime := time.UnixMilli(job.EndTime)

	if endTime <= 0 {
		endTime = 0
	}

	return &model.JobMeta{
		JobID:      job.JobID,
		Platform:   "FLINK",
		JobName:    job.JobName,
		Status:     status,
		StartTime:  startTime,
		EndTime:    time.UnixMilli(endTime),
		DurationMs: job.Duration,
	}
}

func (p *FlinkPlugin) mapState(state string) string {
	switch state {
	case "CREATED", "INITIALIZING", "RUNNING", "RECONCILING", "FLAGGING":
		return "RUNNING"
	case "FAILED":
		return "FAILED"
	case "CANCELED", "CANCELLING":
		return "KILLED"
	case "FINISHED", "SUSPENDED":
		return "SUCCESS"
	default:
		return state
	}
}

func GetFlinkCheckpoints(ctx context.Context, restURL, jobID string) ([]flinkCheckpoint, error) {
	url := fmt.Sprintf("%s/jobs/%s/checkpoints", restURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", resp.StatusCode)
	}

	var checkpoints struct {
		History []flinkCheckpoint `json:"history"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&checkpoints); err != nil {
		return nil, err
	}

	return checkpoints.History, nil
}

func GetFlinkTaskManagers(ctx context.Context, restURL string) ([]flinkTaskManager, error) {
	url := fmt.Sprintf("%s/taskmanagers", restURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", resp.StatusCode)
	}

	var tmResponse struct {
		TaskManagers []flinkTaskManager `json:"taskmanagers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tmResponse); err != nil {
		return nil, err
	}

	return tmResponse.TaskManagers, nil
}

type FlinkErrorType int

const (
	CheckpointTimeout FlinkErrorType = iota
	TMMemory
	KafkaTimeout
	JobCancelled
	TaskFailed
)

func ClassifyFlinkError(errorMsg string) FlinkErrorType {
	lower := strings.ToLower(errorMsg)
	switch {
	case strings.Contains(lower, "checkpoint timeout") || strings.Contains(lower, "checkpoint timed out"):
		return CheckpointTimeout
	case strings.Contains(lower, "taskmanager memory") || strings.Contains(lower, "tm memory"):
		return TMMemory
	case strings.Contains(lower, "kafka timeout"):
		return KafkaTimeout
	case strings.Contains(lower, "job cancelled") || strings.Contains(lower, "job canceled"):
		return JobCancelled
	case strings.Contains(lower, "task execution failed"):
		return TaskFailed
	}
	return -1
}
