package writer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"aiops2/collector/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

type BatchWriter struct {
	db             *sql.DB
	batchSize      int
	flushInterval  time.Duration
	buffer         []*model.JobMeta
	mu             sync.Mutex
	stopCh         chan struct{}
	flushTimer     *time.Timer
}

func New(dsn string, batchSize int, flushInterval time.Duration) (*BatchWriter, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	w := &BatchWriter{
		db:            db,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]*model.JobMeta, 0, batchSize),
		stopCh:        make(chan struct{}),
	}

	go w.runFlushLoop()
	return w, nil
}

func (w *BatchWriter) Write(job *model.JobMeta) error {
	w.mu.Lock()
	w.buffer = append(w.buffer, job)
	shouldFlush := len(w.buffer) >= w.batchSize
	w.mu.Unlock()

	if shouldFlush {
		return w.Flush()
	}
	return nil
}

func (w *BatchWriter) Flush() error {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return nil
	}

	jobs := w.buffer
	w.buffer = make([]*model.JobMeta, 0, w.batchSize)
	w.mu.Unlock()

	return w.batchInsert(jobs)
}

func (w *BatchWriter) batchInsert(jobs []*model.JobMeta) error {
	if len(jobs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO job_meta (
			job_id, platform, job_name, status, start_time, end_time,
			duration_ms, submit_time, priority, user, queue, exit_code,
			error_msg, logs, metrics, dependency_job_ids, raw_data, create_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			platform=VALUES(platform), status=VALUES(status), end_time=VALUES(end_time),
			duration_ms=VALUES(duration_ms), error_msg=VALUES(error_msg)
	`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		_, err := stmt.ExecContext(ctx,
			job.JobID, job.Platform, job.JobName, job.Status,
			job.StartTime, job.EndTime, job.DurationMs, job.SubmitTime,
			job.Priority, job.User, job.Queue, job.ExitCode,
			job.ErrorMsg, joinLogs(job.Logs), formatMetrics(job.Metrics),
			joinStrings(job.DependencyJobIDs), formatRawData(job.RawData), time.Now(),
		)
		if err != nil {
			log.Printf("insert job %s failed: %v", job.JobID, err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("BatchWriter: flushed %d jobs to StarRocks", len(jobs))
	return nil
}

func (w *BatchWriter) runFlushLoop() {
	for {
		select {
		case <-w.stopCh:
			w.Flush()
			return
		case <-time.After(w.flushInterval):
			w.Flush()
		}
	}
}

func (w *BatchWriter) Stop() {
	close(w.stopCh)
}

func joinLogs(logs []string) string {
	if len(logs) == 0 {
		return ""
	}
	result := ""
	for i, l := range logs {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func formatMetrics(m map[string]float64) string {
	if m == nil {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += fmt.Sprintf("\"%s\":%f", k, v)
		first = false
	}
	result += "}"
	return result
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

func formatRawData(m map[string]any) string {
	if m == nil {
		return "{}"
	}
	return "{}"
}
