package baseline

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type BaselineService struct {
	db *sql.DB
}

type JobBaseline struct {
	Platform  string    `json:"platform"`
	JobName   string    `json:"job_name"`
	P50Ms     int64     `json:"p50_ms"`
	P95Ms     int64     `json:"p95_ms"`
	Count     int       `json:"count"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewBaselineService(db *sql.DB) *BaselineService {
	return &BaselineService{db: db}
}

func (s *BaselineService) Recalculate(ctx context.Context) error {
	query := `
INSERT INTO job_baseline (platform, job_name, p50_duration_ms, p95_duration_ms, sample_count, updated_at)
SELECT
    platform,
    job_name,
    PERCENTILE_CONT(duration_ms, 0.5) as p50,
    PERCENTILE_CONT(duration_ms, 0.95) as p95,
    COUNT(*) as cnt,
    NOW() as updated_at
FROM job_meta
WHERE start_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)
    AND status = 'SUCCESS'
GROUP BY platform, job_name
ON DUPLICATE KEY UPDATE
    p50_duration_ms = VALUES(p50_duration_ms),
    p95_duration_ms = VALUES(p95_duration_ms),
    sample_count = VALUES(sample_count),
    updated_at = VALUES(updated_at)
`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *BaselineService) GetBaseline(ctx context.Context, platform, jobName string) (*JobBaseline, error) {
	query := `
SELECT platform, job_name, p50_duration_ms, p95_duration_ms, sample_count, updated_at
FROM job_baseline
WHERE platform = ? AND job_name = ?
`
	row := s.db.QueryRowContext(ctx, query, platform, jobName)

	var baseline JobBaseline
	err := row.Scan(
		&baseline.Platform,
		&baseline.JobName,
		&baseline.P50Ms,
		&baseline.P95Ms,
		&baseline.Count,
		&baseline.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("baseline not found for %s/%s", platform, jobName)
	}
	if err != nil {
		return nil, err
	}
	return &baseline, nil
}

func (s *BaselineService) GetAllBaselines(ctx context.Context) ([]*JobBaseline, error) {
	query := `
SELECT platform, job_name, p50_duration_ms, p95_duration_ms, sample_count, updated_at
FROM job_baseline
ORDER BY platform, job_name
`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var baselines []*JobBaseline
	for rows.Next() {
		var b JobBaseline
		if err := rows.Scan(
			&b.Platform,
			&b.JobName,
			&b.P50Ms,
			&b.P95Ms,
			&b.Count,
			&b.UpdatedAt,
		); err != nil {
			continue
		}
		baselines = append(baselines, &b)
	}
	return baselines, nil
}

func (s *BaselineService) StartHourlyJob() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err := s.Recalculate(ctx); err != nil {
				fmt.Printf("baseline recalculate failed: %v\n", err)
			}
			cancel()
		}
	}()
}
