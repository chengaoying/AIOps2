package llm

import (
	"testing"
)

func TestInputValidator_Validate(t *testing.T) {
	v := NewInputValidator()

	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantMsg string
	}{
		{"normal", "Show me the logs for job123", true, ""},
		{"short", "ok", false, ""},
		{"empty", "", false, ""},
		{"whitespace", "   ", false, ""},
		{"long", string(make([]byte, 5001)), false, ""},
		{"sql_injection", "SELECT * FROM users; DROP TABLE users;", false, ""},
		{"script_tag", "<script>alert('xss')</script>", false, ""},
		{"unicode_normal", "作业job123正常运行", true, ""},
		{"emoji", "😊", false, ""},
		{"mixed_content", "Job job123 failed with OOM error", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			ok := err == nil
			if ok != tt.wantOK {
				t.Errorf("Validate(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
		})
	}
}

func TestInputValidator_CheckSQLInjection(t *testing.T) {
	v := NewInputValidator()

	injectionPatterns := []string{
		"SELECT * FROM job_meta",
		"DROP TABLE",
		"UNION SELECT",
		"'; DELETE FROM",
		"1=1",
		"OR 1=1",
	}

	for _, pattern := range injectionPatterns {
		err := v.Validate(pattern)
		if err == nil {
			t.Errorf("Expected SQL injection %q to be blocked", pattern)
		}
	}
}

func TestInputValidator_CheckXSS(t *testing.T) {
	v := NewInputValidator()

	xssPatterns := []string{
		"<script>",
		"javascript:",
		"onerror=",
		"onclick=",
		"<iframe>",
	}

	for _, pattern := range xssPatterns {
		err := v.Validate(pattern)
		if err == nil {
			t.Errorf("Expected XSS %q to be blocked", pattern)
		}
	}
}

func TestInputValidator_CheckLength(t *testing.T) {
	v := NewInputValidator()

	short := "short"
	err := v.Validate(short)
	if err == nil {
		t.Error("Expected short input to be blocked")
	}

	long := string(make([]byte, 5001))
	err = v.Validate(long)
	if err == nil {
		t.Error("Expected too long input to be blocked")
	}
}

func TestOutputValidator_Validate(t *testing.T) {
	v := NewOutputValidator()

	tests := []struct {
		name   string
		output string
		wantOK bool
	}{
		{"normal", "Executor OOM caused by memory limit", true},
		{"empty", "", false},
		{"truncated", "Exec", false},
		{"valid_json", `{"diagnosis":"OOM","confidence":0.9}`, true},
		{"normal_text", "The job failed due to resource exhaustion", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.output)
			ok := err == nil
			if ok != tt.wantOK {
				t.Errorf("Validate(%q) ok = %v, want %v", tt.output, ok, tt.wantOK)
			}
		})
	}
}

func TestOutputValidator_CheckJSON(t *testing.T) {
	v := NewOutputValidator()

	_, err := v.ParseJSON(`{"diagnosis":"success","confidence":0.95}`)
	if err != nil {
		t.Errorf("ParseJSON failed: %v", err)
	}

	_, err = v.ParseJSON(`invalid json`)
	if err == nil {
		t.Error("Expected invalid JSON to fail parsing")
	}
}

func TestOutputValidator_CheckStructure(t *testing.T) {
	v := NewOutputValidator()

	validOutput := `The diagnosis result is Executor OOM.`
	err := v.Validate(validOutput)
	if err != nil {
		t.Errorf("Valid output should pass: %v", err)
	}
}

func TestOutputValidator_RemoveSensitive(t *testing.T) {
	v := NewOutputValidator()

	outputs := []string{
		"password: secret123",
		"token: abc123",
		"api_key: xyz",
		"Job executed successfully",
	}

	for _, output := range outputs {
		cleaned := v.RemoveSensitive(output)
		if len(cleaned) >= len(output) {
			t.Errorf("RemoveSensitive should shorten %q", output)
		}
	}
}

