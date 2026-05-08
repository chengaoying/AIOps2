# 12 - 作业诊断页面详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

作业诊断页面提供单作业诊断功能，用户输入平台和作业ID即可快速获取诊断结果。

### 页面定位

- **入口**: Sidebar "作业诊断 → 作业诊断"
- **功能**: 快速诊断单个作业
- **用户**: 运维工程师收到告警后进行诊断

---

## 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ AIOps                                    [集群: 生产集群 ▼]  [用户] │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│ 📊 Dashboard│  作业诊断                                         │
│             │  ┌─────────────────────────────────────────────┐ │
│ 🗄️ 元仓    │  │                                             │ │
│             │  │  输入作业信息开始诊断                         │ │
│ 🔍 作业诊断 ▼│  │                                             │ │
│   ├ 作业诊断 │  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐   │ │
│   └ 诊断历史 │  │  │ YARN │ │ Hive │ │Spark │ │Flink │   │ │
│             │  │  └──────┘ └──────┘ └──────┘ └──────┘   │ │
│ 💬 AI助手 │  │                                             │ │
│             │  │  ┌─────────────────────────────────────────┐││
│ ⚙️ 系统配置▼│  │  │ 搜索: [作业名/ID...              ]      │││
│   ├ 用户管理 │  │  │         [队列▼] [任务类型▼] [🔍搜索]  │││
│   ├ 集群配置 │  │  └─────────────────────────────────────────┘││
│   └ 系统配置 │  │                                             │ │
└─────────────┴──│  ┌─────────────────────────────────────────┐││
                  │  │ 搜索结果:                               │││
                  │  │ ┌───────────────────────────────────┐ │││
                  │  │ │ ○ spark_job_001  [Spark]  prod  batch │││
                  │  │ │ ● spark_job_002  [Spark]  prod  stream│││
                  │  │ │ ○ hive_query_042   [Hive]  dev   select│││
                  │  │ └───────────────────────────────────┘ │││
                  │  │                                             │ │
                  │  │  已选择: spark_job_002  [Spark]          │ │
                  │  │                                             │ │
                  │  │           [🚀 开始诊断]                   │ │
                  │  └─────────────────────────────────────────┘ │ │
                  └─────────────────────────────────────────────┘
```

---

## 诊断表单

### 表单字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 平台 | 单选按钮组 | 是 | YARN / Hive / Spark / Flink |
| 搜索 | 输入框 + 下拉选择 | 是 | 通过作业名/ID/队列/任务类型搜索 |
| 队列 | 下拉选择 | 否 | 筛选特定队列 |
| 任务类型 | 下拉选择 | 否 | batch / streaming |

### 平台选择器

```tsx
interface PlatformSelectorProps {
    value: Platform | null;
    onChange: (platform: Platform) => void;
    options?: Platform[];
}
```

**状态**

| 状态 | 样式 |
|------|------|
| 默认 | 灰色背景 #F5F5F5，深灰文字 |
| Hover | 浅平台色背景 |
| 选中 | 平台色背景，白色文字 |

### 作业搜索

```tsx
interface JobSearchProps {
    platform: Platform;
    searchQuery: string;
    onSearchChange: (query: string) => void;
    onSelectJob: (job: JobSummary) => void;
    selectedJob?: JobSummary;
    filters: {
        queue?: string;
        taskType?: string;
    };
    onFilterChange: (filters: SearchFilters) => void;
}

interface SearchFilters {
    queue?: string;      // 队列筛选
    taskType?: string;   // 任务类型: batch/streaming
}

