package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "collector"})
	})

	// Job collection endpoint
	r.POST("/api/v1/collect", handleCollect)

	// Batch write jobs to StarRocks
	go startCollectorWorker()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("AIOps Collector starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start collector: %v", err)
	}
}

func startCollectorWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Collector worker: batching jobs to StarRocks...")
		// Batch write logic here
	}
}

func handleCollect(c *gin.Context) {
	var job struct {
		Platform  string `json:"platform"`
		JobID     string `json:"job_id"`
		JobName   string `json:"job_name"`
		Status    string `json:"status"`
		ErrorMsg  string `json:"error_msg"`
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	log.Printf("Collector: received job %s from %s", job.JobID, job.Platform)
	c.JSON(200, gin.H{"status": "collected"})
}