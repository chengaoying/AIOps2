package nlquery

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type NLQueryParser struct {
	classifier   *IntentClassifier
	extractor   *EntityExtractor
	generator   *SQLGenerator
	db          *sql.DB
}

type NaturalQueryRequest struct {
	Query string `json:"query"`
	User  string `json:"user"`
}

type NaturalQueryResponse struct {
	SQL        string         `json:"sql"`
	Intent     string         `json:"intent"`
	Entities   []string       `json:"entities"`
	Results    []map[string]any `json:"results"`
	ChartType  string         `json:"chart_type"`
	DurationMs int64          `json:"duration_ms"`
}

type SQLQueryRequest struct {
	SQL   string `json:"sql"`
	Limit int    `json:"limit"`
}

type SQLQueryResponse struct {
	Columns []string   `json:"columns"`
	Rows    [][]any    `json:"rows"`
	Count   int        `json:"count"`
}

func New(db *sql.DB) *NLQueryParser {
	return &NLQueryParser{
		classifier: NewIntentClassifier(),
		extractor:   NewEntityExtractor(),
		generator:   NewSQLGenerator(),
		db:          db,
	}
}

func (p *NLQueryParser) ParseNaturalQuery(ctx context.Context, req *NaturalQueryRequest) (*NaturalQueryResponse, error) {
	start := time.Now()

	intent := p.classifier.Classify(req.Query)

	entities := p.extractor.Extract(req.Query)

	sqlQuery := p.generator.Generate(intent, entities)

	chartType := p.generator.GetChartType(intent.Type)

	var results []map[string]any
	if p.db != nil {
		results, err := p.executeQuery(ctx, sqlQuery, 100)
		if err != nil {
			results = []map[string]any{{"error": err.Error()}}
		}
	}

	var entityStrs []string
	if len(entities.Platforms) > 0 {
		entityStrs = append(entityStrs, fmt.Sprintf("platform=%s", entities.Platforms[0]))
	}
	if len(entities.JobIDs) > 0 {
		entityStrs = append(entityStrs, fmt.Sprintf("job_id=%s", entities.JobIDs[0]))
	}
	if entities.TimeRange != nil {
		entityStrs = append(entityStrs, fmt.Sprintf("time_range=%s", entities.TimeRange.Expr))
	}

	return &NaturalQueryResponse{
		SQL:        sqlQuery,
		Intent:     string(intent.Type),
		Entities:   entityStrs,
		Results:    results,
		ChartType:  chartType,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

func (p *NLQueryParser) ExecuteSQL(ctx context.Context, req *SQLQueryRequest) (*SQLQueryResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	rows, err := p.db.QueryContext(ctx, req.SQL)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results [][]any
	for rows.Next() {
		if len(results) >= limit {
			break
		}

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		results = append(results, values)
	}

	return &SQLQueryResponse{
		Columns: columns,
		Rows:    results,
		Count:   len(results),
	}, nil
}

func (p *NLQueryParser) executeQuery(ctx context.Context, sqlQuery string, limit int) ([]map[string]any, error) {
	rows, err := p.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		if len(results) >= limit {
			break
		}

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}
