package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"aiops2/diagnosis-api/internal/cache"
	"aiops2/diagnosis-api/internal/context"
	"aiops2/diagnosis-api/internal/engine"
	"aiops2/diagnosis-api/internal/kb"
	"aiops2/diagnosis-api/internal/llm"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		log.Printf("db open failed: %v, continuing without db", err)
		db = nil
	}
	if db != nil {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
	}

	kbInstance := kb.NewHybridKnowledgeBase()

	var ctxBuilder *context.ContextBuilder
	if db != nil {
		ctxBuilder = context.New(db)
	}

	var cacheInstance *cache.Cache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr != "" {
		cacheInstance, err = cache.New(redisAddr, time.Hour)
		if err != nil {
			log.Printf("cache init failed: %v", err)
		}
	}

	llmReasoner := llm.NewLLMReasoner(
		os.Getenv("LLM_API_KEY"),
		os.Getenv("LLM_ENDPOINT"),
	)

	diagEngine := engine.NewDiagnosisEngine(kbInstance, ctxBuilder, cacheInstance, llmReasoner)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "diagnosis-api"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/dashboard/home", handleDashboardHome)
		api.POST("/diagnosis", handleDiagnosis(diagEngine))
		api.GET("/diagnosis/history", handleDiagnosisHistory)
		api.POST("/knowledge/retrieve", handleKnowledgeRetrieve(kbInstance))
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

func getDSN() string {
	host := os.Getenv("STARROCKS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("STARROCKS_PORT")
	if port == "" {
		port = "9030"
	}
	user := os.Getenv("STARROCKS_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("STARROCKS_PASSWORD")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/aiops?charset=utf8mb4", user, password, host, port)
}

func handleDashboardHome(c *gin.Context) {
	c.JSON(200, gin.H{
		"date":    time.Now().Format("2006-01-02"),
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

func handleDiagnosis(diagEngine *engine.DiagnosisEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req engine.DiagnosisRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		result, err := diagEngine.Diagnose(c.Request.Context(), &req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, result)
	}
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

func handleKnowledgeRetrieve(kbInstance kb.KnowledgeBase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Platform string `json:"platform"`
			ErrorMsg string `json:"error_msg"`
			TopK     int    `json:"top_k"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		if req.TopK == 0 {
			req.TopK = 5
		}

		cards, err := kbInstance.Retrieve(context.Background(), &kb.RetrieveRequest{
			Platform: req.Platform,
			ErrorMsg: req.ErrorMsg,
			TopK:     req.TopK,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"cards": cards})
	}
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