func TestPromptBuilder_Build(t *testing.T) {
	pb := NewPromptBuilder()

	tests := []struct {
		name      string
		job       *JobContext
		cards     []*KnowledgeCard
		wantError bool
	}{
		{
			name: "normal_job",
			job: &JobContext{
				JobID:    "job123",
				Platform: "SPARK",
				ErrorMsg: "Executor OOM",
				JobName:  "test_job",
				User:     "admin",
			},
			cards: []*KnowledgeCard{
				{RootCause: "Memory limit exceeded", Confidence: 0.9},
			},
			wantError: false,
		},
		{
			name: "nil_job",
			job:  nil,
			cards: []*KnowledgeCard{
				{RootCause: "Test", Confidence: 0.5},
			},
			wantError: true,
		},
		{
			name: "empty_cards",
			job: &JobContext{
				JobID:    "job123",
				Platform: "YARN",
				ErrorMsg: "Error",
			},
			cards:   []*KnowledgeCard{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := pb.Build(tt.job, tt.cards)
			if (err != nil) != tt.wantError {
				t.Errorf("Build() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && prompt == "" {
				t.Error("Build() returned empty prompt")
			}
		})
	}
}

func TestPromptBuilder_Escape(t *testing.T) {
	pb := NewPromptBuilder()

	tests := []struct {
		input    string
		expected string
	}{
		{"normal text", "normal text"},
		{"text with 'quote'", "text with ''quote''"},
		{"text with \"double\"", "text with \"\"double\"\""},
		{"multi\nline\", \"multi\\nline\""},
		{"\ttab\", \"\\ttab\""},
	}

	for _, tt := range tests {
		result := pb.Escape(tt.input)
		if result != tt.expected {
			t.Errorf("Escape(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestPromptBuilder_BuildWithContext(t *testing.T) {
	pb := NewPromptBuilder()

	job := &JobContext{
		JobID:    "job123",
		Platform: "SPARK",
		ErrorMsg: "OOM",
		JobName:  "data_processing",
		User:     "admin",
		Duration: 300000,
	}

	cards := []*KnowledgeCard{
		{
			RootCause:  "Executor memory too small",
			Suggestions: []Suggestion{{Action: "Increase executor memory"}},
			Confidence: 0.95,
		},
	}

	prompt, err := pb.Build(job, cards)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if prompt == "" {
		t.Fatal("Build() returned empty prompt")
	}
}

func TestPromptBuilder_EmptyJob(t *testing.T) {
	pb := NewPromptBuilder()

	_, err := pb.Build(nil, []*KnowledgeCard{{RootCause: "test"}})
	if err == nil {
		t.Error("Expected error for nil job")
	}
}

func TestPromptBuilder_LongJobID(t *testing.T) {
	pb := NewPromptBuilder()

	job := &JobContext{
		JobID:    string(make([]byte, 300)),
		Platform: "YARN",
		ErrorMsg: "Error",
	}

	_, err := pb.Build(job, nil)
	if err == nil {
		t.Error("Expected error for too long job ID")
	}
}

func TestInputValidator_MultipleValidation(t *testing.T) {
	v := NewInputValidator()

	if err := v.Validate("normal query"); err != nil {
		t.Error("First validation should pass")
	}
	if err := v.Validate("short"); err == nil {
		t.Error("Second validation should fail")
	}
}

func TestOutputValidator_EmptyOutput(t *testing.T) {
	v := NewOutputValidator()

	err := v.Validate("")
	if err == nil {
		t.Error("Empty output should be invalid")
	}
}

func TestPromptBuilder_BuildWithNoSuggestions(t *testing.T) {
	pb := NewPromptBuilder()

	job := &JobContext{
		JobID:    "job123",
		Platform: "HIVE",
		ErrorMsg: "Error",
	}

	cards := []*KnowledgeCard{
		{RootCause: "Unknown cause", Confidence: 0.3},
	}

	prompt, err := pb.Build(job, cards)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if prompt == "" {
		t.Error("Build() should produce prompt even without suggestions")
	}
}
