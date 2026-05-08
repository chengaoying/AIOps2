package nlquery

import (
	"fmt"
	"strings"
	"time"
)

type SQLGenerator struct {
	templates map[IntentType]string
}

func NewSQLGenerator() *SQLGenerator {
	return &SQLGenerator{
		templates: map[IntentType]string{
			IntentPerformanceAnalysis: `
SELECT
    job_id,
    platform,
    job_name,
    duration_ms,
    start_time,
    end_time
FROM unified_job_view
WHERE {{conditions}}
ORDER BY start_time DESC
LIMIT 100`,

			IntentResourceAnalysis: `
SELECT
    job_id,
    platform,
    memory_used_mb,
    cpu_used_cores,
    duration_ms
FROM job_metrics
WHERE {{conditions}}
ORDER BY memory_used_mb DESC
LIMIT 20`,

			IntentFailureAnalysis: `
SELECT
    job_id,
    platform,
    job_name,
    status,
    error_msg,
    exit_code,
    end_time
FROM unified_job_view
WHERE status = 'FAILED'
    {{conditions}}
ORDER BY end_time DESC
LIMIT 50`,

			IntentTrendAnalysis: `
SELECT
    DATE(start_time) as date,
    COUNT(*) as job_count,
    AVG(duration_ms) as avg_duration,
    MAX(duration_ms) as max_duration,
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failure_count
FROM unified_job_view
WHERE {{conditions}}
GROUP BY DATE(start_time)
ORDER BY date DESC`,

			IntentJobQuery: `
SELECT
    job_id,
    platform,
    job_name,
    status,
    start_time,
    end_time,
    duration_ms
FROM unified_job_view
WHERE {{conditions}}
ORDER BY start_time DESC
LIMIT 50`,

			IntentMetricsQuery: `
SELECT
    job_id,
    platform,
    metric_name,
    metric_value,
    timestamp
FROM job_metrics
WHERE {{conditions}}
ORDER BY timestamp DESC
LIMIT 100`,
		},
	}
}

func (g *SQLGenerator) Generate(intent *Intent, entities *ExtractedEntities) string {
	template, ok := g.templates[intent.Type]
	if !ok {
		template = g.templates[IntentPerformanceAnalysis]
	}

	conditions := g.buildConditions(intent, entities)
	sql := strings.Replace(template, "{{conditions}}", conditions, 1)

	return strings.TrimSpace(sql)
}

func (g *SQLGenerator) buildConditions(intent *Intent, entities *ExtractedEntities) string {
	var conditions []string

	if len(entities.Platforms) > 0 {
		platform := entities.Platforms[0]
		conditions = append(conditions, fmt.Sprintf("platform = '%s'", strings.ToUpper(platform)))
	}

	if len(entities.JobIDs) > 0 {
		jobID := entities.JobIDs[0]
		conditions = append(conditions, fmt.Sprintf("job_id = '%s'", jobID))
	}

	if len(entities.Users) > 0 {
		user := entities.Users[0]
		conditions = append(conditions, fmt.Sprintf("user_name = '%s'", user))
	}

	if entities.TimeRange != nil {
		startStr := entities.TimeRange.Start.Format("2006-01-02 15:04:05")
		endStr := entities.TimeRange.End.Format("2006-01-02 15:04:05")
		conditions = append(conditions, fmt.Sprintf("start_time >= '%s'", startStr))
		conditions = append(conditions, fmt.Sprintf("start_time <= '%s'", endStr))
	}

	if len(entities.Metrics) > 0 {
		metric := entities.Metrics[0]
		switch metric {
		case "memory_used_mb":
			conditions = append(conditions, "memory_used_mb > 0")
		case "cpu_used_cores":
			conditions = append(conditions, "cpu_used_cores > 0")
		case "duration_ms":
			conditions = append(conditions, "duration_ms > 0")
		}
	}

	if len(conditions) == 0 {
		start := time.Now().Add(-24 * time.Hour)
		end := time.Now()
		conditions = append(conditions, fmt.Sprintf("start_time >= '%s'", start.Format("2006-01-02 15:04:05")))
		conditions = append(conditions, fmt.Sprintf("start_time <= '%s'", end.Format("2006-01-02 15:04:05")))
	}

	return strings.Join(conditions, "\n    AND ")
}

func (g *SQLGenerator) GetChartType(intent IntentType) string {
	switch intent {
	case IntentPerformanceAnalysis, IntentTrendAnalysis:
		return "line"
	case IntentResourceAnalysis:
		return "bar"
	case IntentFailureAnalysis, IntentJobQuery, IntentMetricsQuery:
		return "table"
	default:
		return "line"
	}
}
