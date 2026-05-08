package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "diagnosis-api"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/dashboard/home", handleDashboardHome)
		api.POST("/diagnosis", handleDiagnosis)
		api.GET("/diagnosis/history", handleDiagnosisHistory)
		api.POST("/knowledge/retrieve", handleKnowledgeRetrieve)
		api.POST("/assistant/chat", handleChat)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("AIOps Diagnosis API starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleDashboardHome(c *gin.Context) {
	c.JSON(200, gin.H{
		"date":    "2026-05-08",
		"cluster": "生产集群",
		"today_stats": gin.H{
			"total_count":   156,
			"success_count": 144,
			"failed_count":  12,
			"success_rate": 92.3,
		},
		"platform_dist": []gin.H{
			{"platform": "YARN", "count": 45, "percentage": 28},
			{"platform": "SPARK", "count": 67, "percentage": 42},
			{"platform": "HIVE", "count": 23, "percentage": 14},
			{"platform": "FLINK", "count": 21, "percentage": 13},
		},
		"recent_jobs": []gin.H{
			{"job_id": "spark_job_001", "platform": "SPARK", "status": "FAILED", "timestamp": "10:32", "root_cause": "Executor OOM"},
			{"job_id": "hive_query_042", "platform": "HIVE", "status": "SUCCESS", "timestamp": "10:28"},
			{"job_id": "yarn_app_089", "platform": "YARN", "status": "SUCCESS", "timestamp": "10:15"},
		},
	})
}

func handleDiagnosis(c *gin.Context) {
	var req struct {
		JobID   string `json:"job_id"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(200, gin.H{
		"job_id":     req.JobID,
		"platform":   "SPARK",
		"status":     "FAILED",
		"root_cause": "Executor 内存溢出，导致 Task 被 Kill",
		"confidence": 0.92,
		"suggestions": []gin.H{
			{"action": "增加 executor 内存", "risk": "低", "detail": "将 spark.executor.memory 从 4g 增加到 6g", "command": "--conf spark.executor.memory=6g"},
			{"action": "优化数据分区", "risk": "中", "detail": "使用 salting 策略解决数据倾斜问题"},
		},
	})
}

func handleDiagnosisHistory(c *gin.Context) {
	c.JSON(200, gin.H{
		"jobs": []gin.H{
			{"job_id": "spark_job_001", "platform": "SPARK", "status": "FAILED", "time": "10:32", "root_cause": "Executor OOM"},
			{"job_id": "hive_query_042", "platform": "HIVE", "status": "SUCCESS", "time": "10:28"},
			{"job_id": "yarn_app_089", "platform": "YARN", "status": "SUCCESS", "time": "10:15"},
		},
		"total": 3,
	})
}

func handleKnowledgeRetrieve(c *gin.Context) {
	var req struct {
		Platform string `json:"platform"`
		ErrorMsg string `json:"error_msg"`
		TopK     int    `json:"top_k"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(200, gin.H{
		"cards": []gin.H{
			{
				"id":          "KB-20260508-00001",
				"platform":    "SPARK",
				"error_type":  "Executor OOM",
				"root_cause":  "Executor 分配的内存不足以处理数据集",
				"confidence":  0.92,
				"suggestions": []gin.H{
					{"action": "增加 executor 内存", "risk": "低", "detail": "将 spark.executor.memory 从 4g 增加到 6g"},
				},
			},
		},
	})
}

func handleChat(c *gin.Context) {
	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(200, gin.H{
		"reply": "您好，我是 AIOps AI 助手。关于 " + req.Message + "，我可以帮您分析和解答。",
	})
}