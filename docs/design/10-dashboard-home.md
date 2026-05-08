# 10 - Dashboard 首页详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

Dashboard 首页是用户进入系统的第一个页面，作为"作业诊断大屏"展示系统概览和快速入口。

### 页面定位

- **入口**: 用户登录后第一个看到的页面
- **功能**: 展示系统健康状态 + 快速诊断入口
- **用户**: 运维工程师日常巡检

---

## 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ AIOps                                    [集群: 生产集群 ▼]  [用户] │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│ [Dashboard] │  作业诊断大屏                                      │
│             │  ┌─────────────────────────────────────────────┐ │
│ [Metastore] │  │  今日概览                                     │ │
│             │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐   │ │
│ [Diagnosis] ▼│  │  │ 诊断总数 │ │ 成功数   │ │ 失败数   │   │ │
│   ├ 作业诊断 │  │  │   156    │ │   144    │ │   12     │   │ │
│   └ 诊断历史 │  │  │   ↑12%  │ │  91.3%  │ │   ↓8%   │   │ │
│             │  │  └──────────┘ └──────────┘ └──────────┘   │ │
│ [AI Assistant]│  └─────────────────────────────────────────────┘ │
│             │                                                   │
│ [Settings] ▼│  ┌─────────────────────────────────────────────┐ │
│   ├ 用户管理 │  │  平台分布                                     │ │
│   ├ 集群配置 │  │  ████████████████████                    │ │
│   └ 系统配置 │  │  YARN   Spark   Hive   Flink              │ │
└─────────────┴──│  │  45      67     23      21           │─┘
                  └─────────────────────────────────────────────┘
                  ┌─────────────────────────────────────────────┐
                  │  最近诊断                                     │
                  │  ┌─────────────────────────────────────────┐│
                  │  │ spark_job_001  [Spark] FAILED   10:32  ││
                  │  │ Executor OOM                     [诊断] ││
                  │  ├─────────────────────────────────────────┤│
                  │  │ hive_query_042  [Hive] SUCCESS  10:28 ││
                  │  │ 执行成功                              [查看]││
                  │  ├─────────────────────────────────────────┤│
                  │  │ yarn_app_089    [YARN] SUCCESS  10:15 ││
                  │  └─────────────────────────────────────────┘│
                  │  [查看全部历史 →]                              │
                  └─────────────────────────────────────────────┘
```

---

## 组件设计

### 1. 统计卡片 (StatCard)

```tsx
interface StatCardProps {
    title: string;
    value: number | string;
    trend?: {
        direction: 'up' | 'down';
        percentage: number;
    };
    color?: string;
}
```

**样式**

| 状态 | 数值颜色 |
|------|----------|
| 诊断总数 | #2563EB (主色) |
| 成功数 | #10B981 (绿色) |
| 失败数 | #EF4444 (红色) |
| 成功率 | 根据数值变化 |

**布局**: 三列网格，间距 16px

### 2. 平台分布条形图

```tsx
interface PlatformBarChartProps {
    data: {
        platform: Platform;
        count: number;
        percentage: number;
    }[];
}
```

**样式**

| 平台 | 颜色 |
|------|------|
| YARN | #FF6B6B |
| Spark | #4ECDC4 |
| Hive | #FFE66D |
| Flink | #45B7D1 |

**交互**: Hover 显示具体数值

### 3. 最近诊断列表

```tsx
interface RecentDiagnosisListProps {
    items: RecentDiagnosisItem[];
    onViewAll: () => void;
}

interface RecentDiagnosisItem {
    jobId: string;
    platform: Platform;
    status: 'SUCCESS' | 'FAILED' | 'RUNNING';
    timestamp: Date;
    rootCause?: string;
}
```

**列表项样式**

| 状态 | 左侧色条 |
|------|----------|
| SUCCESS | #10B981 (绿色) |
| FAILED | #EF4444 (红色) |
| RUNNING | #2563EB (蓝色) |

**操作按钮**: 失败项显示"诊断"按钮，成功项显示"查看"按钮

---

## API 接口

### 获取首页数据

```go
// GET /api/v1/dashboard/home
type DashboardHomeResponse struct {
    Date          string         `json:"date"` // 当前日期
    Cluster       string         `json:"cluster"`
    TodayStats    TodayStats    `json:"today_stats"`
    PlatformDist  []PlatformCount `json:"platform_dist"`
    RecentJobs    []RecentJob    `json:"recent_jobs"`
}

type TodayStats struct {
    TotalCount   int   `json:"total_count"`
    SuccessCount int   `json:"success_count"`
    FailedCount  int   `json:"failed_count"`
    SuccessRate  float64 `json:"success_rate"`
    Trends       Trends `json:"trends"`
}

type Trends struct {
    TotalChange   float64 `json:"total_change"`   // 百分比
    SuccessChange float64 `json:"success_change"`
    FailedChange  float64 `json:"failed_change"`
}

type PlatformCount struct {
    Platform   string  `json:"platform"`
    Count      int     `json:"count"`
    Percentage float64 `json:"percentage"`
}

type RecentJob struct {
    JobID      string   `json:"job_id"`
    Platform   string   `json:"platform"`
    Status     string   `json:"status"`
    Timestamp  string   `json:"timestamp"`
    RootCause  string   `json:"root_cause,omitempty"`
}
```

---

## 页面状态

### 加载状态

```
┌─────────────────────────────────────────────────────────────┐
│  作业诊断大屏                                    [加载中...] │
│                                                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                │
│  │  ████    │ │  ████    │ │  ████    │                │
│  │  ████    │ │  ████    │ │  ████    │                │
│  └──────────┘ └──────────┘ └──────────┘                │
│                                                             │
│  ┌────────────────────────────────────────┐                │
│  │            ████████████               │                │
│  └────────────────────────────────────────┘                │
│                                                             │
│  ┌────────────────────────────────────────┐                │
│  │  ████████████████████████████          │                │
│  └────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

### 空状态

```
┌─────────────────────────────────────────────────────────────┐
│  作业诊断大屏                                              │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                       │   │
│  │           今日暂无诊断记录                            │   │
│  │                                                       │   │
│  │           开始第一次诊断吧                             │   │
│  │                                                       │   │
│  │           [去诊断 →]                                 │   │
│  │                                                       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 错误状态

```
┌─────────────────────────────────────────────────────────────┐
│  作业诊断大屏                                    [重试]      │
│                                                             │
│  ⚠️ 数据加载失败                                            │
│     请检查网络连接后重试                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 响应式设计

| 断点 | 布局变化 |
|------|----------|
| ≥1200px | 三列统计卡片 + 平台分布 + 最近诊断 |
| 768-1199px | 三列统计卡片，平台分布和最近诊断堆叠 |
| <768px | 单列统计卡片横向滚动，其他堆叠 |

---

**关联文档**
- 07-dashboard.md - Dashboard 框架设计
- 09-component-specs.md - 通用组件规范
