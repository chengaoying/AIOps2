package context

import (
	"testing"
)

func TestExtractErrorPatterns_Empty(t *testing.T) {
	b := New(nil)

	patterns := b.extractErrorPatterns("")
	if len(patterns) != 0 {
		t.Errorf("extractErrorPatterns(\"\") returned %d patterns, want 0", len(patterns))
	}
}

func TestExtractErrorPatterns_Normal(t *testing.T) {
	b := New(nil)

	patterns := b.extractErrorPatterns("Executor OOM error")
	if len(patterns) < 1 {
		t.Error("extractErrorPatterns should return at least the original message")
	}
}

func TestExtractErrorPatterns_Keywords(t *testing.T) {
	b := New(nil)

	testCases := []struct {
		msg      string
		contains string
	}{
		{"OutOfMemoryError in executor", "OutOfMemory"},
		{"Shuffle error during fetch", "Shuffle"},
		{"Connection timeout to driver", "Connection"},
		{"Auth failed for user", "Auth"},
	}

	for _, tc := range testCases {
		patterns := b.extractErrorPatterns(tc.msg)
		found := false
		for _, p := range patterns {
			if p == tc.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("extractErrorPatterns(%q) should contain %q", tc.msg, tc.contains)
		}
	}
}

func TestExtractErrorPatterns_CaseInsensitive(t *testing.T) {
	b := New(nil)

	patterns := b.extractErrorPatterns("OOM error")
	found := false
	for _, p := range patterns {
		if p == "OOM" || p == "oom" {
			found = true
		}
	}
	if !found {
		t.Error("extractErrorPatterns should be case-insensitive for OOM")
	}
}

func TestExtractKeyError_Empty(t *testing.T) {
	result := extractKeyError("")
	if result != "" {
		t.Errorf("extractKeyError(\"\") = %q, want empty string", result)
	}
}

func TestExtractKeyError_SingleLine(t *testing.T) {
	result := extractKeyError("Single line error")
	if result != "Single line error" {
		t.Errorf("extractKeyError(\"Single line error\") = %q", result)
	}
}

func TestExtractKeyError_MultiLine(t *testing.T) {
	msg := "First line is the key error\nSecond line\nThird line"
	result := extractKeyError(msg)
	if result != "First line is the key error" {
		t.Errorf("extractKeyError multi-line = %q, want \"First line is the key error\"", result)
	}
}

func TestExtractKeyError_WithNewlines(t *testing.T) {
	msg := "Error: OOM\nCaused by: memory limit"
	result := extractKeyError(msg)
	if result != "Error: OOM" {
		t.Errorf("extractKeyError with newlines = %q", result)
	}
}

func TestExtractKeyError_Whitespace(t *testing.T) {
	msg := "   Error with whitespace   \nNext"
	result := extractKeyError(msg)
	if result != "Error with whitespace" {
		t.Errorf("extractKeyError whitespace = %q", result)
	}
}

func TestContextBuilder_New(t *testing.T) {
	b := New(nil)
	if b == nil {
		t.Error("New(nil) should return non-nil ContextBuilder")
	}
	if b.db != nil {
		t.Error("New(nil) should set db to nil")
	}
}
