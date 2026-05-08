# 04 - 诊断引擎详细设计

**创建时间**: 2026-05-06
**更新**: 2026-05-06 (v2 - 知识库架构)
**状态**: 待实现

---

## 概述

诊断引擎是 Phase 1 的核心，负责在 30 秒内给出根因分析和修复建议。

### 核心架构

**原方案**: 规则引擎 (Trie树) + LLM降级
**新方案**: 知识库检索 + LLM推理

```
异常作业 → 上下文构建 → [知识库检索] → LLM推理 → 结果
                                      ↓
                               [降级规则] → 结果 (知识库无命中时)
```

### 性能目标

| 场景 | 响应时间 | 说明 |
|------|----------|------|
| 知识库检索 | < 100ms | 向量+关键词混合检索 |
| LLM推理 | < 30s | 含知识上下文 |
| 端到端诊断 | < 30s | 目标 |

---

## 诊断流程

### 完整诊断流程

```go
func (e *DiagnosisEngine) Diagnose(ctx context.Context, req *DiagnosisRequest) (*DiagnosisResult, error) {
    // 1. 上下文构建
    diagCtx, err := e.contextBuilder.Build(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. 知识库检索
    knowledgeCards, err := e.knowledgeBase.Retrieve(ctx, &RetrieveRequest{
        Platform:  req.Platform,
        ErrorMsg:  req.ErrorMsg,
        TopK:      5,
    })
    if err != nil {
        return nil, err
    }

    // 3. LLM推理
    result, err := e.llmReasoner.Reason(ctx, &ReasonRequest{
        Job:      req,
        Context:  diagCtx,
        Knowledge: knowledgeCards,
    })
    if err != nil {
        // 降级到规则
        return e.ruleEngine.Fallback(req)
    }

    // 4. 更新知识库使用统计
    go e.updateKnowledgeUsage(knowledgeCards, result)

    return result, nil
}
```

### 降级策略

```
知识库检索无命中
    │
    ├── 知识库为空 → 触发首次采集
    │
    ├── 检索结果置信度低 (< 0.5)
    │       │
    │       └──► 降级到规则引擎
    │
    └── LLM调用失败
            │
            ├── 重试 3 次
            │       │
            │       ├── 成功 → 返回结果
            │       └── 失败 → 降级到规则引擎
            │
            └── 规则引擎匹配 → 返回结果
```

---

## 上下文构建

### Context Builder

```go
type ContextBuilder struct {
    db     *sql.DB
    redis  *redis.Client
}

type DiagnosisContext struct {
    Job           *JobMeta       // 当前作业
    JobChain      []*JobMeta     // 上下游作业链
    SimilarCases  []*JobMeta     // 相似案例
    RelatedLogs   []string       // 相关日志摘要
    Metrics       map[string]any // 关键指标
    ErrorPatterns []string       // 错误模式
}
```

### 上下文构建流程

```go
func (b *ContextBuilder) Build(ctx context.Context, req *DiagnosisRequest) (*DiagnosisContext, error) {
    // 1. 获取作业详情
    job, err := b.getJobDetails(req.JobID)
    if err != nil {
        return nil, err
    }

    // 2. 获取上下游作业链
    jobChain, err := b.getJobChain(job)
    if err != nil {
        return nil, err
    }

    // 3. 获取相似历史案例
    similarCases, err := b.findSimilarCases(job)
    if err != nil {
        return nil, err
    }

    // 4. 提取错误模式
    errorPatterns := b.extractErrorPatterns(job.ErrorMsg)

    // 5. 获取关键指标
    metrics := b.getKeyMetrics(job)

    return &DiagnosisContext{
        Job:          job,
        JobChain:     jobChain,
        SimilarCases: similarCases,
        ErrorPatterns: errorPatterns,
        Metrics:      metrics,
    }, nil
}
```

### 错误模式提取

```go
func (b *ContextBuilder) extractErrorPatterns(errorMsg string) []string {
    // 使用LLM提取多种错误表述
    prompt := fmt.Sprintf(`从以下错误信息中提取核心错误模式：

错误信息: %s

请提取3-5种不同的错误表述方式，包括：
1. 原始错误信息
2. 简化后的关键词
3. 可能的根因关键词

以JSON数组格式返回。`, errorMsg)

    response, err := b.llm.Call(prompt)
    // 解析并返回
}
```

---

## 知识库检索

### 检索接口

```go
type KnowledgeBase interface {
    Retrieve(ctx context.Context, req *RetrieveRequest) ([]*KnowledgeCard, error)
}

type RetrieveRequest struct {
    Platform   string
    ErrorMsg   string
    ErrorPatterns []string  // 多表述错误模式
    TopK       int
}

type KnowledgeCard struct {
    ID           string
    Platform     string
    ErrorType   string
    ErrorPatterns []string
    RootCause   string
    Suggestions []Suggestion
    Confidence float64
    Source     SourceInfo
    UsageCount int
    VoteScore  int
}
```

