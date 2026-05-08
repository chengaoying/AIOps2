package context

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aiops2/diagnosis-api/internal/engine"
)

type ContextBuilder struct {
	db *sql.DB
}

func New(db *sql.DB) *ContextBuilder {
	return &ContextBuilder{db: db}
}

func (b *ContextBuilder) Build(ctx context.Context, req *engine.DiagnosisRequest) (*engine.DiagnosisContext, error) {
	job, err := b.getJobDetails(ctx, req.JobID)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	jobChain, _ := b.getJobChain(ctx, job)
	similarCases, _ := b.findSimilarCases(ctx, job)
	errorPatterns := b.extractErrorPatterns(job.ErrorMsg)

	return &engine.DiagnosisContext{
		Job:          job,
		JobChain:     jobChain,
		SimilarCases: similarCases,
		RelatedLogs:  []string{},
		Metrics:      make(map[string]any),
	}, nil
}

func (b *ContextBuilder) getJobDetails(ctx context.Context, jobID string) (*engine.JobMeta, error) {
	query := `SELECT job_id, platform, job_name, status, start_time, end_time,
	          duration_ms, user, queue, error_msg, exit_code
	          FROM job_meta WHERE job_id = ?`

	var job engine.JobMeta
	var startTime, endTime sql.NullTime
	var errorMsg sql.NullString

	err := b.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.JobID, &job.Platform, &job.JobName, &job.Status,
		&startTime, &endTime, &job.DurationMs,
		&job.User, &job.Queue, &errorMsg, &job.ExitCode,
	)
	if err != nil {
		return nil, err
	}

	if startTime.Valid {
		job.StartTime = startTime.Time
	}
	if endTime.Valid {
		job.EndTime = endTime.Time
	}
	if errorMsg.Valid {
		job.ErrorMsg = errorMsg.String
	}

	return &job, nil
}

func (b *ContextBuilder) getJobChain(ctx context.Context, job *engine.JobMeta) ([]*engine.JobMeta, error) {
	query := `SELECT j.job_id, j.platform, j.job_name, j.status, j.start_time,
	          j.end_time, j.duration_ms, j.user, j.queue
	          FROM job_meta j
	          INNER JOIN job_dependency d ON j.job_id = d.dependency_job_id
	          WHERE d.job_id = ? AND j.start_time BETWEEN ? AND ?`

	rows, err := b.db.QueryContext(ctx, query, job.JobID,
		job.StartTime.Add(-1*time.Hour), job.EndTime.Add(1*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chain []*engine.JobMeta
	for rows.Next() {
		var j engine.JobMeta
		var startTime, endTime sql.NullTime
		if err := rows.Scan(&j.JobID, &j.Platform, &j.JobName, &j.Status,
			&startTime, &endTime, &j.DurationMs, &j.User, &j.Queue); err != nil {
			continue
		}
		if startTime.Valid {
			j.StartTime = startTime.Time
		}
		if endTime.Valid {
			j.EndTime = endTime.Time
		}
		chain = append(chain, &j)
	}

	return chain, nil
}

func (b *ContextBuilder) findSimilarCases(ctx context.Context, job *engine.JobMeta) ([]*engine.JobMeta, error) {
	query := `SELECT job_id, platform, job_name, status, start_time, end_time,
	          duration_ms, user, error_msg
	          FROM job_meta
	          WHERE platform = ? AND status = 'FAILED'
	          AND error_msg LIKE ?
	          ORDER BY start_time DESC LIMIT 5`

	pattern := "%" + extractKeyError(job.ErrorMsg) + "%"
	rows, err := b.db.QueryContext(ctx, query, job.Platform, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []*engine.JobMeta
	for rows.Next() {
		var j engine.JobMeta
		var startTime, endTime sql.NullTime
		var errorMsg sql.NullString
		if err := rows.Scan(&j.JobID, &j.Platform, &j.JobName, &j.Status,
			&startTime, &endTime, &j.DurationMs, &j.User, &errorMsg); err != nil {
			continue
		}
		if startTime.Valid {
			j.StartTime = startTime.Time
		}
		if endTime.Valid {
			j.EndTime = endTime.Time
		}
		if errorMsg.Valid {
			j.ErrorMsg = errorMsg.String
		}
		cases = append(cases, &j)
	}

	return cases, nil
}

func (b *ContextBuilder) extractErrorPatterns(errorMsg string) []string {
	if errorMsg == "" {
		return []string{}
	}

	patterns := []string{errorMsg}

	keywords := []string{
		"OutOfMemory", "OOM", "memory",
		"Shuffle", "shuffle",
		"Timeout", "timeout",
		"Connection", "connection",
		"Auth", "auth",
	}

	for _, kw := range keywords {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(kw)) {
			patterns = append(patterns, kw)
		}
	}

	return patterns
}

func extractKeyError(msg string) string {
	if msg == "" {
		return ""
	}
	lines := strings.Split(msg, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return msg
}
