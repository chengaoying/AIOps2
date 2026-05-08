package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"aiops2/collector/internal/config"
	"aiops2/collector/internal/model"
	"aiops2/collector/internal/queue"
	"aiops2/collector/internal/registry"
	"aiops2/collector/internal/wal"
	"aiops2/collector/internal/writer"
)

type CollectorApp struct {
	cfg      *config.Config
	registry *registry.Registry
	queue    *queue.MemoryQueue
	wal      *wal.WAL
	writer   *writer.BatchWriter
}

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("NewApp failed: %v", err)
	}
	defer app.Stop()

	r := gin.Default()

	r.GET("/health", app.handleHealth)
	r.POST("/api/v1/collect", app.handleCollect)
	r.GET("/api/v1/stats", app.handleStats)

	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	log.Printf("AIOps Collector starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start collector: %v", err)
	}
}

func NewApp(cfg *config.Config) (*CollectorApp, error) {
	app := &CollectorApp{
		cfg:      cfg,
		registry: registry.New(),
		queue:    queue.New(10000),
	}

	w, err := wal.New(cfg.WAL.Dir, cfg.WAL.MaxFileSize)
	if err != nil {
		return nil, err
	}
	app.wal = w

	dsn := cfg.StarRocks.DSN()
	wr, err := writer.New(dsn, cfg.BatchWriter.BatchSize, cfg.BatchWriter.FlushInterval)
	if err != nil {
		return nil, err
	}
	app.writer = wr

	go app.collectLoop()

	return app, nil
}

func (app *CollectorApp) Stop() {
	if app.writer != nil {
		app.writer.Stop()
	}
	if app.wal != nil {
		app.wal.Close()
	}
}

func (app *CollectorApp) handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "collector",
		"queue": gin.H{
			"size": app.queue.Size(),
			"max":  10000,
		},
	})
}

func (app *CollectorApp) handleCollect(c *gin.Context) {
	var job model.JobMeta
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	job.StartTime = time.Now()

	if app.queue.IsFull() {
		if err := app.wal.Write(&job); err != nil {
			log.Printf("WAL write failed: %v", err)
		}
	} else {
		app.queue.Enqueue(&job)
	}

	c.JSON(200, gin.H{"status": "collected"})
}

func (app *CollectorApp) handleStats(c *gin.Context) {
	c.JSON(200, gin.H{
		"queue_size": app.queue.Size(),
		"queue_max":  10000,
	})
}

func (app *CollectorApp) collectLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			app.flushToWriter()
		case <-sigCh:
			app.Stop()
			return
		}
	}
}

func (app *CollectorApp) flushToWriter() {
	for {
		item, ok := app.queue.Dequeue()
		if !ok {
			break
		}

		job, ok := item.(*model.JobMeta)
		if !ok {
			continue
		}

		if err := app.writer.Write(job); err != nil {
			log.Printf("Write to StarRocks failed: %v, write to WAL", err)
			if err := app.wal.Write(job); err != nil {
				log.Printf("WAL write failed: %v", err)
			}
		}
	}
}
