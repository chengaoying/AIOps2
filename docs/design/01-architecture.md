# 01 - 整体架构详细设计

**引用文档**: `../architecture-review-report.md`

本文档为架构设计索引，详细内容请参考 `architecture-review-report.md`。

---

## 核心架构决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 消息队列 | 去掉Kafka，Collector直写StarRocks | 简化架构，降低复杂度 |
| 诊断策略 | 知识库+LLM混合（知识库检索+LLM推理） | 可维护性强，覆盖面广，支持持续学习 |
| LLM安全 | 三层防护 | 输入验证+结构化Prompt+输出验证 |
| 高可用 | Collector/API无状态微服务 | 便于扩展和故障恢复 |
| 知识库 | Vector存储+全文检索 | 高效相似度匹配 |
| 自愈系统 | Phase 2 | Phase 1聚焦诊断核心价值 |

---

## 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                     AIOps 大数据智能诊断平台                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │    YARN     │    │    Hive     │    │   Spark     │  ...   │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘         │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            ▼                                    │
│              ┌─────────────────────────┐                        │
│              │     Collector Agent      │                        │
│              │  ┌─────────────────┐  │                        │
│              │  │ Plugin Registry  │  │                        │
│              │  │  Memory Queue   │  │                        │
│              │  │   WAL (Disk)    │  │                        │
│              │  │ Batch Writer    │  │                        │
│              │  │ Backpressure    │  │                        │
│              │  └─────────────────┘  │                        │
│              └──────────┬────────────┘                        │
│                         │                                      │
│                         ▼                                      │
│              ┌─────────────────────────┐                        │
│              │       StarRocks          │                        │
│              │  ┌─────────────────┐  │                        │
│              │  │  unified_job_view│  │                        │
│              │  │  job_dependency  │  │                        │
│              │  │ mv_resource_trend│  │                        │
│              │  └─────────────────┘  │                        │
│              └──────────┬────────────┘                        │
│                         │                                      │
│                         ▼                                      │
│              ┌─────────────────────────┐                        │
│              │     Diagnosis API         │                        │
│              │  ┌─────────────────┐  │                        │
│              │  │ Context Builder │  │                        │
│              │  │  Redis Cache   │  │                        │
│              │  │   LLM Caller   │  │                        │
│              │  └─────────────────┘  │                        │
│              └──────────┬────────────┘                        │
│                         │                                      │
│         ┌───────────────┼───────────────┐                     │
│         ▼               ▼               ▼                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐             │
│  │  Dashboard  │ │  Chat UI    │ │Alert Manager│             │
│  └─────────────┘ └─────────────┘ └─────────────┘             │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│                      知识库模块 (Knowledge Base)                   │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  ┌───────────────┐  ┌───────────────┐  ┌──────────────┐ │  │
│  │  │ 文档采集器    │  │  LLM分析提炼器 │  │  知识检索器   │ │  │
│  │  │ Doc Collector│  │ LLM Analyzer  │  │Retriever     │ │  │
│  │  └───────┬───────┘  └───────┬───────┘  └──────┬───────┘ │  │
│  │          │                  │                  │         │  │
│  │          ▼                  ▼                  ▼         │  │
│  │  ┌─────────────────────────────────────────────────┐   │  │
│  │  │              知识库存储 (Knowledge Store)        │   │  │
│  │  │   ┌─────────────────┐  ┌──────────────────┐    │   │  │
│  │  │   │ Vector Store   │  │  Full-text Index │    │   │  │
│  │  │   │  (向量存储)    │  │   (全文索引)     │    │   │  │
│  │  │   └─────────────────┘  └──────────────────┘    │   │  │
│  │  └─────────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 核心数据流

### 1. 数据采集流
```
YARN/Hive/Spark/Flink
    → Collector Plugin (REST API/SDK)
    → Memory Queue (10000条上限)
    → WAL (磁盘持久化)
    → Batch Writer (每5秒或1000条)
    → StarRocks
```

### 2. 知识库构建流
```
文档来源
    ├── 内部文档 (Confluence/Wiki/内部Wiki)
    ├── 官方文档 (Hadoop/Spark/Hive/Flink官方文档)
    └── 运维记录 (历史诊断记录、故障报告)
            │
            ▼
    文档采集器 (Doc Collector)
            │
            ▼
    文档解析器 (PDF/MD/HTML/Doc)
            │
            ▼
    LLM分析提炼器 (LLM Analyzer)
    ├── 实体提取 (错误类型、平台、根因、解决方案)
    ├── 结构化知识 (知识卡片)
    └── 向量化 (Embedding)
            │
            ▼
    知识库存储 (Knowledge Store)
    ├── Vector Store (向量检索)
    └── Full-text Index (全文检索)
```

