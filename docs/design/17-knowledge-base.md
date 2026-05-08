# 17 - 知识库详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

知识库是 AIOps 诊断系统的核心组件，通过采集、提炼、存储大数据平台的运维知识，为诊断提供依据。

### 设计目标

1. **海量知识管理**: 支持导入官方文档和内部积累
2. **智能检索**: 向量检索+关键词检索融合
3. **持续学习**: 新案例自动入库学习
4. **可解释性**: 每条诊断结果可追溯到参考知识

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      知识库模块架构                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐      ┌──────────────────┐            │
│  │   文档采集器      │      │   知识管理API    │            │
│  │   Doc Collector   │      │  Knowledge API   │            │
│  └─────────┬──────────┘      └─────────┬──────────┘            │
│            │                           │                         │
│            ▼                           ▼                         │
│  ┌──────────────────────────────────────────────────┐         │
│  │                  LLM 分析提炼器                      │         │
│  │                  LLM Analyzer                     │         │
│  │  ┌────────────────────────────────────────────┐   │         │
│  │  │ 1. 文档解析 (PDF/MD/HTML/Doc)            │   │         │
│  │  │ 2. 实体提取 (错误类型/平台/解决方案)      │   │         │
│  │  │ 3. 知识结构化 (Knowledge Card)           │   │         │
│  │  │ 4. 向量化 (Embedding Generation)           │   │         │
│  │  └────────────────────────────────────────────┘   │         │
│  └──────────────────────────┬───────────────────────┘         │
│                             │                                  │
│                             ▼                                  │
│  ┌──────────────────────────────────────────────────┐         │
│  │                  知识库存储层                      │         │
│  │  ┌─────────────────┐  ┌──────────────────────┐   │         │
│  │  │  Vector Store  │  │  Full-text Index    │   │         │
│  │  │  (Milvus)      │  │  (Elasticsearch)     │   │         │
│  │  └─────────────────┘  └──────────────────────┘   │         │
│  │  ┌─────────────────┐  ┌──────────────────────┐   │         │
│  │  │  Metadata Store │  │   Graph Store       │   │         │
│  │  │  (StarRocks)  │  │   (可选)            │   │         │
│  │  └─────────────────┘  └──────────────────────┘   │         │
│  └──────────────────────────────────────────────────┘         │
│                             │                                  │
│                             ▼                                  │
│  ┌──────────────────────────────────────────────────┐         │
│  │                  知识检索服务                      │         │
│  │  ┌────────────────────────────────────────────┐   │         │
│  │  │ Hybrid Search: Vector + Keyword + RRF     │   │         │
│  │  │ Re-ranking: Platform/ErrorType Filter      │   │         │
│  │  └────────────────────────────────────────────┘   │         │
│  └──────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 知识来源

### 文档类型

| 类型 | 来源 | 示例 | 采集方式 |
|------|------|------|----------|
| 官方文档 | Apache官方 | Hadoop YARN Troubleshooting | Web爬虫 |
| 官方文档 | Spark官方 | Spark Monitoring & Tuning | Web爬虫 |
| 官方文档 | Hive官方 | Hive Error Messages | Web爬虫 |
| 官方文档 | Flink官方 | Flink Fault Tolerance | Web爬虫 |
| 内部WIKI | Confluence | 故障处理手册 | Confluence API |
| 内部WIKI | GitBook | 运维指南 | GitBook API |
| 历史诊断 | 诊断系统 | 成功诊断案例 | 内部API |
| 告警记录 | 告警系统 | 告警处理记录 | 内部API |

### 文档格式支持

| 格式 | 解析方式 |
|------|----------|
| Markdown (.md) | 直接解析 |
| HTML | HTML解析器 |
| PDF | PDF解析器 |
| Word (.docx) | Docx解析器 |
| Confluence | Confluence API |

---

## 知识卡片 (Knowledge Card)

### 数据结构