interface JobSummary {
    jobId: string;
    name: string;
    platform: Platform;
    queue: string;
    taskType: string;
    status: string;
}
```

**搜索交互**

```
┌─────────────────────────────────────────────────────────────┐
│  🔍 搜索: [作业名/ID/队列/任务类型                    ]  │
│                                                             │
│  输入提示:                                                  │
│  ├─ spark_job_001     (作业ID精确匹配)                    │
│  ├─ etl_batch         (作业名模糊搜索)                    │
│  ├─ queue:prod        (队列筛选)                          │
│  ├─ type:streaming    (任务类型筛选)                      │
│  └─ FAILED            (状态筛选)                          │
└─────────────────────────────────────────────────────────────┘
```

**搜索结果列表**

| 字段 | 说明 |
|------|------|
| 选择 | 单选按钮 |
| 作业ID | 可点击复制 |
| 平台 | PlatformBadge |
| 队列 | 所属队列 |
| 任务类型 | batch / streaming |
| 状态 | SUCCESS / FAILED / RUNNING |

**验证规则**

- YARN: 以 `application_` 开头
- Hive: 以 `hive_query_` 或 `query_` 开头
- Spark: 以 `spark_job_` 或 `application_` 开头
- Flink: 以 `flink_job_` 开头

**错误提示**

| 错误类型 | 提示文案 |
|----------|----------|
| 空输入 | 请输入作业ID或选择作业 |
| 未选择 | 请从搜索结果中选择一个作业 |
| 格式错误 | 作业ID格式不正确 |
| 平台不匹配 | 该作业ID与所选平台不匹配 |

### 快捷筛选

**队列筛选**

根据所选平台动态加载可用队列:

```
队列:
├─ 全部
├─ root
├─ root.prod
├─ root.dev
└─ root.test
```

**任务类型筛选**

| 平台 | 选项 |
|------|------|
| Hive | SELECT / INSERT / JOIN / AGG |
| Spark | batch / streaming |
| Flink | batch / streaming |
| YARN | - |

---

## 诊断结果展示

### 诊断进行中

```
┌─────────────────────────────────────────────────────────────┐
│  诊断进行中...                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                     │   │
│  │  ⚙️ 规则匹配中                                      │   │
│  │  ████████████████████░░░░░░░░  60%              │   │
│  │                                                     │   │
│  │  [步骤说明]                                         │   │
│  │  ✓ 1. 规则匹配                                     │   │
│  │  → 2. 上下文查询                                   │   │
│  │  ○ 3. AI分析 (如需要)                            │   │
│  │  ○ 4. 生成建议                                     │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 诊断完成

```
┌─────────────────────────────────────────────────────────────┐
│  ✓ 诊断完成                                    耗时: 2.3s │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ▌ spark_job_001  [Spark]           置信度: 88%   │   │
│  │ ─────────────────────────────────────────────────── │   │
│  │                                                     │   │
│  │ 🔍 根因分析                                         │   │
│  │ Executor内存不足导致OOM，数据倾斜导致部分Executor...    │   │
│  │                                                     │   │
│  │ 📋 修复建议                                         │   │
│  │ ┌───────────────────────────────────────────────┐ │   │
│  │ │ 1. 增加executor内存                           │ │   │
│  │ │    4g → 6g                                   │ │   │
│  │ │    [低风险] [复制命令] [查看详情]               │ │   │
│  │ ├───────────────────────────────────────────────┤ │   │
│  │ │ 2. 解决数据倾斜                               │ │   │
│  │ │    添加salting策略                            │ │   │
│  │ │    [中风险] [复制命令] [查看详情]             │ │   │
│  │ └───────────────────────────────────────────────┘ │   │
│  │                                                     │   │
│  │ 📊 上下文信息                                       │   │
│  │ 上游作业: hive_query_001 (SUCCESS)              │   │
│  │ 下游作业: spark_job_002 (RUNNING)               │   │
│  │ 相似案例: 3个类似OOM问题曾被成功诊断              │   │
│  │                                                     │   │
│  │ [查看详细上下文 →]                                 │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 诊断失败

```
┌─────────────────────────────────────────────────────────────┐
│  ✗ 诊断失败                                    [重试]      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                     │   │
│  │  ⚠️ 未找到该作业                                    │   │
│  │                                                     │   │
│  │  可能原因:                                         │   │
│  │  • 作业ID输入错误                                  │   │
│  │  • 作业数据尚未采集                                │   │
│  │  • 该作业已被清理                                  │   │
│  │                                                     │   │
│  │  [重新输入]                                       │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 组件定义

### DiagnosisForm

```tsx
interface DiagnosisFormProps {
    onSubmit: (platform: Platform, jobId: string, queue?: string, taskType?: string) => void;
    loading: boolean;
    error?: string;
}
```

### JobSearchPanel

```tsx
interface JobSearchPanelProps {
    platform: Platform;
    onSelectJob: (job: JobSummary) => void;
    selectedJob?: JobSummary;
    loading?: boolean;
}
```

### DiagnosisProgress

```tsx
interface DiagnosisProgressProps {
    step: 'rule_match' | 'context_query' | 'ai_analyze' | 'generate';
    progress: number; // 0-100
    message?: string;
}
```

### DiagnosisResult

