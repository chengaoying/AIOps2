# 05 - NL Query Parser 详细设计

**创建时间**: 2026-05-06
**负责人**: ENG
**状态**: 待实现

---

## 概述

NL Query Parser 将用户的自然语言查询转换为结构化的 SQL/API 调用，支持性能分析、资源分析、故障分析和趋势分析。

### 支持的查询类型

| 类型 | 示例问题 | 转换为 |
|------|----------|---------|
| 性能分析 | "为什么变慢了" | SQL: 执行时间趋势 |
| 资源分析 | "哪些作业最耗资源" | SQL: 资源使用排行 |
| 故障分析 | "最近有哪些失败" | SQL: 失败作业列表 |
| 趋势分析 | "本周性能变化" | SQL: 周趋势对比 |

---

## 自然语言解析流程

```
用户输入
    ↓
NL Query Parser
    ├── 意图识别 (Intent Classification)
    ├── 实体提取 (Entity Extraction)
    ├── SQL 生成 (SQL Generation)
    └── API 调用 → 返回结果
```

---

## 意图识别

### 支持的意图

| Intent | 描述 | 触发词 |
|--------|------|--------|
| PERFORMANCE_ANALYSIS | 性能分析 | 慢、变慢、性能、执行时间 |
| RESOURCE_ANALYSIS | 资源分析 | 资源、内存、CPU、耗费 |
| FAILURE_ANALYSIS | 故障分析 | 失败、错误、异常、挂了 |
| TREND_ANALYSIS | 趋势分析 | 趋势、变化、增长、本周、上周 |
| JOB_QUERY | 作业查询 | 作业、job、任务 |
| METRICS_QUERY | 指标查询 | 指标、监控 |

### 意图分类器

```go
type IntentClassifier struct {
    model *clf.Classifier
}

type Intent struct {
    Type    IntentType
    Confidence float64
    Keywords []string
}

func (c *IntentClassifier) Classify(text string) (*Intent, error) {
    // 1. 关键词匹配
    for _, intent := range c.intents {
        score := c.matchKeywords(text, intent.Keywords)
        if score > intent.Threshold {
            return &Intent{
                Type:      intent.Type,
                Confidence: score,
                Keywords:  intent.MatchedKeywords,
            }, nil
        }
    }

    // 2. 默认归类为性能分析
    return &Intent{
        Type:      PERFORMANCE_ANALYSIS,
        Confidence: 0.5,
    }, nil
}
```

---

## 实体提取

### 实体类型

| Entity Type | 示例 | 提取方式 |
|-------------|------|----------|
| PLATFORM | YARN、Hive、Spark、Flink | 枚举匹配 |
| JOB_ID | spark_job_001、hive_query_123 | 正则匹配 |
| TIME_RANGE | 最近1小时、最近1天、本周 | 时间表达式解析 |
| METRIC_NAME | 执行时间、内存使用、CPU | 词典匹配 |
| USER | user1、admin | 用户名正则 |

### 实体提取器

```go
type EntityExtractor struct {
    platformPatterns []*regexp.Regexp
    jobIDPatterns   []*regexp.Regexp
    timeExprParser  *.TimeExpressionParser
    metricDict      map[string]string
}

type ExtractedEntities struct {
    Platforms   []string
    JobIDs     []string
    TimeRange  *TimeRange
    Metrics    []string
    Users      []string
}
```

### 时间表达式解析

```go
type TimeRange struct {
    Start time.Time
    End   time.Time
    Expr  string // 原始表达式
}

func (p *TimeExpressionParser) Parse(expr string) (*TimeRange, error) {
    switch expr {
    case "最近1小时":
        return &TimeRange{
            Start: time.Now().Add(-1 * time.Hour),
            End:   time.Now(),
            Expr:  expr,
        }, nil
    case "最近1天":
        return &TimeRange{
            Start: time.Now().Add(-24 * time.Hour),
            End:   time.Now(),
            Expr:  expr,
        }, nil
    case "本周":
        // 计算本周一
        now := time.Now()
        weekday := int(now.Weekday())
        start := now.AddDate(0, 0, -weekday+1)
        return &TimeRange{
            Start: start,
            End:   now,
            Expr:  expr,
        }, nil
    // ... 更多时间表达式
    }
}
```