```go
type KnowledgeCard struct {
    // 唯一标识
    ID          string    `json:"id"`          // KB-YYYYMMDD-XXXXX

    // 平台信息
    Platform    string    `json:"platform"`     // YARN/HIVE/SPARK/FLINK

    // 错误信息
    ErrorType  string    `json:"error_type"`  // 错误类型
    ErrorPatterns []string `json:"error_patterns"` // 多种错误表述（同一错误的多种说法）

    // LLM提炼的结构化知识
    RootCause   string      `json:"root_cause"`   // 根因分析
    Suggestions []Suggestion `json:"suggestions"`  // 修复建议

    // 来源信息
    Source      SourceInfo `json:"source"`      // 来源信息
    SourceDoc   string     `json:"source_doc"`  // 原始文档片段

    // 质量信息
    Confidence  float64   `json:"confidence"`   // 置信度 0-1
    UsageCount  int       `json:"usage_count"` // 被引用次数
    VoteScore   int       `json:"vote_score"`  // 用户投票分数

    // 标签
    Tags        []string   `json:"tags"`        // 标签

    // 向量表示
    Embedding  []float32 `json:"-"`           // 不暴露到API

    // 时间戳
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type SourceInfo struct {
    Type     string `json:"type"`      // official/internal/diagnosis
    Title    string `json:"title"`     // 文档标题
    URL      string `json:"url"`        // 文档链接
    Author   string `json:"author"`     // 作者
}

type Suggestion struct {
    Action   string `json:"action"`   // 修复动作
    Risk     string `json:"risk"`     // 低/中/高
    Detail   string `json:"detail"`   // 详细说明
    Command  string `json:"command"`  // 可执行命令
}
```

### 知识卡片示例

```json
{
    "id": "KB-20260506-00001",
    "platform": "SPARK",
    "error_type": "Executor OOM",
    "error_patterns": [
        "OutOfMemoryError: Executor memory exceeded",
        "ExecutorLost: executor ... exited with exit code 137",
        "Container killed by YARN for exceeding memory limits"
    ],
    "root_cause": "Executor分配的内存不足以处理数据集，导致OOM被系统Kill",
    "suggestions": [
        {
            "action": "增加executor内存",
            "risk": "低",
            "detail": "将executor-memory从4g增加到6g或更高",
            "command": "--conf spark.executor.memory=6g"
        },
        {
            "action": "优化数据分区",
            "risk": "中",
            "detail": "使用salting策略解决数据倾斜问题",
            "command": null
        }
    ],
    "source": {
        "type": "official",
        "title": "Spark Configuration - Memory Management",
        "url": "https://spark.apache.org/docs/latest/configuration.html"
    },
    "confidence": 0.92,
    "usage_count": 156,
    "vote_score": 42,
    "tags": ["memory", "oom", "executor", "spark"]
}
```

---

## LLM分析提炼流程

### 提炼Pipeline

```
原始文档
    │
    ▼
┌─────────────────────────────┐
│  1. 文档解析                │
│  - 提取标题、段落、代码块   │
│  - 识别表格和列表           │
│  - 保留格式信息             │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  2. 段落分割                │
│  - 按段落/章节分割          │
│  - 保留上下文信息           │
│  - 标注段落类型             │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  3. LLM实体提取             │
│  - 提取错误类型             │
│  - 提取错误模式             │
│  - 提取解决方案              │
│  - 提取配置参数             │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  4. 知识结构化               │
│  - 构建Knowledge Card       │
│  - 生成同义错误模式         │
│  - 生成修复建议             │
│  - 提取命令和配置           │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  5. 向量化                   │
│  - 生成Embedding            │
│  - 存储到Vector Store       │
└─────────────┬───────────────┘
              │
              ▼
        Knowledge Card
```

### LLM提炼Prompt

```go
const ExtractionPrompt = `你是一个大数据平台知识提炼专家。从以下文档片段中提取诊断知识。

## 文档片段
{{.Content}}

## 平台
{{.Platform}}

## 输出要求
请提取以下信息，以JSON格式输出：

{
    "error_type": "错误类型（英文）",
    "error_patterns": ["错误表述1", "错误表述2", "错误表述3"],
    "root_cause": "根因分析（中文，50字以内）",
    "suggestions": [
        {
            "action": "修复动作（中文）",
            "risk": "低/中/高",
            "detail": "详细说明（中文）",
            "command": "相关命令（如有）"
        }
    ],
    "tags": ["标签1", "标签2"]
}

注意事项：
1. error_patterns需要包含多种错误表述方式
2. suggestions需要包含风险等级
3. 如果文档片段不包含错误诊断信息，返回空JSON {}`
```

---

## 知识检索

### 混合检索流程

```
用户查询
    │
    ├──► 向量检索 (Vector Search)
    │         - 错误信息向量化
    │         - Top-K 相似检索
    │         - 返回Top 20
    │
    ├──► 关键词检索 (Keyword Search)
    │         - 分词
    │         - BM25排序
    │         - 返回Top 20
    │
    ▼
融合排序 (RRF Fusion)
    - k=60 (通常取60)
    - RRF_score = Σ 1/(k+rank)
    │
    ▼
平台/错误类型过滤
    │
    ▼
重排 (Re-ranking)
    - 参考知识质量
    - 时效性
    │
    ▼
返回 Top 5 知识卡片
```

