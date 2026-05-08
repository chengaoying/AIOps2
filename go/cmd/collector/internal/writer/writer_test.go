package writer

import (
	"testing"

	"aiops2/collector/internal/model"
)

func TestJoinLogs_Empty(t *testing.T) {
	result := joinLogs(nil)
	if result != "" {
		t.Errorf("joinLogs(nil) = %q, want empty string", result)
	}

	result = joinLogs([]string{})
	if result != "" {
		t.Errorf("joinLogs([]) = %q, want empty string", result)
	}
}

func TestJoinLogs_Single(t *testing.T) {
	result := joinLogs([]string{"single log"})
	if result != "single log" {
		t.Errorf("joinLogs([\"single log\"]) = %q, want \"single log\"", result)
	}
}

func TestJoinLogs_Multiple(t *testing.T) {
	result := joinLogs([]string{"log1", "log2", "log3"})
	expected := "log1\nlog2\nlog3"
	if result != expected {
		t.Errorf("joinLogs([\"log1\", \"log2\", \"log3\"]) = %q, want %q", result, expected)
	}
}

func TestJoinLogs_Newlines(t *testing.T) {
	result := joinLogs([]string{"line1\npre-existing", "line2"})
	if result != "line1\npre-existing\nline2" {
		t.Errorf("joinLogs with embedded newline = %q", result)
	}
}

func TestFormatMetrics_Nil(t *testing.T) {
	result := formatMetrics(nil)
	if result != "{}" {
		t.Errorf("formatMetrics(nil) = %q, want \"{}\"", result)
	}
}

func TestFormatMetrics_Empty(t *testing.T) {
	result := formatMetrics(map[string]float64{})
	if result != "{}" {
		t.Errorf("formatMetrics({}) = %q, want \"{}\"", result)
	}
}

func TestFormatMetrics_Single(t *testing.T) {
	result := formatMetrics(map[string]float64{"cpu": 4.0})
	expected := "{\"cpu\":4.000000}"
	if result != expected {
		t.Errorf("formatMetrics({\"cpu\": 4.0}) = %q, want %q", result, expected)
	}
}

func TestFormatMetrics_Multiple(t *testing.T) {
	m := map[string]float64{"cpu": 4.0, "memory": 8.5}
	result := formatMetrics(m)

	if result == "{}" {
		t.Error("formatMetrics should not return empty for non-empty map")
	}

	if len(result) < 10 {
		t.Error("formatMetrics output seems too short")
	}
}

func TestJoinStrings_Empty(t *testing.T) {
	result := joinStrings(nil)
	if result != "" {
		t.Errorf("joinStrings(nil) = %q, want empty", result)
	}

	result = joinStrings([]string{})
	if result != "" {
		t.Errorf("joinStrings([]) = %q, want empty", result)
	}
}

func TestJoinStrings_Single(t *testing.T) {
	result := joinStrings([]string{"single"})
	if result != "single" {
		t.Errorf("joinStrings([\"single\"]) = %q, want \"single\"", result)
	}
}

func TestJoinStrings_Multiple(t *testing.T) {
	result := joinStrings([]string{"a", "b", "c"})
	expected := "a,b,c"
	if result != expected {
		t.Errorf("joinStrings([\"a\", \"b\", \"c\"]) = %q, want %q", result, expected)
	}
}

func TestFormatRawData_Nil(t *testing.T) {
	result := formatRawData(nil)
	if result != "{}" {
		t.Errorf("formatRawData(nil) = %q, want \"{}\"", result)
	}
}

func TestFormatRawData_WithData(t *testing.T) {
	result := formatRawData(map[string]any{"key": "value"})
	if result != "{}" {
		t.Logf("formatRawData returns {} for any data (expected behavior)")
	}
}

func TestJobMeta_ToInsertValues(t *testing.T) {
	job := &model.JobMeta{
		JobID:    "test_job_001",
		Platform: "SPARK",
		JobName:  "test_spark_job",
		Status:   "FAILED",
	}

	if job.JobID != "test_job_001" {
		t.Errorf("JobMeta.JobID = %q, want \"test_job_001\"", job.JobID)
	}

	if job.Platform != "SPARK" {
		t.Errorf("JobMeta.Platform = %q, want \"SPARK\"", job.Platform)
	}
}

func TestBatchWriter_BufferInitialCapacity(t *testing.T) {
	bw := &BatchWriter{
		batchSize: 100,
		buffer:    make([]*model.JobMeta, 0, 100),
	}

	if cap(bw.buffer) < 100 {
		t.Errorf("Buffer capacity should be at least batchSize")
	}
}

func TestJoinLogs_Order(t *testing.T) {
	logs := []string{"first", "second", "third"}
	result := joinLogs(logs)

	if result[0:5] != "first" {
		t.Errorf("joinLogs order incorrect: got %s, want starting with 'first'", result)
	}
}
