package kb

import (
	"context"
	"sort"
	"strings"
	"sync"

	"aiops2/diagnosis-api/internal/engine"
)

type KnowledgeBase interface {
	Retrieve(ctx context.Context, req *RetrieveRequest) ([]*engine.KnowledgeCard, error)
}

type RetrieveRequest struct {
	Platform string
	ErrorMsg string
	TopK     int
}

type HybridKnowledgeBase struct {
	cards []*engine.KnowledgeCard
	mu    sync.RWMutex
}

func NewHybridKnowledgeBase() *HybridKnowledgeBase {
	return &HybridKnowledgeBase{
		cards: make([]*engine.KnowledgeCard, 0),
	}
}

func (kb *HybridKnowledgeBase) Retrieve(ctx context.Context, req *RetrieveRequest) ([]*engine.KnowledgeCard, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	if len(kb.cards) == 0 {
		return kb.getDefaultCards(), nil
	}

	var results []*engine.KnowledgeCard
	keyword := req.ErrorMsg

	for _, card := range kb.cards {
		if req.Platform != "" && card.Platform != req.Platform {
			continue
		}

		score := kb.matchScore(card, keyword)
		if score > 0 {
			cardCopy := *card
			cardCopy.Confidence = score
			results = append(results, &cardCopy)
		}
	}

	if len(results) == 0 {
		return kb.getDefaultCards(), nil
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	if req.TopK > 0 && len(results) > req.TopK {
		results = results[:req.TopK]
	}

	return results, nil
}

func (kb *HybridKnowledgeBase) matchScore(card *engine.KnowledgeCard, keyword string) float64 {
	if keyword == "" {
		return 0.5
	}

	keywordLower := strings.ToLower(keyword)

	if strings.Contains(strings.ToLower(card.ErrorType), keywordLower) {
		return 0.9
	}
	if strings.Contains(strings.ToLower(card.RootCause), keywordLower) {
		return 0.7
	}
	if strings.Contains(keywordLower, "oom") && strings.Contains(strings.ToLower(card.ErrorType), "memory") {
		return 0.8
	}
	if strings.Contains(keywordLower, "shuffle") && strings.Contains(strings.ToLower(card.ErrorType), "shuffle") {
		return 0.8
	}
	if strings.Contains(keywordLower, "timeout") && strings.Contains(strings.ToLower(card.ErrorType), "timeout") {
		return 0.8
	}

	return 0.3
}

func (kb *HybridKnowledgeBase) AddCard(card *engine.KnowledgeCard) {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.cards = append(kb.cards, card)
}

func (kb *HybridKnowledgeBase) getDefaultCards() []*engine.KnowledgeCard {
	return []*engine.KnowledgeCard{
		{
			ID:          "KB-SPANK-OOM-001",
			Platform:    "SPARK",
			ErrorType:   "Executor OOM",
			RootCause:   "Executor分配的内存不足以处理数据集",
			Confidence:  0.85,
			Suggestions: []engine.Suggestion{
				{Action: "增加executor内存", Risk: "低", Detail: "将spark.executor.memory从4g增加到6g", Command: "--conf spark.executor.memory=6g"},
				{Action: "优化数据分区", Risk: "中", Detail: "使用salting策略解决数据倾斜问题"},
			},
		},
		{
			ID:          "KB-SPANK-SHUFFLE-001",
			Platform:    "SPARK",
			ErrorType:   "Shuffle Error",
			RootCause:   "Shuffle过程中数据拉取失败",
			Confidence:  0.80,
			Suggestions: []engine.Suggestion{
				{Action: "增加shuffle分区数", Risk: "低", Detail: "增加spark.sql.shuffle.partitions"},
				{Action: "增加executor内存", Risk: "低", Detail: "增加spark.executor.memory"},
			},
		},
		{
			ID:          "KB-HIVE-MEM-001",
			Platform:    "HIVE",
			ErrorType:   "OutOfMemoryError",
			RootCause:   "Hive执行内存不足",
			Confidence:  0.85,
			Suggestions: []engine.Suggestion{
				{Action: "增加heap size", Risk: "低", Detail: "设置hive.heap.size=4g"},
				{Action: "优化查询", Risk: "中", Detail: "减少并发查询数"},
			},
		},
		{
			ID:          "KB-FLINK-CHECKPOINT-001",
			Platform:    "FLINK",
			ErrorType:   "Checkpoint Timeout",
			RootCause:   "Checkpoint在指定时间内未完成",
			Confidence:  0.80,
			Suggestions: []engine.Suggestion{
				{Action: "增加checkpoint timeout", Risk: "低", Detail: "设置execution.checkpointing.timeout=10min"},
				{Action: "减少状态大小", Risk: "中", Detail: "优化状态后端配置"},
			},
		},
		{
			ID:          "KB-YARN-OOM-001",
			Platform:    "YARN",
			ErrorType:   "Container OOM",
			RootCause:   "Container内存超限被Kill",
			Confidence:  0.85,
			Suggestions: []engine.Suggestion{
				{Action: "增加container内存", Risk: "低", Detail: "增加yarn.nodemanager.resource.memory-mb"},
				{Action: "调整虚拟内存比例", Risk: "中", Detail: "设置yarn.nodemanager.vmem-pmem-ratio=2.1"},
			},
		},
	}
}
