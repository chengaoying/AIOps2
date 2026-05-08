package yarn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aiops2/collector/internal/model"
)

type YARNPlugin struct {
	apiURL   string
	atsURL   string
	user     string
	interval time.Duration
	client  *http.Client
}

type yarnApp struct {
	AppID          string `json:"id"`
	AppName       string `json:"name"`
	AppState      string `json:"state"`
	AppType       string `json:"type"`
	StartedTime   int64  `json:"startedTime"`
	FinishedTime  int64  `json:"finishedTime"`
	ElapsedTime   int64  `json:"elapsedTime"`
	Queue         string `json:"queue"`
	User          string `json:"user"`
	ExitCode      int    `json:"exitCode"`
	Diagnostics   string `json:"diagnostics,omitempty"`
	FinalStatus   string `json:"finalStatus,omitempty"`
}

type yarnAppsResponse struct {
	Apps struct {
		App []yarnApp `json:"app"`
	} `json:"apps"`
}

func New(apiURL, atsURL, user string, interval time.Duration) *YARNPlugin {
	return &YARNPlugin{
		apiURL:   apiURL,
		atsURL:   atsURL,
		user:     user,
		interval: interval,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *YARNPlugin) Name() string {
	return "yarn"
}

func (p *YARNPlugin) Init(ctx context.Context, cfg model.PluginConfig) error {
	p.apiURL = cfg.APIURL
	if user, ok := cfg.Extra["user"].(string); ok {
		p.user = user
	}
	return nil
}

func (p *YARNPlugin) Collect(ctx context.Context, appID string) (*model.JobMeta, error) {
	url := fmt.Sprintf("%s/ws/v1/cluster/apps/%s", p.apiURL, appID)

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

	var app yarnApp
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.toJobMeta(&app), nil
}

func (p *YARNPlugin) CollectAll(ctx context.Context) ([]*model.JobMeta, error) {
	url := fmt.Sprintf("%s/ws/v1/cluster/apps?states=RUNNING,ACCEPTED,SUCCEEDED,FAILED,KILLED", p.apiURL)

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

	var appsResp yarnAppsResponse
	if err := json.NewDecoder(resp.Body).Decode(&appsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	jobs := make([]*model.JobMeta, 0, len(appsResp.Apps.App))
	for _, app := range appsResp.Apps.App {
		jobs = append(jobs, p.toJobMeta(&app))
	}

	return jobs, nil
}

func (p *YARNPlugin) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/ws/v1/cluster/info", p.apiURL)

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

func (p *YARNPlugin) toJobMeta(app *yarnApp) *model.JobMeta {
	status := p.mapState(app.AppState, app.FinalStatus)
	startTime := time.UnixMilli(app.StartedTime)
	endTime := time.UnixMilli(app.FinishedTime)

	if endTime.Before(time.Unix(0, 0)) {
		endTime = time.Time{}
	}

	return &model.JobMeta{
		JobID:      app.AppID,
		Platform:   "YARN",
		JobName:    app.AppName,
		Status:     status,
		StartTime:  startTime,
		EndTime:    endTime,
		DurationMs: app.ElapsedTime,
		User:       app.User,
		Queue:      app.Queue,
		ExitCode:   app.ExitCode,
		ErrorMsg:   app.Diagnostics,
	}
}

func (p *YARNPlugin) mapState(state, finalStatus string) string {
	switch state {
	case "RUNNING", "ACCEPTED":
		return "RUNNING"
	case "SUCCEEDED":
		return "SUCCESS"
	case "FAILED":
		if finalStatus == "KILLED" || strings.Contains(strings.ToLower(state), "kill") {
			return "KILLED"
		}
		return "FAILED"
	default:
		return state
	}
}

func DetectYARNOOM(diagnostics string) bool {
	return strings.Contains(diagnostics, "Container killed") &&
		strings.Contains(diagnostics, "out of memory")
}