---

## SQL 生成

### 查询模板

#### 性能分析 SQL

```sql
-- 性能分析: 执行时间趋势
SELECT
    job_id,
    platform,
    job_name,
    duration_ms,
    start_time,
    end_time
FROM unified_job_view
WHERE platform = '{platform}'
    AND start_time >= '{start_time}'
    AND start_time <= '{end_time}'
ORDER BY start_time DESC
LIMIT 100
```

#### 资源分析 SQL

```sql
-- 资源分析: 内存使用排行
SELECT
    job_id,
    platform,
    memory_used_mb,
    cpu_used_cores,
    duration_ms
FROM job_metrics
WHERE start_time >= '{start_time}'
ORDER BY memory_used_mb DESC
LIMIT 20
```

#### 故障分析 SQL

```sql
-- 故障分析: 失败作业列表
SELECT
    job_id,
    platform,
    job_name,
    status,
    error_msg,
    exit_code,
    end_time
FROM unified_job_view
WHERE status = 'FAILED'
    AND end_time >= '{start_time}'
ORDER BY end_time DESC
LIMIT 50
```

#### 趋势分析 SQL

```sql
-- 趋势分析: 周趋势对比
SELECT
    DATE(start_time) as date,
    COUNT(*) as job_count,
    AVG(duration_ms) as avg_duration,
    MAX(duration_ms) as max_duration,
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failure_count
FROM unified_job_view
WHERE start_time >= '{start_time}'
    AND start_time <= '{end_time}'
GROUP BY DATE(start_time)
ORDER BY date DESC
```

---

## 自然语言到 SQL 示例

| 用户输入 | 意图 | 提取实体 | 生成 SQL |
|----------|------|----------|----------|
| 为什么 spark_job_001 变慢了 | PERFORMANCE_ANALYSIS | job_id=spark_job_001 | 执行时间对比查询 |
| 最近1小时哪些作业最耗内存 | RESOURCE_ANALYSIS | time_range=最近1小时, metric=内存 | 内存排行查询 |
| 今天有哪些作业失败了 | FAILURE_ANALYSIS | time_range=今天 | 失败作业列表查询 |
| 对比本周和上周的性能 | TREND_ANALYSIS | time_range=本周/上周 | 周趋势对比查询 |

---

## API 接口

### 自然语言查询

```go
// POST /api/v1/query/natural
type NaturalQueryRequest struct {
    Query string `json:"query"` // 自然语言查询
    User  string `json:"user"`  // 用户
}

type NaturalQueryResponse struct {
    SQL        string         `json:"sql"`        // 生成的 SQL
    Intent     string         `json:"intent"`     // 识别的意图
    Entities   []string       `json:"entities"`   // 提取的实体
    Results    []map[string]any `json:"results"`  // 查询结果
    ChartType  string         `json:"chart_type"` // 推荐图表类型
    DurationMs int64          `json:"duration_ms"`// 查询耗时
}
```

### SQL 查询

```go
// POST /api/v1/query/sql
type SQLQueryRequest struct {
    SQL   string `json:"sql"`
    Limit int    `json:"limit"` // 默认 100
}

type SQLQueryResponse struct {
    Columns []string         `json:"columns"`
    Rows    [][]any          `json:"rows"`
    Count   int              `json:"count"`
}
```

---

## 图表类型推荐

| 意图 | 推荐图表 | 说明 |
|------|----------|------|
| PERFORMANCE_ANALYSIS | line | 执行时间趋势 |
| RESOURCE_ANALYSIS | bar | 资源排行 |
| FAILURE_ANALYSIS | table | 失败列表 |
| TREND_ANALYSIS | line | 周趋势对比 |

---

## 错误处理

| 错误 | 处理 |
|------|------|
| 无法识别意图 | 返回错误提示，提示用户重新描述 |
| SQL 生成失败 | 回退到简单 SQL，或返回错误 |
| 数据库查询失败 | 返回错误，包含错误信息 |

---

**关联文档**
- 04-diagnosis-engine.md - 诊断引擎
- 06-chat-ui.md - Chat UI（展示查询结果）
