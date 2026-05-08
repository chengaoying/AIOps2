# 13 - 诊断历史页面详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

诊断历史页面展示所有历史诊断记录，支持筛选、搜索和详情查看。

### 页面定位

- **入口**: Sidebar "作业诊断 → 诊断历史"
- **功能**: 查看历史诊断记录、复诊、导出
- **用户**: 运维工程师回顾历史问题

---

## 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ AIOps                                    [集群: 生产集群 ▼]  [用户] │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│ 📊 Dashboard│  诊断历史                                         │
│             │  ┌─────────────────────────────────────────────┐ │
│ 🗄️ 元仓    │  │  [作业诊断] [诊断历史●]                      │ │
│             │  └─────────────────────────────────────────────┘ │
│ 🔍 作业诊断 ▼│  ┌─────────────────────────────────────────────┐ │
│   ├ 作业诊断 │  │  筛选: [平台▼] [状态▼] [时间范围▼]          │ │
│   └ 诊断历史 │  │  搜索: [诊断ID/作业ID...            ] [导出] │ │
│             │  └─────────────────────────────────────────────┘ │
│ 💬 AI助手 │  ┌─────────────────────────────────────────────┐ │
│             │  │  ┌─────────────────────────────────────┐   │ │
│ ⚙️ 系统配置▼│  │  │ 诊断ID        │平台│状态│时长│诊断时间│ │   │
│   ├ 用户管理 │  │  ├─────────────────────────────────────┤   │ │
│   ├ 集群配置 │  │  │ DX-20260506-001│Spark│88%│2.3s│10:32│ │   │
│   └ 系统配置 │  │  │ DX-20260506-002│Hive │95%│0.8s│10:28│ │   │
└─────────────┴──│  │ DX-20260506-003│YARN │75%│5.1s│10:15│ │   │
                  │  └─────────────────────────────────────┘   │ │
                  │  ┌─────────────────────────────────────────┐│ │
                  │  │ < 1 2 3 ... 10 >        共 1,234 条   ││ │
                  │  └─────────────────────────────────────────┘│ │
                  └─────────────────────────────────────────────┘
```

---

## 筛选栏

### 筛选条件

| 条件 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| 平台 | 多选 | 全部 | YARN/Hive/Spark/Flink |
| 状态/置信度 | 多选 | 全部 | 按置信度范围筛选 |
| 时间范围 | 选择器 | 最近7天 | 快捷选项+自定义 |
| 搜索 | 文本 | - | 诊断ID/作业ID/根因关键词 |

### 搜索建议

```
输入关键词时显示建议:
├─ spark_job_001 (作业ID)
├─ Executor OOM (根因关键词)
├─ DX-20260506-001 (诊断ID)
└─ hive_query_042 (作业ID)
```

---

## 历史记录表格

### 列表项

| 列名 | 字段 | 宽度 | 说明 |
|------|------|------|------|
| 诊断ID | diagnosis_id | 150px | 可点击复制 |
| 作业ID | job_id | 150px | 可点击跳转到该作业 |
| 平台 | platform | 80px | PlatformBadge |
| 置信度 | confidence | 80px | 带颜色标识 |
| 诊断耗时 | duration | 80px | 格式: X.Xs |
| 诊断时间 | timestamp | 140px | 格式: MM-DD HH:mm |
| 操作 | actions | 120px | 查看/复诊按钮 |

### 行样式

- Hover: #FAFAFA 背景
- 置信度颜色条: 左侧 3px

| 置信度范围 | 颜色 |
|-----------|------|
| >= 90% | #10B981 (绿色) |
| 70-89% | #F59E0B (橙色) |
| < 70% | #EF4444 (红色) |

---

## 诊断详情抽屉

### 点击"查看"按钮时从右侧滑出

```
┌─────────────────────────────────────────────────────────────────┐
│  诊断详情                                    [×]                 │
│  ┌───────────────────────────────────────────────────────┐   │
│  │ 诊断信息                                               │   │
│  │ 诊断ID: DX-20260506-001                              │   │
│  │ 作业ID: spark_job_001      [Spark]                   │   │
│  │ 诊断时间: 2026-05-06 10:32:00                        │   │
│  │ 诊断耗时: 2.3s                                        │   │
│  ├───────────────────────────────────────────────────────┤   │
│  │ 诊断结果                                               │   │
│  │                                                       │   │
│  │ 🔍 根因分析                                           │   │
│  │ Executor内存不足导致OOM，数据倾斜导致部分Executor...     │   │
│  │                                                       │   │
│  │ 📋 修复建议                                           │   │
│  │ 1. 增加executor内存 (4g→6g) [低风险]                 │   │
│  │ 2. 解决数据倾斜 [中风险]                               │   │
│  │                                                       │   │
│  │ 📊 诊断类型                                           │   │
│  │ [规则诊断] [AI诊断]                                    │   │
│  ├───────────────────────────────────────────────────────┤   │
│  │ 上游作业: hive_query_001 (SUCCESS)                  │   │
│  │ 下游作业: spark_job_002 (RUNNING)                   │   │
│  ├───────────────────────────────────────────────────────┤   │
│  │ 操作                                                   │   │
│  │ [复诊] [复制作业ID] [导出报告]                        │   │
│  └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 批量操作