### 3. 诊断请求流
```
User Request
    → NL Query Parser (自然语言→结构化查询)
    → Context Builder (获取作业链、日志、相似案例)
    → Knowledge Base Retrieval (知识库检索)
            │
            ▼
    ┌─────────────────────────────────┐
    │  知识库检索 (Vector + Keyword)  │
    │  1. 错误模式 → 向量相似度匹配  │
    │  2. 关键词 → 全文检索         │
    │  3. 融合排序                   │
    └─────────────────────────────────┘
            │
            ▼
    LLM推理 (LLM Reasoner)
    ├── 构建诊断Prompt (作业信息+知识上下文)
    └── 生成诊断结果 (根因+建议)
            │
            ▼
    Response (根因+建议+置信度+参考知识)
```

### 4. 告警触发流
```
异常检测 (Trend Analysis)
    → Alert Manager
    → 渠道分发 (钉钉/飞书/企微/邮件)
    → Sidebar Notification
```

---

## 服务划分

| 服务 | 语言 | 职责 | 状态 |
|------|------|------|------|
| Collector Agent | Go | 数据采集、WAL缓冲、批量写入 | Phase 1 |
| Diagnosis API | Go | 知识库检索、LLM调用、缓存 | Phase 1 |
| Knowledge Base API | Go | 知识库管理、文档采集、索引 | Phase 1 |
| Alert Manager | Go | 告警聚合、渠道分发 | Phase 1 |
| Dashboard | React | 诊断表单、结果展示 | Phase 1 |
| Chat UI | React | 自然语言交互 | Phase 1 |

---

## 知识库模块 (Knowledge Base)

### 知识来源

| 来源类型 | 示例 | 采集方式 |
|----------|------|----------|
| 内部运维文档 | 故障处理手册、作业调优指南 | Confluence API / GitBook API |
| 官方技术文档 | Hadoop YARN Docs, Spark Documentation | Web Crawler |
| 历史诊断记录 | 成功诊断案例、用户反馈 | 内部数据库 |
| 告警处理记录 | 告警处理流程、解决方案 | 内部系统 |

### 知识卡片结构

```go
type KnowledgeCard struct {
    ID          string    `json:"id"`
    Platform    string    `json:"platform"`     // YARN/HIVE/SPARK/FLINK
    ErrorType  string    `json:"error_type"`  // 错误类型
    ErrorPattern []string `json:"error_pattern"` // 错误模式（多种表述）

    // LLM提炼的结构化知识
    RootCause  string      `json:"root_cause"`   // 根因分析
    Suggestions []Suggestion `json:"suggestions"`   // 修复建议

    // 元数据
    Source     string      `json:"source"`       // 来源文档
    Confidence float64     `json:"confidence"`   // 置信度
    Tags       []string   `json:"tags"`         // 标签

    // 向量表示
    Embedding  []float32 `json:"embedding"`

    // 时间戳
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
}

type Suggestion struct {
    Action   string `json:"action"`    // 修复动作
    Risk     string `json:"risk"`     // 低/中/高
    Detail   string `json:"detail"`   // 详细说明
    Command  string `json:"command"`  // 可执行命令（如有）
}
```

### 知识库检索流程

```go
func (kb *KnowledgeBase) Retrieve(ctx context.Context, query *DiagnosisQuery) ([]*KnowledgeCard, error) {
    // 1. 向量检索
    vectorResults, err := kb.vectorStore.Search(query.Embedding, topK)

    // 2. 关键词检索
    keywordResults, err := kb.fullTextIndex.Search(query.Keywords, topK)

    // 3. 融合排序 (RRF - Reciprocal Rank Fusion)
    fused := rrfFusion(vectorResults, keywordResults, k=60)

    // 4. 过滤与重排
    //    - 平台匹配
    //    - 错误类型匹配
    //    - 时效性重排

    return fused, nil
}
```

### LLM诊断Prompt构建

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

---

## 外部依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| Go | 1.21+ | 后端语言 |
| React | 18+ | 前端框架 |
| StarRocks | 2.5+ | 元数据仓库 |
| Redis | 7+ | 缓存 |
| Milvus/Qdrant | latest | 向量数据库 |
| Elasticsearch | latest | 全文检索 |
| 通义千问 | latest | LLM |

---

## 核心架构决策详解

### 诊断策略: 知识库+LLM

| 原方案 | 新方案 | 优势 |
|--------|--------|------|
| 规则引擎 (Trie树) | 知识库检索 | 可维护性强，无需手动编写规则 |
| 40+条内置规则 | 海量知识覆盖 | 覆盖面更广 |
| 70%规则/30%LLM | 知识库+LLM推理 | 统一架构，能力可扩展 |
| 规则更新需代码改动 | 知识库自动更新 | 响应更快 |

### 为什么用知识库替代规则？

1. **规则维护成本高**: 新错误类型需要工程师编写新规则
2. **覆盖面有限**: 40+规则无法覆盖所有场景
3. **知识传承困难**: 工程师的经验难以结构化
4. **持续学习**: 知识库可以不断积累新案例

### 知识库优势

1. **海量知识**: 可导入官方文档+内部积累
2. **智能检索**: 向量检索+关键词检索融合
3. **LLM推理**: 基于检索到的知识生成诊断
4. **持续更新**: 新案例可自动入库学习

---

**引用**: 完整架构评审见 `../architecture-review-report.md`
**关联文档**: 17-knowledge-base.md - 知识库详细设计
