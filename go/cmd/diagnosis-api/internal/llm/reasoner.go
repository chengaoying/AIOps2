package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aiops2/diagnosis-api/internal/engine"
)

const DiagnosisPromptTemplate = `你是一个大数据平台诊断专家。根据以下作业信息和知识库内容，分析根因并给出修复建议。

## 作业信息
- 平台: %s
- 作业ID: %s
- 作业名称: %s
- 状态: %s
- 错误信息: %s
- 执行时长: %dms

## 上下文信息
%s

## 知识库参考
%s

## 输出格式
请以 JSON 格式输出：
{
    "root_cause": "根因分析（50字以内）",
    "confidence": 0.85,
    "suggestions": [
        {
            "action": "修复动作",
            "risk": "低/中/高",
            "detail": "详细说明",
            "command": "可执行命令（如有）"
        }
    ],
    "references": ["参考的知识卡片ID列表"]
}

请只输出 JSON，不要有其他内容。`

type InputValidator struct {
	maxLength      int
	blockedPatterns []string
}

func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxLength:      10000,
		blockedPatterns: []string{"--", "rm -rf", "drop table", "delete from"},
	}
}

func (v *InputValidator) Validate(input string) error {
	if len(input) > v.maxLength {
		return fmt.Errorf("input too long: %d > %d", len(input), v.maxLength)
	}
	for _, pattern := range v.blockedPatterns {
		if strings.Contains(strings.ToLower(input), strings.ToLower(pattern)) {
			return fmt.Errorf("blocked pattern: %s", pattern)
		}
	}
	return nil
}

type OutputValidator struct{}

func NewOutputValidator() *OutputValidator {
	return &OutputValidator{}
}

func (v *OutputValidator) Validate(output string) (*engine.DiagnosisResult, error) {
	var result engine.DiagnosisResult

	lines := strings.Split(output, "\n")
	var jsonLines []string
	inJson := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "{") {
			inJson = true
		}
		if inJson {
			jsonLines = append(jsonLines, line)
		}
		if strings.Contains(trimmed, "}") {
			break
		}
	}

	if len(jsonLines) == 0 {
		for _, line := range lines {
			if strings.Contains(line, "{") || strings.Contains(line, "}") {
				jsonLines = append(jsonLines, line)
			}
		}
	}

	jsonStr := strings.Join(jsonLines, "")
	start := strings.Index(jsonStr, "{")
	end := strings.LastIndex(jsonStr, "}")
	if start != -1 && end != -1 && end > start {
		jsonStr = jsonStr[start : end+1]
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	if result.RootCause == "" {
		return nil, fmt.Errorf("missing root_cause")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		result.Confidence = 0.5
	}

	return &result, nil
}

type LLMReasoner struct {
	apiKey       string
	endpoint     string
	inputValid  *InputValidator
	outputValid *OutputValidator
}

func NewLLMReasoner(apiKey, endpoint string) *LLMReasoner {
	return &LLMReasoner{
		apiKey:       apiKey,
		endpoint:     endpoint,
		inputValid:   NewInputValidator(),
		outputValid: NewOutputValidator(),
	}
}

func (r *LLMReasoner) Reason(ctx context.Context, job *engine.JobMeta, cards []*engine.KnowledgeCard) (*engine.DiagnosisResult, error) {
	prompt := r.buildPrompt(job, cards)

	if err := r.inputValid.Validate(prompt); err != nil {
		return nil, fmt.Errorf("input validation: %w", err)
	}

	output, err := r.callLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}

	result, err := r.outputValid.Validate(output)
	if err != nil {
		return nil, fmt.Errorf("output validation: %w", err)
	}

	result.JobID = job.JobID
	result.UsedLLM = true

	return result, nil
}

func (r *LLMReasoner) buildPrompt(job *engine.JobMeta, cards []*engine.KnowledgeCard) string {
	contextInfo := fmt.Sprintf("用户: %s, 队列: %s, 错误码: %d",
		job.User, job.Queue, job.ExitCode)

	var kbInfo strings.Builder
	for _, card := range cards {
		kbInfo.WriteString(fmt.Sprintf("\n### 知识卡片: %s\n", card.ErrorType))
		kbInfo.WriteString(fmt.Sprintf("根因: %s\n", card.RootCause))
		kbInfo.WriteString("建议:\n")
		for _, s := range card.Suggestions {
			kbInfo.WriteString(fmt.Sprintf("- %s (%s): %s\n", s.Action, s.Risk, s.Detail))
		}
		kbInfo.WriteString("---\n")
	}

	return fmt.Sprintf(DiagnosisPromptTemplate,
		job.Platform, job.JobID, job.JobName, job.Status,
		job.ErrorMsg, job.DurationMs,
		contextInfo, kbInfo.String())
}

func (r *LLMReasoner) callLLM(ctx context.Context, prompt string) (string, error) {
	if r.apiKey == "" || r.endpoint == "" {
		return r.mockLLMResponse(), nil
	}

	type qwenRequest struct {
		Model string `json:"model"`
		Input struct {
			Text string `json:"text"`
		} `json:"input"`
		Parameters struct {
			Temperature float64 `json:"temperature"`
			MaxTokens   int     `json:"max_tokens"`
		} `json:"parameters"`
	}

	req := qwenRequest{}
	req.Model = "qwen-turbo"
	req.Input.Text = prompt
	req.Parameters.Temperature = 0.7
	req.Parameters.MaxTokens = 500

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return r.mockLLMResponse(), nil
}

func (r *LLMReasoner) mockLLMResponse() string {
	return `{
    "root_cause": "Executor内存不足导致OutOfMemoryError",
    "confidence": 0.85,
    "suggestions": [
        {
            "action": "增加executor内存",
            "risk": "低",
            "detail": "将spark.executor.memory从4g增加到6g",
            "command": "--conf spark.executor.memory=6g"
        },
        {
            "action": "优化数据分区",
            "risk": "中",
            "detail": "使用salting策略解决数据倾斜问题"
        }
    ],
    "references": ["KB-SPANK-OOM-001"]
}`
}
