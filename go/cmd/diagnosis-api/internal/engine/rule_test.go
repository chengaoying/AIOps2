package engine

import (
	"context"
	"testing"
)

type mockKB struct {
	cards []*KnowledgeCard
}

func (m *mockKB) Retrieve(ctx context.Context, req *RetrieveRequest) ([]*KnowledgeCard, error) {
	var result []*KnowledgeCard
	for _, card := range m.cards {
		if card.Platform == req.Platform {
			result = append(result, card)
		}
	}
	return result, nil
}

type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string) (*DiagnosisResult, error) {
	return nil, nil
}

func (m *mockCache) Set(ctx context.Context, key string, result *DiagnosisResult) error {
	return nil
}

type KnowledgeCard struct {
	ID           string
	Platform     string
	ErrorType    string
	ErrorPatterns []string
	RootCause    string
	Suggestions  []Suggestion
	Confidence   float64
}

type RetrieveRequest struct {
	Platform string
	ErrorMsg string
	TopK     int
}

type Suggestion struct {
	Action string
	Risk   string
	Detail string
}

func TestRuleEngine_Match_YARN_OOM(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{
			{
				ID:        "yarn-oom-001",
				Platform:  "YARN",
				ErrorType: "OOM",
				ErrorPatterns: []string{
					"java.lang.OutOfMemoryError",
					"OutOfMemoryError",
					"memory exhausted",
				},
				RootCause:  "Container内存不足",
				Suggestions: []Suggestion{
					{Action: "增加Container内存", Risk: "中", Detail: "调大yarn.scheduler.maximum-allocation-mb"},
				},
				Confidence: 0.9,
			},
		},
	}

	rules := []Rule{
		{
			ID:          "yarn-oom",
			Platform:    "YARN",
			Patterns:    []string{"OutOfMemoryError", "memory"},
			RootCause:   "内存溢出",
			Suggestions: []Suggestion{{Action: "检查内存配置", Risk: "高"}},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "Container killed due to OutOfMemoryError",
	})

	if !result.Matched {
		t.Error("Expected OOM pattern to match")
	}
}

func TestRuleEngine_Match_Spark_OOM(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{
			{
				ID:        "spark-oom-001",
				Platform:  "SPARK",
				ErrorType: "OOM",
				ErrorPatterns: []string{
					"Executor OOM",
					"OutOfMemoryError",
				},
				RootCause:  "Executor内存不足",
				Suggestions: []Suggestion{
					{Action: "增加executor内存", Risk: "中"},
				},
				Confidence: 0.9,
			},
		},
	}

	rules := []Rule{
		{
			ID:          "spark-oom",
			Platform:    "SPARK",
			Patterns:    []string{"Executor OOM", "OutOfMemoryError"},
			RootCause:   "Executor内存不足",
			Suggestions: []Suggestion{{Action: "增加executor.memory", Risk: "中"}},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "SPARK",
		ErrorMsg: "ExecutorLostError: Executor OOM",
	})

	if !result.Matched {
		t.Error("Expected Spark OOM pattern to match")
	}
}

func TestRuleEngine_Match_Hive_Memory(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{
			{
				ID:        "hive-mem-001",
				Platform:  "HIVE",
				ErrorType: "MEMORY",
				ErrorPatterns: []string{
					"OutOfMemoryError",
					"MemoryException",
				},
				RootCause:  "Hive内存不足",
				Suggestions: []Suggestion{
					{Action: "增加hive内存", Risk: "中"},
				},
				Confidence: 0.85,
			},
		},
	}

	rules := []Rule{
		{
			ID:          "hive-mem",
			Platform:    "HIVE",
			Patterns:    []string{"OutOfMemoryError", "MemoryException"},
			RootCause:   "内存不足",
			Suggestions: []Suggestion{{Action: "调高hive.auto.convert.memory", Risk: "中"}},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "HIVE",
		ErrorMsg: "HiveException: OutOfMemoryError",
	})

	if !result.Matched {
		t.Error("Expected Hive memory pattern to match")
	}
}

