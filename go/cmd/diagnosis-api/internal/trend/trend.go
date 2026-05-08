package trend

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TrendEngine struct {
	db *sql.DB
}

type TrendResult struct {
	Metric     string  `json:"metric"`
	IsAnomaly  bool    `json:"is_anomaly"`
	Message    string  `json:"message"`
	Current    float64 `json:"current"`
	Baseline   float64 `json:"baseline"`
	ChangeRate float64 `json:"change_rate"`
}

func NewTrendEngine(db *sql.DB) *TrendEngine {
	return &TrendEngine{db: db}
}

func (e *TrendEngine) DetectPerformanceDegradation(ctx context.Context, platform, jobName string, currentMs int64) (*TrendResult, error) {
	query := `
SELECT p50_duration_ms, p95_duration_ms
FROM job_baseline
WHERE platform = ? AND job_name = ?
`
	var p50, p95 int64
	err := e.db.QueryRowContext(ctx, query, platform, jobName).Scan(&p50, &p95)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	baseline := float64(p95)
	result := &TrendResult{
		Metric:     "duration_ms",
		Baseline:   baseline,
		Current:    float64(currentMs),
		ChangeRate: float64(currentMs) / baseline,
	}

	if float64(currentMs) > baseline*1.5 {
		result.IsAnomaly = true
		result.Message = fmt.Sprintf("性能退化检测: 当前耗时 %.0fms 超过baseline*1.5 (%.0fms)", float64(currentMs), baseline*1.5)
	}

	return result, nil
}

func (e *TrendEngine) DetectResourceAnomaly(ctx context.Context, platform, jobName string, metricName string, currentValue float64) (*TrendResult, error) {
	query := fmt.Sprintf(`
SELECT AVG(%s), STDDEV(%s)
FROM job_metrics
WHERE platform = ? AND job_name = ?
    AND timestamp >= DATE_SUB(NOW(), INTERVAL 7 DAY)
`, metricName, metricName)

	var avg, stddev sql.NullFloat64
	err := e.db.QueryRowContext(ctx, query, platform, jobName).Scan(&avg, &stddev)
	if err == sql.ErrNoRows || !avg.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result := &TrendResult{
		Metric:     metricName,
		Baseline:   avg.Float64,
		Current:    currentValue,
		ChangeRate: currentValue / avg.Float64,
	}

	if stddev.Valid && currentValue > avg.Float64+2*stddev.Float64 {
		result.IsAnomaly = true
		result.Message = fmt.Sprintf("资源异常检测: 当前值 %.2f 超过 avg+2*stddev (%.2f)", currentValue, avg.Float64+2*stddev.Float64)
	}

	return result, nil
}

func (e *TrendEngine) GetTrendData(ctx context.Context, platform, jobName string, days int) ([]map[string]any, error) {
	query := `
SELECT
    DATE(start_time) as date,
    AVG(duration_ms) as avg_duration,
    MAX(duration_ms) as max_duration,
    COUNT(*) as job_count,
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failure_count
FROM job_meta
WHERE platform = ? AND job_name = ?
    AND start_time >= DATE_SUB(NOW(), INTERVAL ? DAY)
GROUP BY DATE(start_time)
ORDER BY date DESC
`
	rows, err := e.db.QueryContext(ctx, query, platform, jobName, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var date time.Time
		var avgDur, maxDur float64
		var jobCount, failCount int

		if err := rows.Scan(&date, &avgDur, &maxDur, &jobCount, &failCount); err != nil {
			continue
		}

		results = append(results, map[string]any{
			"date":          date.Format("2006-01-02"),
			"avg_duration":  avgDur,
			"max_duration":  maxDur,
			"job_count":     jobCount,
			"failure_count": failCount,
		})
	}
	return results, nil
}