### RRF融合算法

```go
func RRFusion(vectorResults, keywordResults []*SearchResult, k int) []*SearchResult {
    scores := make(map[string]float64)

    // 向量检索得分
    for i, r := range vectorResults {
        scores[r.ID] += 1.0 / float64(k + i + 1)
    }

    // 关键词检索得分
    for i, r := range keywordResults {
        scores[r.ID] += 1.0 / float64(k + i + 1)
    }

    // 排序
    sorted := sortByScore(scores)
    return sorted
}
```

---

## 文档采集器

### 采集配置

```go
type DocCollectorConfig struct {
    // 官方文档爬虫配置
    Crawlers []CrawlerConfig

    // 内部文档API配置
    InternalSources []InternalSourceConfig

    // 采集调度
    Schedule string // cron表达式
}

type CrawlerConfig struct {
    Name    string   // 名称
    URL     string   // 起始URL
    AllowedDomains []string // 允许的域名
    Depth   int      // 爬取深度

    // 解析器
    Parser string // markdown/html/pdf
}

type InternalSourceConfig struct {
    Type     string // confluence/gitbook/internal
    APIURL   string
    Token    string
    Space    string // 空间标识
    Query    string // 搜索查询
}
```

### 调度策略

| 文档类型 | 更新频率 | 说明 |
|----------|----------|------|
| 官方文档 | 每周一次 | 官方发布新版本时触发 |
| 内部WIKI | 每天一次 | 工作时间同步 |
| 历史诊断 | 实时 | 新诊断成功时自动入库 |
| 告警记录 | 每小时 | 增量同步 |

---

## API 接口

### 知识管理

```go
// GET /api/v1/knowledge/cards
type GetKnowledgeCardsRequest struct {
    Platform string `form:"platform"`   // YARN/HIVE/SPARK/FLINK
    ErrorType string `form:"error_type"`
    Search  string `form:"search"`
    Tags    string `form:"tags"`       // 逗号分隔
    Page    int    `form:"page"`
    PageSize int   `form:"page_size"`
}

type GetKnowledgeCardsResponse struct {
    Cards     []KnowledgeCard `json:"cards"`
    Total     int             `json:"total"`
    Page      int             `json:"page"`
    PageSize  int             `json:"page_size"`
}

// GET /api/v1/knowledge/cards/{card_id}
type GetKnowledgeCardResponse struct {
    Card KnowledgeCard `json:"card"`
}

// POST /api/v1/knowledge/cards
type CreateKnowledgeCardRequest struct {
    Platform      string      `json:"platform"`
    ErrorType     string     `json:"error_type"`
    ErrorPatterns []string   `json:"error_patterns"`
    RootCause     string     `json:"root_cause"`
    Suggestions   []Suggestion `json:"suggestions"`
    Source        SourceInfo `json:"source"`
    Tags          []string   `json:"tags"`
}

type CreateKnowledgeCardResponse struct {
    Card KnowledgeCard `json:"card"`
}

// PUT /api/v1/knowledge/cards/{card_id}
type UpdateKnowledgeCardRequest struct {
    RootCause   string      `json:"root_cause,omitempty"`
    Suggestions []Suggestion `json:"suggestions,omitempty"`
    Tags        []string    `json:"tags,omitempty"`
    Confidence  float64     `json:"confidence,omitempty"`
}

// DELETE /api/v1/knowledge/cards/{card_id}
type DeleteKnowledgeCardResponse struct {
    Success bool `json:"success"`
}
```

### 知识检索

```go
// POST /api/v1/knowledge/retrieve
type RetrieveKnowledgeRequest struct {
    Platform   string `json:"platform"`
    ErrorMsg   string `json:"error_msg"`
    JobContext string `json:"job_context,omitempty"`
    TopK       int    `json:"top_k"` // 默认5
}

type RetrieveKnowledgeResponse struct {
    Cards      []KnowledgeCard `json:"cards"`
    QueryEmbedding []float32 `json:"-"` // 不返回
}
```

### 文档采集

```go
// POST /api/v1/knowledge/collect
type CollectDocsRequest struct {
    SourceType string `json:"source_type"` // official/internal
    SourceName string `json:"source_name"`
    URLs       []string `json:"urls,omitempty"` // 指定URL
}

type CollectDocsResponse struct {
    TaskID     string `json:"task_id"`
    Status     string `json:"status"` // started/failed
    Message    string `json:"message"`
}

// GET /api/v1/knowledge/collect/{task_id}
type CollectTaskStatusResponse struct {
    TaskID    string `json:"task_id"`
    Status    string `json:"status"` // running/completed/failed
    Progress  int    `json:"progress"` // 0-100
    Processed int    `json:"processed"` // 已处理文档数
    Created   int    `json:"created"`   // 新增知识卡片数
    Updated   int    `json:"updated"`   // 更新知识卡片数
    Failed    int    `json:"failed"`   // 失败数
    Error     string `json:"error,omitempty"`
}
```