func TestRuleEngine_Match_Flink_Checkpoint(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{
			{
				ID:        "flink-cp-001",
				Platform:  "FLINK",
				ErrorType: "CHECKPOINT_TIMEOUT",
				ErrorPatterns: []string{
					"CheckpointTimeout",
					"checkpoint timeout",
				},
				RootCause:  "Checkpoint超时",
				Suggestions: []Suggestion{
					{Action: "增加checkpoint超时时间", Risk: "中"},
				},
				Confidence: 0.9,
			},
		},
	}

	rules := []Rule{
		{
			ID:          "flink-cp",
			Platform:    "FLINK",
			Patterns:    []string{"CheckpointTimeout", "checkpoint timeout"},
			RootCause:   "Checkpoint超时",
			Suggestions: []Suggestion{{Action: "调高execution.checkpointing.timeout", Risk: "中"}},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "FLINK",
		ErrorMsg: "CheckpointTimeoutException: checkpoint timed out",
	})

	if !result.Matched {
		t.Error("Expected Flink checkpoint pattern to match")
	}
}

func TestRuleEngine_Match_NoMatch(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{},
	}

	rules := []Rule{
		{
			ID:          "yarn-oom",
			Platform:    "YARN",
			Patterns:    []string{"OutOfMemoryError"},
			RootCause:   "内存溢出",
			Suggestions: []Suggestion{{Action: "检查内存", Risk: "中"}},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "Normal execution completed",
	})

	if result.Matched {
		t.Error("Expected no match for normal message")
	}
}

func TestRuleEngine_Match_PlatformMismatch(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{},
	}

	rules := []Rule{
		{
			ID:          "spark-oom",
			Platform:    "SPARK",
			Patterns:    []string{"OutOfMemoryError"},
			RootCause:   "OOM",
			Suggestions: []Suggestion{},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "OutOfMemoryError",
	})

	if result.Matched {
		t.Error("Expected no match due to platform mismatch")
	}
}

func TestRuleEngine_Match_MultiplePatterns(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{},
	}

	rules := []Rule{
		{
			ID:          "yarn-oom",
			Platform:    "YARN",
			Patterns:    []string{"OutOfMemoryError", "memory exhausted", "OOM"},
			RootCause:   "内存问题",
			Suggestions: []Suggestion{},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "Container OOM",
	})

	if !result.Matched {
		t.Error("Expected match for OOM abbreviation")
	}
}

func TestRuleEngine_Match_CaseInsensitive(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{},
	}

	rules := []Rule{
		{
			ID:          "yarn-oom",
			Platform:    "YARN",
			Patterns:    []string{"outofmemoryerror"},
			RootCause:   "OOM",
			Suggestions: []Suggestion{},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "OutOfMemoryError: Heap Space",
	})

	if !result.Matched {
		t.Error("Expected case-insensitive match")
	}
}

func TestRuleEngine_Match_Confidence(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{
			{
				ID:         "test",
				Platform:   "YARN",
				RootCause:  "Test",
				Confidence: 0.95,
			},
		},
	}

	rules := []Rule{
		{
			ID:          "test",
			Platform:    "YARN",
			Patterns:    []string{"error"},
			RootCause:   "Test root cause",
			Suggestions: []Suggestion{},
			Confidence: 0.88,
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "Some error occurred",
	})

	if result.Confidence != 0.88 {
		t.Errorf("Confidence = %v, want 0.88", result.Confidence)
	}
}

func TestRuleEngine_Match_WithKB(t *testing.T) {
	kb := &mockKB{
		cards: []*KnowledgeCard{
			{
				ID:         "yarn-oom-kb",
				Platform:   "YARN",
				RootCause:  "KB Root Cause",
				Confidence: 0.92,
			},
		},
	}

	rules := []Rule{
		{
			ID:          "yarn-oom",
			Platform:    "YARN",
			Patterns:    []string{"OutOfMemoryError"},
			RootCause:   "Rule Root Cause",
			Suggestions: []Suggestion{},
			Confidence: 0.85,
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "OutOfMemoryError occurred",
	})

	if !result.Matched {
		t.Error("Expected match")
	}
	if result.RootCause != "Rule Root Cause" {
		t.Errorf("RootCause = %s, want Rule Root Cause", result.RootCause)
	}
}