### 混合检索

```go
func (kb *KnowledgeBase) Retrieve(ctx context.Context, req *RetrieveRequest) ([]*KnowledgeCard, error) {
    // 1. 构建query embedding
    queryEmbedding, err := kb.embeddingModel.Encode(req.ErrorMsg)
    if err != nil {
        return nil, err
    }

    // 2. 并行执行向量检索和关键词检索
    var vectorResults, keywordResults []*KnowledgeCard
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        vectorResults, _ = kb.vectorStore.Search(queryEmbedding, req.TopK*2)
    }()

    go func() {
        defer wg.Done()
        keywordResults, _ = kb.fullTextIndex.Search(req.ErrorMsg, req.TopK*2)
    }()

    wg.Wait()

    // 3. RRF融合
    fused := kb.rrfFusion(vectorResults, keywordResults, k=60)

    // 4. 平台过滤
    filtered := kb.filterByPlatform(fused, req.Platform)

    // 5. 返回TopK
    return filtered[:min(len(filtered), req.TopK)], nil
}
```

### RRF融合

```go
func (kb *KnowledgeBase) rrfFusion(r1, r2 []*KnowledgeCard, k int) []*KnowledgeCard {
    scores := make(map[string]float64)

    // 向量检索得分
    for i, card := range r1 {
        scores[card.ID] += 1.0 / float64(k + i + 1)
    }

    // 关键词检索得分
    for i, card := range r2 {
        scores[card.ID] += 1.0 / float64(k + i + 1)
    }

    // 排序
    sort.Slice(scores, func(i, j int) bool {
        return scores[i] > scores[j]
    })

    // 返回排序后的卡片
    result := make([]*KnowledgeCard, 0)
    for id := range scores {
        card := kb.getCardByID(id)
        result = append(result, card)
    }
    return result
}
```

---

## LLM推理

### 三层安全防护

#### 第一层：输入验证

```go
type InputValidator struct {
    maxLength     int
    allowedFields []string
    blockedPatterns []string
}

func (v *InputValidator) Validate(input *LLMInput) error {
    if len(input.Prompt) > v.maxLength {
        return ErrInputTooLong
    }

    for _, pattern := range v.blockedPatterns {
        if strings.Contains(input.Prompt, pattern) {
            return ErrBlockedPattern
        }
    }

    return nil
}
```

#### 第二层：结构化Prompt

```go
const DiagnosisPromptTemplate = `你是一个大数据平台诊断专家。根据以下作业信息和知识库内容，分析根因并给出修复建议。

## 作业信息
- 平台: {{.Platform}}
- 作业ID: {{.JobID}}
- 作业名称: {{.JobName}}
- 状态: {{.Status}}
- 错误信息: {{.ErrorMsg}}
- 执行时长: {{.Duration}}ms

## 上下文信息
{{.Context}}

## 知识库参考
{{range .KnowledgeCards}}
### 知识卡片: {{.ErrorType}}
根因: {{.RootCause}}
建议:
{{range .Suggestions}}
- {{.Action}} ({{.Risk}}): {{.Detail}}
{{end}}
---
{{end}}

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
```

#### 第三层：输出验证

```go
type OutputValidator struct {
    schema *jsonschema.Schema
}

func (v *OutputValidator) Validate(output string) (*DiagnosisResult, error) {
    var result DiagnosisResult
    if err := json.Unmarshal([]byte(output), &result); err != nil {
        return nil, ErrInvalidJSON
    }

    if result.RootCause == "" {
        return nil, ErrMissingRootCause
    }

    if result.Confidence < 0 || result.Confidence > 1 {
        return nil, ErrInvalidConfidence
    }

    return &result, nil
}
```

### 通义千问集成

```go
type QwenClient struct {
    apiKey   string
    endpoint string
    model    string
    client   *http.Client
}

func (c *QwenClient) Call(ctx context.Context, prompt string) (string, error) {
    req := &QwenRequest{
        Model: c.model,
        Input: QwenInput{
            Text: prompt
        },
        Parameters: QwenParameters{
            Temperature: 0.7,
            MaxTokens: 500,
        },
    }

    resp, err := c.client.Post(c.endpoint, req)
    if err != nil {
        return "", err
    }

    return resp.Output.Text, nil
}
```

---

## 降级规则引擎

### 保留规则引擎作为降级方案

```go
type RuleEngine struct {
    trie *TrieNode
    rules []*Rule
}

// 降级触发条件
func (e *RuleEngine) ShouldFallback(err error) bool {
    return errors.Is(err, ErrKnowledgeBaseEmpty) ||
           errors.Is(err, ErrLLMTimeout) ||
           errors.Is(err, ErrLLMRateLimit)
}

