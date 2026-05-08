package spark

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aiops2/collector/internal/model"
)

type SparkPlugin struct {
	historyServer string
	livyURL       string
	user          string
	interval      time.Duration
	client        *http.Client
}

type sparkApp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Duration    int64  `json:"duration"`
	SparkUser   string `json:"sparkUser"`
	AppType     string `json:"appType"`
}

type sparkAppsResponse []sparkApp

type sparkExecutor struct {
	ID           string `json:"id"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	MemoryUsed   int64  `json:"memoryUsed"`
	MemoryTotal  int64  `json:"memoryTotal"`
	TasksFailed  int    `json:"tasksFailed"`
	LastUpdate   int64  `json:"lastUpdate"`
}

type sparkStage struct {
	StageID    int    `json:"stageId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	NumTasks   int    `json:"numTasks"`
	FailedTasks int   `json:"failedTasks"`
}

func New(historyServer, livyURL, user string, interval time.Duration) *SparkPlugin {
	return &SparkPlugin{
		historyServer: historyServer,
		livyURL:       livyURL,
		user:          user,
		interval:      interval,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *SparkPlugin) Name() string {
	return "spark"
}

func (p *SparkPlugin) Init(ctx context.Context, cfg model.PluginConfig) error {
	p.historyServer = cfg.APIURL
	if url, ok := cfg.Extra["livy_url"].(string); ok {
		p.livyURL = url
	}
	return nil
}

func (p *SparkPlugin) Collect(ctx context.Context, appID string) (*model.JobMeta, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s", p.historyServer, appID)

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

	var app sparkApp
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.toJobMeta(&app), nil
}

func (p *SparkPlugin) CollectAll(ctx context.Context) ([]*model.JobMeta, error) {
	url := fmt.Sprintf("%s/api/v1/applications", p.historyServer)

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

	var apps sparkAppsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	jobs := make([]*model.JobMeta, 0, len(apps))
	for _, app := range apps {
		jobs = append(jobs, p.toJobMeta(&app))
	}

	return jobs, nil
}

func (p *SparkPlugin) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/applications", p.historyServer)

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

func (p *SparkPlugin) toJobMeta(app *sparkApp) *model.JobMeta {
	status := p.mapState(app.State)
	startTime := time.UnixMilli(app.StartTime)
	endTime := time.UnixMilli(app.EndTime)

	if endTime <= 0 {
		endTime = 0
	}

	return &model.JobMeta{
		JobID:      app.ID,
		Platform:   "SPARK",
		JobName:    app.Name,
		Status:     status,
		StartTime:  startTime,
		EndTime:    time.UnixMilli(endTime),
		DurationMs: app.Duration,
		User:       app.SparkUser,
	}
}

func (p *SparkPlugin) mapState(state string) string {
	switch state {
	case "RUNNING":
		return "RUNNING"
	case "SUCCEEDED":
		return "SUCCESS"
	case "FAILED":
		return "FAILED"
	case "KILLED":
		return "KILLED"
	default:
		return state
	}
}

func GetSparkExecutors(ctx context.Context, historyServer, appID string) ([]sparkExecutor, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/executors", historyServer, appID)

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

	var executors []sparkExecutor
	if err := json.NewDecoder(resp.Body).Decode(&executors); err != nil {
		return nil, err
	}

	return executors, nil
}

func GetSparkStages(ctx context.Context, historyServer, appID string) ([]sparkStage, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/stages", historyServer, appID)

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

	var stages []sparkStage
	if err := json.NewDecoder(resp.Body).Decode(&stages); err != nil {
		return nil, err
	}

	return stages, nil
}

func DetectSparkOOM(executorLogs string) bool {
	lower := strings.ToLower(executorLogs)
	return strings.Contains(lower, "outofmemoryerror") ||
		(strings.Contains(lower, "executorlost") && strings.Contains(lower, "memory"))
}

func DetectShuffleError(logs string) bool {
	return strings.Contains(strings.ToLower(logs), "shufflefetchfailed")
}