func TestRuleEngine_Match_EmptyRules(t *testing.T) {
	kb := &mockKB{cards: []*KnowledgeCard{}}

	engine := NewRuleEngine(kb, []Rule{})

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "Any error",
	})

	if result.Matched {
		t.Error("Expected no match with empty rules")
	}
}

func TestRuleEngine_Match_EmptyErrorMsg(t *testing.T) {
	kb := &mockKB{cards: []*KnowledgeCard{}}

	rules := []Rule{
		{
			ID:       "test",
			Platform: "YARN",
			Patterns: []string{"error"},
		},
	}

	engine := NewRuleEngine(kb, rules)

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "",
	})

	if result.Matched {
		t.Error("Expected no match with empty error message")
	}
}

func TestRuleEngine_Match_AllPlatforms(t *testing.T) {
	kb := &mockKB{cards: []*KnowledgeCard{}}

	rules := []Rule{
		{ID: "yarn-err", Platform: "YARN", Patterns: []string{"yarn_error"}},
		{ID: "hive-err", Platform: "HIVE", Patterns: []string{"hive_error"}},
		{ID: "spark-err", Platform: "SPARK", Patterns: []string{"spark_error"}},
		{ID: "flink-err", Platform: "FLINK", Patterns: []string{"flink_error"}},
	}

	engine := NewRuleEngine(kb, rules)
	ctx := context.Background()

	tests := []struct {
		platform string
		msg      string
		want     bool
	}{
		{"YARN", "yarn_error happened", true},
		{"HIVE", "hive_error happened", true},
		{"SPARK", "spark_error happened", true},
		{"FLINK", "flink_error happened", true},
		{"YARN", "hive_error happened", false},
		{"HIVE", "spark_error happened", false},
	}

	for _, tt := range tests {
		result := engine.Match(ctx, &MatchRequest{
			Platform: tt.platform,
			ErrorMsg:  tt.msg,
		})
		if result.Matched != tt.want {
			t.Errorf("Match(%s, %s) = %v, want %v", tt.platform, tt.msg, result.Matched, tt.want)
		}
	}
}

func TestRuleEngine_AddRule(t *testing.T) {
	kb := &mockKB{cards: []*KnowledgeCard{}}
	engine := NewRuleEngine(kb, []Rule{})

	engine.AddRule(Rule{
		ID:       "new-rule",
		Platform: "YARN",
		Patterns: []string{"new_error"},
	})

	ctx := context.Background()
	result := engine.Match(ctx, &MatchRequest{
		Platform: "YARN",
		ErrorMsg: "new_error occurred",
	})

	if !result.Matched {
		t.Error("Expected newly added rule to match")
	}
}

func TestRuleEngine_RuleCount(t *testing.T) {
	kb := &mockKB{cards: []*KnowledgeCard{}}
	rules := []Rule{
		{ID: "1", Platform: "YARN", Patterns: []string{"a"}},
		{ID: "2", Platform: "YARN", Patterns: []string{"b"}},
		{ID: "3", Platform: "HIVE", Patterns: []string{"c"}},
	}

	engine := NewRuleEngine(kb, rules)

	count := engine.RuleCount()
	if count != 3 {
		t.Errorf("RuleCount() = %d, want 3", count)
	}
}

func TestRuleEngine_GetRulesByPlatform(t *testing.T) {
	kb := &mockKB{cards: []*KnowledgeCard{}}
	rules := []Rule{
		{ID: "1", Platform: "YARN", Patterns: []string{"a"}},
		{ID: "2", Platform: "YARN", Patterns: []string{"b"}},
		{ID: "3", Platform: "HIVE", Patterns: []string{"c"}},
	}

	engine := NewRuleEngine(kb, rules)

	yarnRules := engine.GetRulesByPlatform("YARN")
	if len(yarnRules) != 2 {
		t.Errorf("YARN rules count = %d, want 2", len(yarnRules))
	}

	hiveRules := engine.GetRulesByPlatform("HIVE")
	if len(hiveRules) != 1 {
		t.Errorf("HIVE rules count = %d, want 1", len(hiveRules))
	}
}