### 知识反馈

```go
// POST /api/v1/knowledge/cards/{card_id}/vote
type VoteCardRequest struct {
    Score int `json:"score"` // +1 / -1
}

type VoteCardResponse struct {
    NewScore int `json:"new_score"`
}

// POST /api/v1/knowledge/cards/{card_id}/report
type ReportIncorrectRequest struct {
    Reason string `json:"reason"`
    Detail string `json:"detail"`
}

type ReportIncorrectResponse struct {
    Success bool `json:"success"`
    Message string `json:"message"`
}
```

---

## 诊断集成

### 诊断服务调用知识库

```go
func (d *DiagnosisService) DiagnoseWithKnowledge(req *DiagnosisRequest) (*DiagnosisResult, error) {
    // 1. 构建检索query
    query := &RetrieveKnowledgeRequest{
        Platform: req.Platform,
        ErrorMsg: req.ErrorMsg,
        TopK:     5,
    }

    // 2. 检索知识库
    knowledgeResp, err := d.kbClient.Retrieve(query)
    if err != nil {
        return nil, err
    }

    // 3. 构建LLM Prompt
    prompt := buildDiagnosisPrompt(req, knowledgeResp.Cards)

    // 4. LLM推理
    llmResp, err := d.llm.Call(prompt)
    if err != nil {
        return nil, err
    }

    // 5. 解析结果
    result := parseDiagnosisResult(llmResp, knowledgeResp.Cards)

    // 6. 更新知识使用统计
    go d.updateUsageStats(knowledgeResp.Cards)

    return result, nil
}
```

### 诊断历史自动入库

```go
func (d *DiagnosisService) onDiagnosisSuccess(result *DiagnosisResult) {
    // 诊断成功时，自动提取知识入库
    card := &CreateKnowledgeCardRequest{
        Platform:      result.Platform,
        ErrorType:     classifyError(result.ErrorMsg),
        ErrorPatterns: []string{result.ErrorMsg},
        RootCause:     result.RootCause,
        Suggestions:    result.Suggestions,
        Source: SourceInfo{
            Type:  "diagnosis",
            Title: fmt.Sprintf("诊断案例 - %s", result.JobID),
        },
        Tags: []string{"auto-generated"},
    }

    // 异步入库
    go d.kbClient.Create(card)
}
```

---

## 存储设计

### 向量存储 (Vector Store)

使用 Milvus：

```go
type VectorStoreConfig struct {
    Type      string // milvus
    Host      string
    Port      int
    Collection string // collection name

    // 向量配置
    Dimension   int    // 1536 (text-embedding-3-small)
    MetricType string // L2/IP/COSINE
}

// Milvus 配置示例
vectorStore:
  type: milvus
  host: localhost
  port: 19530
  collection: AIOps_knowledge
  dimension: 1536
  metric_type: COSINE
```

### 全文索引 (Full-text Index)

使用 Elasticsearch：

```go
type FullTextIndexConfig struct {
    Host     string
    Port     int
    Index    string

    // 分词器
    Analyzer string // ik_max_word
}
```

### 元数据存储 (Metadata Store)

使用 StarRocks：

```sql
CREATE TABLE knowledge_cards (
    id          VARCHAR(50) PRIMARY KEY,
    platform    VARCHAR(20) NOT NULL,
    error_type  VARCHAR(100) NOT NULL,
    root_cause  TEXT,
    confidence  FLOAT,
    usage_count INT DEFAULT 0,
    vote_score  INT DEFAULT 0,
    source_type VARCHAR(20),
    source_url  VARCHAR(500),
    created_at  DATETIME,
    updated_at  DATETIME,

    INDEX idx_platform (platform),
    INDEX idx_error_type (error_type),
    INDEX idx_created (created_at)
);
```

---

## 监控指标

| 指标 | 类型 | 描述 |
|------|------|------|
| kb_cards_total | Gauge | 知识卡片总数 |
| kb_cards_by_platform | Gauge | 各平台知识卡片数 |
| kb_retrieve_total | Counter | 知识库检索次数 |
| kb_retrieve_latency | Histogram | 检索延迟 |
| kb_llm_extraction_total | Counter | LLM提炼次数 |
| kb_llm_extraction_latency | Histogram | 提炼延迟 |
| kb_collect_total | Counter | 采集任务总数 |
| kb_collect_processed | Counter | 已处理文档数 |
| kb_confidence_distribution | Histogram | 置信度分布 |

