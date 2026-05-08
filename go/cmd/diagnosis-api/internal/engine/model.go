package engine

import "time"

type DiagnosisRequest struct {
	Platform string `json:"platform"`
	JobID    string `json:"job_id"`
	ErrorMsg string `json:"error_msg,omitempty"`
	UseCache bool   `json:"use_cache"`
	ForceLLM bool   `json:"force_llm"`
}

type DiagnosisResult struct {
	JobID       string       `json:"job_id"`
	Status      string       `json:"status"`
	RootCause   string       `json:"root_cause"`
	Confidence  float64      `json:"confidence"`
	Suggestions []Suggestion `json:"suggestions"`
	References  []string     `json:"references"`
	UsedCache   bool         `json:"used_cache"`
	UsedLLM     bool         `json:"used_llm"`
	Fallback    bool         `json:"fallback"`
	DurationMs  int64        `json:"duration_ms"`
}

type Suggestion struct {
	Action  string `json:"action"`
	Risk    string `json:"risk"`
	Detail  string `json:"detail"`
	Command string `json:"command,omitempty"`
}

type DiagnosisContext struct {
	Job          *JobMeta       `json:"job"`
	JobChain     []*JobMeta     `json:"job_chain"`
	SimilarCases []*JobMeta     `json:"similar_cases"`
	RelatedLogs  []string       `json:"related_logs"`
	Metrics      map[string]any `json:"metrics"`
}

type JobMeta struct {
	JobID      string    `json:"job_id"`
	Platform   string    `json:"platform"`
	JobName    string    `json:"job_name"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	DurationMs int64     `json:"duration_ms"`
	User       string    `json:"user"`
	Queue      string    `json:"queue"`
	ErrorMsg   string    `json:"error_msg"`
	ExitCode   int       `json:"exit_code"`
}

type KnowledgeCard struct {
	ID            string       `json:"id"`
	Platform     string       `json:"platform"`
	ErrorType   string       `json:"error_type"`
	RootCause   string       `json:"root_cause"`
	Suggestions  []Suggestion `json:"suggestions"`
	Confidence   float64      `json:"confidence"`
	Source       string       `json:"source"`
	UsageCount   int          `json:"usage_count"`
}