// 规则匹配（降级时使用）
func (e *RuleEngine) Match(job *JobMeta) *DiagnosisResult {
    key := buildSearchKey(job)
    node := e.trie.Search(key)
    matchedRules := node.CollectRules()

    if len(matchedRules) == 0 {
        return &DiagnosisResult{
            RootCause:  "未找到匹配规则",
            Confidence: 0.0,
            Status:     "no_match",
        }
    }

    return buildResult(matchedRules[0])
}
```

### 内置降级规则（精简版）

| 规则ID | 错误模式 | 根因 | 建议 | 置信度 |
|--------|----------|------|------|--------|
| SPARK-OOM | OutOfMemoryError | Executor内存不足 | 增加executor memory | 0.85 |
| SPARK-SHUFFLE | ShuffleFetchFailed | Shuffle失败 | 增加partition | 0.80 |
| HIVE-MEM | OutOfMemoryError | 内存不足 | 增加heap size | 0.85 |
| FLINK-CHECKPOINT | Checkpoint timeout | 超时 | 增加timeout | 0.80 |

---

## 缓存与限流

### Redis缓存

```go
type Cache struct {
    client *redis.Client
    ttl    time.Duration
}

func (c *Cache) Get(key string) (*DiagnosisResult, error) {
    val, err := c.client.Get(key)
    if err == redis.Nil {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    var result DiagnosisResult
    json.Unmarshal([]byte(val), &result)
    return &result, nil
}

func (c *Cache) Set(key string, result *DiagnosisResult) error {
    val, _ := json.Marshal(result)
    return c.client.Set(key, val, c.ttl)
}
```

### 令牌桶限流

```go
type RateLimiter struct {
    bucket   chan struct{}
    rate     float64
    capacity int
}

func NewRateLimiter(qps float64, capacity int) *RateLimiter {
    limiter := &RateLimiter{
        bucket:   make(chan struct{}, capacity),
        rate:     qps,
        capacity: capacity,
    }
    go limiter.fill()
    return limiter
}

func (l *RateLimiter) Allow() bool {
    select {
    case l.bucket <- struct{}{}:
        return true
    default:
        return false
    }
}
```

---

## API接口

### 诊断请求

```go
// POST /api/v1/diagnosis
type DiagnosisRequest struct {
    Platform  string `json:"platform"`
    JobID     string `json:"job_id"`
    ErrorMsg  string `json:"error_msg,omitempty"` // 可选，提供则直接检索知识库
    UseCache  bool   `json:"use_cache"`
    ForceLLM  bool   `json:"force_llm"`
}

type DiagnosisResponse struct {
    JobID       string           `json:"job_id"`
    Status      string           `json:"status"`
    RootCause   string          `json:"root_cause"`
    Confidence  float64         `json:"confidence"`
    Suggestions []Suggestion     `json:"suggestions"`
    Context     *DiagnosisContext `json:"context,omitempty"`
    References  []string        `json:"references"`  // 知识卡片ID列表
    UsedCache   bool            `json:"used_cache"`
    UsedLLM     bool            `json:"used_llm"`
    Fallback    bool            `json:"fallback"`     // 是否降级
    DurationMs  int64          `json:"duration_ms"`
}
```

### 知识检索

```go
// POST /api/v1/diagnosis/knowledge-retrieve
type KnowledgeRetrieveRequest struct {
    Platform string `json:"platform"`
    ErrorMsg string `json:"error_msg"`
    TopK     int    `json:"top_k"`
}

type KnowledgeRetrieveResponse struct {
    Cards []*KnowledgeCard `json:"cards"`
}
```

---

## 监控指标

| 指标 | 类型 | 描述 |
|------|------|------|
| diagnosis_requests_total | Counter | 诊断请求总数 |
| diagnosis_kb_hits_total | Counter | 知识库命中次数 |
| diagnosis_kb_misses_total | Counter | 知识库未命中次数 |
| diagnosis_llm_calls_total | Counter | LLM调用次数 |
| diagnosis_llm_failures_total | Counter | LLM失败次数 |
| diagnosis_fallback_total | Counter | 降级到规则次数 |
| diagnosis_duration_seconds | Histogram | 诊断耗时 |
| diagnosis_cache_hits_total | Counter | 缓存命中次数 |

---

## 知识库集成

诊断引擎依赖知识库模块，详见 [17-knowledge-base.md](./17-knowledge-base.md)

### 依赖关系

```
DiagnosisEngine
    │
    ├──► ContextBuilder (内部)
    ├──► KnowledgeBase (外部依赖)
    │         │
    │         ├──► VectorStore (Milvus/Qdrant)
    │         ├──► FullTextIndex (Elasticsearch)
    │         └──► MetadataStore (StarRocks)
    │
    ├──► LLMReasoner (内部)
    │         │
    │         └──► LLM Caller (通义千问)
    │
    └──► RuleEngine (降级用)
```

---

**关联文档**
- 01-architecture.md - 系统架构（含知识库模块）
- 05-nl-query-parser.md - NL Query Parser
- 06-chat-ui.md - Chat UI
- 17-knowledge-base.md - 知识库详细设计