### 多选功能

```
┌─────────────────────────────────────────────────────────────┐
│  ✓ 已选择 3 项                        [批量导出] [批量删除]  │
└─────────────────────────────────────────────────────────────┘
```

### 批量导出

支持导出格式:
- JSON (完整数据)
- CSV (表格数据)
- PDF (诊断报告)

---

## API 接口

### 获取诊断历史

```go
// GET /api/v1/diagnosis/history
type GetHistoryRequest struct {
    Platform  []string `form:"platform"`   // 逗号分隔
    Confidence string   `form:"confidence"` // e.g. "70-89"
    Start     string   `form:"start"`      // 开始时间
    End       string   `form:"end"`        // 结束时间
    Search    string   `form:"search"`    // 搜索关键词
    Page      int      `form:"page"`
    PageSize  int      `form:"page_size"`
}

type GetHistoryResponse struct {
    Records    []DiagnosisRecord `json:"records"`
    Pagination Pagination        `json:"pagination"`
}

type DiagnosisRecord struct {
    DiagnosisID string    `json:"diagnosis_id"`
    JobID       string    `json:"job_id"`
    Platform    string    `json:"platform"`
    Confidence  float64   `json:"confidence"`
    DurationMs  int64     `json:"duration_ms"`
    Timestamp   time.Time `json:"timestamp"`
    RootCause   string    `json:"root_cause"`
    DiagnosisType string   `json:"diagnosis_type"` // rule/llm
}
```

### 获取诊断详情

```go
// GET /api/v1/diagnosis/{diagnosis_id}
type GetDiagnosisResponse struct {
    DiagnosisID  string          `json:"diagnosis_id"`
    JobID        string          `json:"job_id"`
    Platform     string          `json:"platform"`
    Confidence   float64         `json:"confidence"`
    DurationMs   int64           `json:"duration_ms"`
    Timestamp    time.Time       `json:"timestamp"`
    RootCause    string          `json:"root_cause"`
    Suggestions  []Suggestion     `json:"suggestions"`
    Context      *DiagnosisContext `json:"context,omitempty"`
    DiagnosisType string         `json:"diagnosis_type"`
}
```

### 复诊

```go
// POST /api/v1/diagnosis/{diagnosis_id}/re-diagnosis
type ReDiagnosisResponse struct {
    TaskID string `json:"task_id"` // 用于轮询进度
}
```

### 导出

```go
// POST /api/v1/diagnosis/history/export
type ExportRequest struct {
    Format    string   `form:"format"`    // json/csv/pdf
    RecordIDs []string `form:"record_ids"` // 空表示全部
}

type ExportResponse struct {
    DownloadURL string `json:"download_url"`
    ExpiresAt   int64  `json:"expires_at"`
}
```

---

## 状态设计

### 加载状态

表格骨架屏，分页器禁用态

### 空状态

```
┌─────────────────────────────────────────────────────────────┐
│  筛选: [全部平台▼] [全部状态▼] [最近7天▼]             │
│  搜索: [诊断ID/作业ID...                            ]       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│              暂无诊断历史记录                                  │
│                                                             │
│              开始第一次诊断吧                                  │
│                                                             │
│              [去诊断 →]                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 错误状态

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│              ⚠️ 数据加载失败                                  │
│                                                             │
│              [重试]                                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 响应式设计

| 断点 | 表格列变化 |
|------|-------------|
| ≥1200px | 完整7列 |
| 768-1199px | 隐藏"诊断耗时"列 |
| <768px | 卡片式列表，折叠非关键信息 |

---

**关联文档**
- 12-diagnosis-job.md - 作业诊断页面
- 07-dashboard.md - Dashboard 框架
- 09-component-specs.md - 通用组件规范
