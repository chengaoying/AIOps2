package nlquery

import (
	"regexp"
	"strings"
	"time"

	"aiops2/diagnosis-api/internal/nlquery/keywords"
)

type IntentType string

const (
	IntentPerformanceAnalysis IntentType = "PERFORMANCE_ANALYSIS"
	IntentResourceAnalysis    IntentType = "RESOURCE_ANALYSIS"
	IntentFailureAnalysis     IntentType = "FAILURE_ANALYSIS"
	IntentTrendAnalysis       IntentType = "TREND_ANALYSIS"
	IntentJobQuery            IntentType = "JOB_QUERY"
	IntentMetricsQuery        IntentType = "METRICS_QUERY"
)

type Intent struct {
	Type       IntentType
	Confidence float64
	Keywords   []string
}

type IntentClassifier struct {
	intents []IntentDefinition
}

type IntentDefinition struct {
	Type      IntentType
	Keywords  []string
	Threshold float64
}

func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{
		intents: []IntentDefinition{
			{
				Type:      IntentPerformanceAnalysis,
				Keywords:  keywords.PerformanceKeywords,
				Threshold: 0.3,
			},
			{
				Type:      IntentResourceAnalysis,
				Keywords:  keywords.ResourceKeywords,
				Threshold: 0.3,
			},
			{
				Type:      IntentFailureAnalysis,
				Keywords:  keywords.FailureKeywords,
				Threshold: 0.3,
			},
			{
				Type:      IntentTrendAnalysis,
				Keywords:  keywords.TrendKeywords,
				Threshold: 0.3,
			},
			{
				Type:      IntentJobQuery,
				Keywords:  keywords.JobKeywords,
				Threshold: 0.3,
			},
			{
				Type:      IntentMetricsQuery,
				Keywords:  keywords.MetricsKeywords,
				Threshold: 0.3,
			},
		},
	}
}

func (c *IntentClassifier) Classify(text string) *Intent {
	text = strings.ToLower(text)
	bestIntent := &Intent{
		Type:       IntentPerformanceAnalysis,
		Confidence: 0.5,
	}
	bestScore := 0.0

	for _, intent := range c.intents {
		score := c.matchKeywords(text, intent.Keywords)
		if score > bestScore && score >= intent.Threshold {
			bestScore = score
			bestIntent = &Intent{
				Type:       intent.Type,
				Confidence: score,
				Keywords:   intent.Keywords,
			}
		}
	}

	return bestIntent
}

func (c *IntentClassifier) matchKeywords(text string, keywordList []string) float64 {
	matched := 0
	for _, kw := range keywordList {
		if strings.Contains(text, strings.ToLower(kw)) {
			matched++
		}
	}
	if len(keywordList) == 0 {
		return 0
	}
	return float64(matched) / float64(len(keywordList))
}