```tsx
interface DiagnosisResultProps {
    jobId: string;
    platform: Platform;
    confidence: number;
    rootCause: string;
    suggestions: Suggestion[];
    context?: DiagnosisContext;
    duration: number;
    onSuggestionClick: (suggestion: Suggestion) => void;
}
```

### SuggestionItem

```tsx
interface SuggestionItemProps {
    index: number;
    action: string;
    detail?: string;
    risk: 'low' | 'medium' | 'high';
    onCopy?: () => void;
    onViewDetail?: () => void;
}
```

### DiagnosisContext

```tsx
interface DiagnosisContext {
    upstreamJobs: JobSummary[];
    downstreamJobs: JobSummary[];
    similarCases: SimilarCase[];
    metrics: Record<string, number>;
}
```

---

## API 接口

### 搜索作业

```go
// GET /api/v1/diagnosis/job/search
type SearchJobsRequest struct {
    Platform string `form:"platform"`  // YARN/HIVE/SPARK/FLINK
    Search   string `form:"search"`    // 作业名/ID 搜索
    Queue    string `form:"queue"`     // 队列筛选
    TaskType string `form:"task_type"`  // batch/streaming
    Status   string `form:"status"`    // SUCCESS/FAILED/RUNNING
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
}

type SearchJobsResponse struct {
    Jobs       []JobSummary `json:"jobs"`
    Pagination Pagination   `json:"pagination"`
}

type JobSummary struct {
    JobID    string `json:"job_id"`
    Name     string `json:"name"`
    Platform string `json:"platform"`
    Queue    string `json:"queue"`
    TaskType string `json:"task_type"`
    Status   string `json:"status"`
}

// GET /api/v1/diagnosis/job/queues
type GetQueuesResponse struct {
    Queues []string `json:"queues"` // 可用队列列表
}
```

### 提交诊断请求

```go
// POST /api/v1/diagnosis/job
type DiagnosisJobRequest struct {
    Platform string `json:"platform"` // YARN/HIVE/SPARK/FLINK
    JobID   string `json:"job_id"`
    Queue   string `json:"queue,omitempty"`    // 队列
    TaskType string `json:"task_type,omitempty"` // batch/streaming
}

type DiagnosisJobResponse struct {
    Status      string             `json:"status"` // running/success/failed
    JobID       string             `json:"job_id"`
    Diagnosis   *DiagnosisResult   `json:"diagnosis,omitempty"`
    Error       *DiagnosisError    `json:"error,omitempty"`
    ProgressURL string             `json:"progress_url,omitempty"` // 用于轮询
}

type DiagnosisResult struct {
    RootCause    string             `json:"root_cause"`
    Confidence   float64            `json:"confidence"`
    Suggestions  []Suggestion      `json:"suggestions"`
    Context      *DiagnosisContext  `json:"context,omitempty"`
    DurationMs   int64              `json:"duration_ms"`
}

type DiagnosisError struct {
    Code    string `json:"code"`    // JOB_NOT_FOUND/ANALYSIS_FAILED
    Message string `json:"message"`
}
```

### 获取诊断进度（长轮询）

```go
// GET /api/v1/diagnosis/{task_id}/progress
type DiagnosisProgressResponse struct {
    Step      string  `json:"step"`
    Progress  int     `json:"progress"` // 0-100
    Message   string  `json:"message"`
    Completed bool    `json:"completed"`
}
```

---

## 诊断状态流

```
用户提交
    ↓
诊断进行中 (Progress)
    ↓
┌─────────────────────────────────┐
│ 规则引擎匹配 ──→ 命中 ──→ 返回结果 │
│     ↓ 未命中                     │
│  LLM分析 (如启用)               │
│     ↓                           │
│ 返回结果 (或降级结果)             │
└─────────────────────────────────┘
    ↓
诊断完成 / 诊断失败
```

---

## 快捷操作

### 诊断完成后可执行的操作

| 操作 | 说明 |
|------|------|
| 复制建议 | 将修复命令复制到剪贴板 |
| 查看详情 | 展开建议的详细说明 |
| 查看上下文 | 跳转到作业详情页 |
| 复制作业ID | 复制作业ID便于分享 |
| 重新诊断 | 清空当前结果，重新输入 |

---

**关联文档**
- 04-diagnosis-engine.md - 诊断引擎
- 07-dashboard.md - Dashboard 框架
- 09-component-specs.md - 通用组件规范