---

## 状态设计

### 空状态

```
┌─────────────────────────────────────────────────────────────┐
│  知识库为空                                               │
│                                                             │
│  知识库尚未初始化，请先采集文档                             │
│                                                             │
│  采集源:                                                   │
│  • 官方文档 (Hadoop/Spark/Hive/Flink)                    │
│  • 内部WIKI                                               │
│  • 历史诊断案例                                            │
│                                                             │
│  [开始采集官方文档 →]                                     │
└─────────────────────────────────────────────────────────────┘
```

### 知识卡片详情

```
┌─────────────────────────────────────────────────────────────┐
│  知识卡片                                    [×]             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ID: KB-20260506-00001                              │   │
│  │ 平台: [Spark]                                     │   │
│  │ 错误类型: Executor OOM                            │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 错误模式                                          │   │
│  │ • OutOfMemoryError: Executor memory exceeded       │   │
│  │ • ExecutorLost: exited with exit code 137         │   │
│  │ • Container killed by YARN for exceeding memory    │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 根因                                              │   │
│  │ Executor分配的内存不足以处理数据集，导致OOM被Kill │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 修复建议                                          │   │
│  │ 1. 增加executor内存 [低风险]                     │   │
│  │    spark.executor.memory: 4g → 6g               │   │
│  │    [复制命令]                                      │   │
│  │ 2. 优化数据分区 [中风险]                         │   │
│  │    使用salting策略解决数据倾斜                    │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 来源: Spark官方文档 - Memory Management          │   │
│  │ 置信度: 92%  被引用: 156次  评分: +42          │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 标签: [memory] [oom] [executor] [spark]          │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 操作                                              │   │
│  │ [编辑] [复制] [报告错误] [查看来源]              │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 性能目标

| 指标 | 目标 | 说明 |
|------|------|------|
| 知识检索延迟 | < 100ms | P95 |
| LLM 提炼延迟 | < 30s/文档 | 含 API 调用 |
| 知识卡片数量 | > 10000 | Phase 1 目标 |
| 检索召回率 | > 90% | 与人工诊断对比 |

## 向量数据库选型

### 为什么选择 Milvus

| 维度 | 说明 |
|------|------|
| 成熟度 | CNCF 毕业项目，1000+ 企业生产使用 |
| 规模化 | 支持 10M+ 向量，Phase 2 扩展无忧 |
| 索引类型 | 支持 HNSW/IVF/DiskANN 多种索引 |
| 生态 | 与 StarRocks 都是 CNCF 项目，集成方便 |

### Milvus 配置

```yaml
# docker-compose.yml
services:
  etcd:
    image: quay.io/coreos/etcd:v3.5.5
    environment:
      - ETCD_AUTO_COMPACTION_MODE=revision
      - ETCD_AUTO_COMPACTION_RETENTION=1000
      - ETCD_QUOTA_BACKEND_BYTES=4294967296
    volumes:
      - etcd_data:/etcd

  minio:
    image: minio/minio:latest
    environment:
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin
    volumes:
      - minio_data:/minio_data
    command: server /minio_data --console-address ":9001"

  milvus:
    image: milvusdb/milvus:v3.1.0
    ports:
      - "19530:19530"
      - "9091:9091"
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
    volumes:
      - milvus_data:/var/lib/milvus
    depends_on:
      - etcd
      - minio

volumes:
  etcd_data:
  minio_data:
  milvus_data:

---

## 扩展性设计

### 多租户支持

```go
type TenantContext struct {
    TenantID string
    UserID   string
    Roles    []string
}

// 知识卡片增加租户字段
type KnowledgeCard struct {
    // ...
    TenantID string `json:"tenant_id"`
}
```

### 知识图谱扩展 (Phase 2)

```
知识卡片
    │
    ├──► 错误类型图谱
    │         │
    │         └──► 根因关系
    │                   │
    │                   └──► 作业链路
    │
    └──► 解决方案图谱
              │
              └──► 效果评估
```

---

**关联文档**
- 01-architecture.md - 系统架构
- 04-diagnosis-engine.md - 诊断引擎
- 08-alert-notification.md - 告警通知
- PRD.md - 产品设计文档
- phase1-plan.md - Phase 1 实施计划

---

## 设计更新日志

| 日期 | 更新内容 | 说明 |
|------|----------|------|
| 2026-05-08 | 向量数据库选型 Milvus | CNCF 毕业，成熟度高，支持大规模 |
