# 02 - Collector Agent 详细设计

**创建时间**: 2026-05-06
**负责人**: ENG
**状态**: 待实现

---

## 概述

Collector Agent 是数据采集的核心组件，负责从 YARN/Hive/Spark/Flink 采集作业元数据，并通过 WAL 机制保证故障恢复。

---

## 核心接口

### Collector Interface

```go
type Collector interface {
    // 采集插件名称
    Name() string

    // 初始化插件（读取配置、建立连接）
    Init(ctx context.Context, cfg PluginConfig) error

    // 采集单个作业的元数据
    Collect(ctx context.Context, jobID string) (*JobMeta, error)

    // 采集所有活跃作业
    CollectAll(ctx context.Context) ([]*JobMeta, error)

    // 健康检查
    Health(ctx context.Context) error
}
```

### JobMeta 数据结构

```go
type JobMeta struct {
    JobID         string            // 作业ID
    Platform      string            // 平台类型: YARN/Hive/Spark/Flink
    JobName       string            // 作业名称
    Status        string            // 状态: RUNNING/SUCCESS/FAILED/KILLED
    StartTime     time.Time         // 开始时间
    EndTime       time.Time         // 结束时间
    DurationMs    int64             // 执行时长(毫秒)
    SubmitTime    time.Time         // 提交时间
    Priority      string            // 优先级
    User          string            // 提交用户
    Queue         string            // 队列
    ExitCode      int               // 退出码
    ErrorMsg      string            // 错误信息
    Logs          []string          // 日志片段
    Metrics       map[string]float64 // 指标
    DependencyJobIDs []string       // 依赖作业ID列表
    RawData       map[string]any    // 原始数据(用于调试)
}
```

---

## 插件注册表 (Plugin Registry)

### 结构

```go
type Registry struct {
    plugins map[string]Collector
    mu      sync.RWMutex
}

func (r *Registry) Register(name string, plugin Collector) error
func (r *Registry) Get(name string) (Collector, error)
func (r *Registry) List() []string
func (r *Registry) InitAll(ctx context.Context, configs map[string]PluginConfig) error
```

### 配置格式

```yaml
collector:
  plugins:
    yarn:
      enabled: true
      api_url: "http://rm:8088"
      ats_url: "http://ats:8188"
      interval: 5s
    hive:
      enabled: true
      hs2_url: "http://hs2:10000"
      hook_enabled: true
    spark:
      enabled: true
      history_server: "http://spark-history:18080"
      livy_url: "http://livy:8998"
      interval: 10s
    flink:
      enabled: true
      rest_url: "http://flink:8081"
      metrics_enabled: true
```

---

## WAL 缓冲机制

### 内存队列

```go
type MemoryQueue struct {
    maxSize int
    data    []*JobMeta
    mu      sync.Mutex
    cond    *sync.Cond
}
```

- **上限**: 10000 条
- **溢出策略**: 背压控制（丢弃最旧数据）

### WAL 文件格式

```
WAL-{timestamp}-{sequence}.wal
├── Header (magic + version)
├── Records[]
│   ├── RecordType (Insert/Update/Delete)
│   ├── Timestamp
│   ├── Platform
│   ├── JobMeta JSON
│   └── Checksum
└── Footer (record count + checksum)
```

### 写入流程

```
Collect() → MemoryQueue.Enqueue()
              ↓
         [队列已满?] → 是 → WAL.Write() + DequeueOldest()
              ↓ 否
         [达到批次阈值?] → 是 → BatchWrite()
              ↓ 否
         [定时器触发?] → 是 → BatchWrite()
```

### 批量写入器

```go
type BatchWriter struct {
    batchSize    int           // 1000条
    flushInterval time.Duration // 5秒
    writer        io.Writer
}
```

---

## 背压控制

### 触发条件

| 条件 | 阈值 | 动作 |
|------|------|------|
| WAL 文件大小 | > 1GB | 丢弃最旧数据 |
| 内存队列 | > 10000 条 | 丢弃最旧数据 |
| 写入失败 | 连续3次 | 告警 + 降级 |

### 降级策略

1. **WAL 溢出**: 丢弃最旧 10% 数据，写入标记
2. **StarRocks 不可用**: 数据保留在 WAL，告警
3. **恢复后**: 从 WAL 读取并回放

---

## 健康检查

### 检查项

| 检查项 | 频率 | 失败动作 |
|--------|------|----------|
| Plugin 连接 | 30s | 标记为 DOWN |
| 内存队列使用率 | 10s | > 80% 告警 |
| WAL 文件大小 | 1min | > 1GB 告警 |
| Batch Writer | 每次写入后 | 失败重试3次 |

### 健康状态输出

```json
{
  "status": "healthy",
  "plugins": {
    "yarn": {"status": "up", "last_collect": "2026-05-06T10:00:00Z"},
    "spark": {"status": "up", "last_collect": "2026-05-06T10:00:05Z"}
  },
  "queue": {"size": 150, "max": 10000},
  "wal": {"size_mb": 500, "max_mb": 1024}
}
```

---

## 错误处理

### 错误分类

| 错误类型 | 处理策略 |
|----------|----------|
| 临时网络错误 | 重试 3 次，指数退避 |
| 认证失败 | 告警，标记插件为 DOWN |
| 数据解析错误 | 记录错误日志，跳过该条 |
| StarRocks 写入失败 | 数据保留在 WAL，触发降级 |

### 重试配置

```go
type RetryConfig struct {
    MaxAttempts int           // 3
    InitialDelay time.Duration // 1s
    MaxDelay    time.Duration // 30s
    Multiplier  float64       // 2.0
}
```

---

## 监控指标

| 指标 | 类型 | 描述 |
|------|------|------|
| collector_jobs_collected_total | Counter | 采集作业总数 |
| collector_queue_size | Gauge | 当前队列大小 |
| collector_wal_size_bytes | Gauge | WAL 文件大小 |
| collector_batch_write_duration_seconds | Histogram | 批量写入耗时 |
| collector_plugin_errors_total | Counter | 插件错误数 |

---

**关联文档**
- 03-data-plugins.md - 各平台采集插件详细设计
- architecture-review-report.md - 架构评审
