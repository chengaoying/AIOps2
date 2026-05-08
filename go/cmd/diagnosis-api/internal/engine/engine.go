package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"aiops2/diagnosis-api/internal/cache"
	"aiops2/diagnosis-api/internal/context"
	"aiops2/diagnosis-api/internal/kb"
	"aiops2/diagnosis-api/internal/llm"
	"aiops2/diagnosis-api/internal/limiter"
)

type DiagnosisEngine struct {
	kb           kb.KnowledgeBase
	ctxBuilder   *context.ContextBuilder
	llmReasoner  *llm.LLMReasoner
	cache        *cache.Cache
	rateLimiter  *limiter.RateLimiter
}

func NewDiagnosisEngine(kb kb.KnowledgeBase, ctxBuilder *context.ContextBuilder, cache *cache.Cache, llmReasoner *llm.LLMReasoner) *DiagnosisEngine {
	return &DiagnosisEngine{
		kb:          kb,
		ctxBuilder:  ctxBuilder,
		llmReasoner: llmReasoner,
		cache:       cache,
		rateLimiter:  limiter.New(10.0, 20),
	}
}

func (e *DiagnosisEngine) Diagnose(ctx context.Context, req *DiagnosisRequest) (*DiagnosisResult, error) {
	start := time.Now()

	if req.UseCache {
		cacheKey := cache.BuildCacheKey(req.JobID)
		if cached, err := e.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			cached.DurationMs = time.Since(start).Milliseconds()
			return cached, nil
		}
	}

	if !e.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	diagCtx, err := e.ctxBuilder.Build(ctx, req)
	if err != nil {
		log.Printf("context build failed: %v", err)
	}

	cards, err := e.kb.Retrieve(ctx, &kb.RetrieveRequest{
		Platform: req.Platform,
		ErrorMsg: req.ErrorMsg,
		TopK:     5,
	})
	if err != nil {
		log.Printf("kb retrieve failed: %v", err)
	}

	var job *JobMeta
	if diagCtx != nil && diagCtx.Job != nil {
		job = diagCtx.Job
	} else {
		job = &JobMeta{
			JobID:    req.JobID,
			Platform: req.Platform,
			ErrorMsg: req.ErrorMsg,
		}
	}

	result, err := e.llmReasoner.Reason(ctx, job, cards)
	if err != nil {
		log.Printf("llm reason failed: %v, falling back", err)
		return e.fallback(job, cards), nil
	}

	result.DurationMs = time.Since(start).Milliseconds()

	if e.cache != nil && !req.UseCache {
		cacheKey := cache.BuildCacheKey(req.JobID)
		if err := e.cache.Set(ctx, cacheKey, result); err != nil {
			log.Printf("cache set failed: %v", err)
		}
	}

	return result, nil
}

func (e *DiagnosisEngine) fallback(job *JobMeta, cards []*KnowledgeCard) *DiagnosisResult {
	result := &DiagnosisResult{
		JobID:      job.JobID,
		Status:     "FAILED",
		Fallback:   true,
		Confidence: 0.5,
	}

	if len(cards) > 0 {
		card := cards[0]
		result.RootCause = card.RootCause
		result.Confidence = card.Confidence * 0.8
		result.Suggestions = card.Suggestions
		result.References = []string{card.ID}
	} else {
		result.RootCause = "未找到匹配规则，使用默认诊断"
		result.Suggestions = []Suggestion{
			{Action: "检查作业日志", Risk: "中", Detail: "查看详细错误信息"},
			{Action: "联系运维人员", Risk: "高", Detail: "无法自动诊断的问题"},
		}
	}

	return result
}

func (e *DiagnosisEngine) Health(ctx context.Context) error {
	if e.kb == nil {
		return fmt.Errorf("knowledge base not initialized")
	}
	return nil
}
