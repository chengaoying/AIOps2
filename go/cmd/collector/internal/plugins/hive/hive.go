package hive

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aiops2/collector/internal/model"
)

type HivePlugin struct {
	hs2Host     string
	hs2Port     int
	database    string
	hookEnabled bool
	hookURL     string
	db         *sql.DB
}

type hiveQuery struct {
	QueryID      string `json:"query_id"`
	QueryText    string `json:"query_text"`
	QueryState   string `json:"state"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	ErrorMessage string `json:"error_message,omitempty"`
	User         string `json:"user"`
	Queue        string `json:"queue_name"`
}

func New(hs2Host string, hs2Port int, database string, hookEnabled bool, hookURL string) *HivePlugin {
	return &HivePlugin{
		hs2Host:     hs2Host,
		hs2Port:     hs2Port,
		database:    database,
		hookEnabled: hookEnabled,
		hookURL:     hookURL,
	}
}

func (p *HivePlugin) Name() string {
	return "hive"
}

func (p *HivePlugin) Init(ctx context.Context, cfg model.PluginConfig) error {
	p.hs2Host = cfg.APIURL
	if port, ok := cfg.Extra["hs2_port"].(float64); ok {
		p.hs2Port = int(port)
	}
	if db, ok := cfg.Extra["database"].(string); ok {
		p.database = db
	}
	if enabled, ok := cfg.Extra["hook_enabled"].(bool); ok {
		p.hookEnabled = enabled
	}
	if url, ok := cfg.Extra["hook_url"].(string); ok {
		p.hookURL = url
	}

	dsn := fmt.Sprintf("hive2://%s:%d/%s", p.hs2Host, p.hs2Port, p.database)
	var err error
	p.db, err = sql.Open("hive", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	p.db.SetMaxOpenConns(5)
	p.db.SetConnMaxLifetime(5 * time.Minute)

	return nil
}

func (p *HivePlugin) Collect(ctx context.Context, queryID string) (*model.JobMeta, error) {
	query := `SELECT query_id, query_text, state, start_time, end_time, error_message, user, queue_name
	          FROM sys.query_data WHERE query_id = ?`

	var hq hiveQuery
	var queryText, errorMsg sql.NullString
	var startTime, endTime sql.NullInt64
	var user, queueName sql.NullString

	err := p.db.QueryRowContext(ctx, query, queryID).Scan(
		&hq.QueryID, &queryText, &hq.QueryState, &startTime, &endTime,
		&errorMsg, &user, &queueName,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	hq.QueryText = queryText.String
	hq.ErrorMessage = errorMsg.String
	if startTime.Valid {
		hq.StartTime = startTime.Int64
	}
	if endTime.Valid {
		hq.EndTime = endTime.Int64
	}
	if user.Valid {
		hq.User = user.String
	}
	if queueName.Valid {
		hq.Queue = queueName.String
	}

	return p.toJobMeta(&hq), nil
}

func (p *HivePlugin) CollectAll(ctx context.Context) ([]*model.JobMeta, error) {
	query := `SELECT query_id, query_text, state, start_time, end_time, error_message, user, queue_name
	          FROM sys.query_data WHERE state IN ('RUNNING', 'FINISHED', 'FAILED', 'ERROR')
	          AND start_time > UNIX_TIMESTAMP() - 86400`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var jobs []*model.JobMeta
	for rows.Next() {
		var hq hiveQuery
		var queryText, errorMsg sql.NullString
		var startTime, endTime sql.NullInt64
		var user, queueName sql.NullString

		if err := rows.Scan(&hq.QueryID, &queryText, &hq.QueryState, &startTime, &endTime,
			&errorMsg, &user, &queueName); err != nil {
			continue
		}

		hq.QueryText = queryText.String
		hq.ErrorMessage = errorMsg.String
		if startTime.Valid {
			hq.StartTime = startTime.Int64
		}
		if endTime.Valid {
			hq.EndTime = endTime.Int64
		}
		if user.Valid {
			hq.User = user.String
		}
		if queueName.Valid {
			hq.Queue = queueName.String
		}

		jobs = append(jobs, p.toJobMeta(&hq))
	}

	return jobs, nil
}

func (p *HivePlugin) Health(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("db not initialized")
	}
	return p.db.PingContext(ctx)
}

func (p *HivePlugin) toJobMeta(hq *hiveQuery) *model.JobMeta {
	status := "RUNNING"
	switch hq.QueryState {
	case "FINISHED":
		status = "SUCCESS"
	case "FAILED", "ERROR":
		status = "FAILED"
	}

	jobName := hq.QueryText
	if len(jobName) > 100 {
		jobName = jobName[:100] + "..."
	}

	startTime := time.UnixMilli(hq.StartTime)
	endTime := time.UnixMilli(hq.EndTime)
	if endTime <= 0 {
		endTime = 0
	}

	return &model.JobMeta{
		JobID:      hq.QueryID,
		Platform:   "HIVE",
		JobName:    jobName,
		Status:     status,
		StartTime:  startTime,
		EndTime:    time.UnixMilli(endTime),
		DurationMs: hq.EndTime - hq.StartTime,
		User:       hq.User,
		Queue:      hq.Queue,
		ErrorMsg:   hq.ErrorMessage,
	}
}

type ErrorType int

const (
	SemanticError ErrorType = iota
	SerdeError
	MemoryError
	PermissionError
)

func ClassifyHiveError(errorMsg string) ErrorType {
	lower := strings.ToLower(errorMsg)
	switch {
	case strings.Contains(lower, "semanticexception"):
		if strings.Contains(lower, "column not found") || strings.Contains(lower, "table not found") {
			return SemanticError
		}
	case strings.Contains(lower, "serdeexception"):
		return SerdeError
	case strings.Contains(lower, "outofmemoryerror") || strings.Contains(lower, "java heap space"):
		return MemoryError
	case strings.Contains(lower, "authorizationexception") || strings.Contains(lower, "permission denied"):
		return PermissionError
	}
	return -1
}
