package nlquery

import (
	"regexp"
	"strings"
	"time"
)

type ExtractedEntities struct {
	Platforms []string
	JobIDs    []string
	TimeRange *TimeRange
	Metrics   []string
	Users     []string
}

type TimeRange struct {
	Start time.Time
	End   time.Time
	Expr  string
}

type EntityExtractor struct {
	platformPattern *regexp.Regexp
	jobIDPattern     *regexp.Regexp
	userPattern      *regexp.Regexp
	timePatterns     map[string]*regexp.Regexp
	metricDict       map[string]string
}

var platformMap = map[string]bool{
	"yarn":   true,
	"hive":   true,
	"spark":  true,
	"flink":  true,
	"yarn":   true,
	"spark":  true,
}

func NewEntityExtractor() *EntityExtractor {
	return &EntityExtractor{
		platformPattern: regexp.MustCompile(`(?i)(yarn|hive|spark|flink)`),
		jobIDPattern:     regexp.MustCompile(`[\w_]*(job|query|app|task)[\w_]*`),
		userPattern:      regexp.MustCompile(`(?i)user[\s:=]*(\w+)`),
		timePatterns: map[string]*regexp.Regexp{
			"hour":  regexp.MustCompile(`(?i)(最近|过去)?[\s]*(\d+)\s*(小时|小时前)`),
			"day":   regexp.MustCompile(`(?i)(最近|过去)?[\s]*(\d+)\s*(天|天前)`),
			"week":  regexp.MustCompile(`(?i)(本|上|这)?\s*周`),
			"month": regexp.MustCompile(`(?i)(本|上|这)?\s*月`),
		},
		metricDict: map[string]string{
			"内存":  "memory_used_mb",
			"CPU":   "cpu_used_cores",
			"执行时间": "duration_ms",
			"耗时":  "duration_ms",
		},
	}
}

func (e *EntityExtractor) Extract(text string) *ExtractedEntities {
	entities := &ExtractedEntities{
		Platforms: e.extractPlatforms(text),
		JobIDs:    e.extractJobIDs(text),
		Users:     e.extractUsers(text),
	}
	entities.TimeRange = e.extractTimeRange(text)
	entities.Metrics = e.extractMetrics(text)
	return entities
}

func (e *EntityExtractor) extractPlatforms(text string) []string {
Matches:
	for _, match := range e.platformPattern.FindAllString(text, -1) {
		if platformMap[strings.ToLower(match)] {
			return []string{strings.ToUpper(match)}
		}
	}
	return nil
}

func (e *EntityExtractor) extractJobIDs(text string) []string {
	var jobIDs []string
	for _, match := range e.jobIDPattern.FindAllString(text, -1) {
		if len(match) > 3 {
			jobIDs = append(jobIDs, match)
		}
	}
	return jobIDs
}

func (e *EntityExtractor) extractUsers(text string) []string {
	matches := e.userPattern.FindAllStringSubmatch(text, -1)
	var users []string
	for _, match := range matches {
		if len(match) > 1 {
			users = append(users, match[1])
		}
	}
	return users
}

func (e *EntityExtractor) extractTimeRange(text string) *TimeRange {
	for patternKey, pattern := range e.timePatterns {
		if matches := pattern.FindAllStringSubmatch(text, -1); len(matches) > 0 {
			return e.parseTimePattern(patternKey, matches[0])
		}
	}
	return e.parseTimePattern("default", nil)
}

func (e *EntityExtractor) parseTimePattern(patternKey string, matches []string) *TimeRange {
	now := time.Now()
	switch patternKey {
	case "hour":
		if len(matches) > 2 {
			return &TimeRange{
				Start: now.Add(-time.Hour),
				End:   now,
				Expr:  matches[0],
			}
		}
		return &TimeRange{
			Start: now.Add(-1 * time.Hour),
			End:   now,
			Expr:  "最近1小时",
		}
	case "day":
		if len(matches) > 2 {
			return &TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
				Expr:  matches[0],
			}
		}
		return &TimeRange{
			Start: now.Add(-24 * time.Hour),
			End:   now,
			Expr:  "最近1天",
		}
	case "week":
		weekday := int(now.Weekday())
		start := now.AddDate(0, 0, -weekday+1)
		return &TimeRange{
			Start: start,
			End:   now,
			Expr:  "本周",
		}
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return &TimeRange{
			Start: start,
			End:   now,
			Expr:  "本月",
		}
	default:
		return &TimeRange{
			Start: now.Add(-24 * time.Hour),
			End:   now,
			Expr:  "最近1天",
		}
	}
}

func (e *EntityExtractor) extractMetrics(text string) []string {
	var metrics []string
	for key, value := range e.metricDict {
		if strings.Contains(text, key) {
			metrics = append(metrics, value)
		}
	}
	return metrics
}
