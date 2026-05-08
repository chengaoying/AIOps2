package cache

import (
	"context"
	"testing"
	"time"

	"aiops2/diagnosis-api/internal/engine"
)

func TestBuildCacheKey(t *testing.T) {
	tests := []struct {
		jobID   string
		want    string
	}{
		{"spark_001", "diagnosis:spark_001"},
		{"hive_query_042", "diagnosis:hive_query_042"},
		{"yarn_app_089", "diagnosis:yarn_app_089"},
		{"", "diagnosis:"},
	}

	for _, tt := range tests {
		got := BuildCacheKey(tt.jobID)
		if got != tt.want {
			t.Errorf("BuildCacheKey(%q) = %q, want %q", tt.jobID, got, tt.want)
		}
	}
}

func TestCache_BuildCacheKey_Format(t *testing.T) {
	key := BuildCacheKey("test_job")

	if key == "" {
		t.Error("BuildCacheKey should not return empty string")
	}

	if len(key) < 10 {
		t.Error("BuildCacheKey seems too short for a proper cache key")
	}
}

func TestCache_BuildCacheKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	jobIDs := []string{"job1", "job2", "job3", "job1", "job2"}

	for _, jobID := range jobIDs {
		key := BuildCacheKey(jobID)
		keys[key] = true
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 unique keys, got %d", len(keys))
	}
}
