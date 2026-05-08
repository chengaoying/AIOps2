# 03 - 数据采集插件详细设计

**创建时间**: 2026-05-06
**负责人**: ENG
**状态**: 待实现

---

## 概述

Phase 1 支持 4 个大数据平台的采集：
- YARN (资源管理器)
- Hive (查询引擎)
- Spark (计算引擎)
- Flink (流处理引擎)

---

## YARN Plugin

### 采集源

| 数据源 | URL | 采集内容 | 频率 |
|--------|-----|----------|------|
| ResourceManager REST API | `http://rm:8088/ws/v1/cluster` | Application 状态 | 5s |
| Application Timeline Server | `http://ats:8188/ws/v1/APPLICATION` | Container 日志、作业耗时 | 5s |

### API 端点

```
# 获取所有 Application
GET /ws/v1/cluster/apps

# 获取单个 Application 详情
GET /ws/v1/cluster/apps/{app_id}

# 获取 Application 的 Containers
GET /ws/v1/cluster/apps/{app_id}/containers

# 获取 Application 的 Attempts
GET /ws/v1/cluster/apps/{app_id}/attempts
```

### 采集字段映射

| YARN Field | JobMeta Field | 说明 |
|------------|---------------|------|
| appId | JobID | 作业ID |
| name | JobName | 作业名称 |
| state | Status | RUNNING/ACCEPTED/SUCCEEDED/FAILED/KILLED |
| startedTime | StartTime | 开始时间(ms) |
| finishedTime | EndTime | 结束时间(ms) |
| elapsedTime | DurationMs | 执行时长(ms) |
| queue | Queue | 队列名 |
| user | User | 提交用户 |
| exitCode | ExitCode | 退出码 |
| diagnostics | ErrorMsg | 错误信息 |

### Container 采集

```go
type ContainerInfo struct {
    ContainerID string
    NodeID     string
    State      string
    ExitCode   int
    Logs       string // ATS 日志 URL
}
```

### OOM 检测规则

```go
// 检测 YARN OOM
func detectYARNOOM(diagnostics string) bool {
    return strings.Contains(diagnostics, "Container killed") &&
           strings.Contains(diagnostics, "out of memory")
}
```

---

## Hive Plugin

### 采集源

| 数据源 | 方式 | 采集内容 | 频率 |
|--------|------|----------|------|
| HiveServer2 API | JDBC/Thrift | 查询计划、执行日志 | 事件触发 |
| Hive Hook | Hook 回调 | 语义错误、SerDe 冲突 | 实时 |

### HS2 API 端点

```python
# 获取查询列表
SHOW QUERIES

# 获取查询详情
SELECT * FROM sys.query_data WHERE query_id = '{query_id}'

# 获取查询日志
SELECT * FROM sys.query_log WHERE query_id = '{query_id}'
```

### HiveQL 错误分类

| 错误类型 | 错误模式 | 检测关键词 |
|----------|----------|------------|
| 语义错误 | SemanticException | column not found, table not found |
| SerDe 冲突 | SerDeException | SerDe mismatch |
| 内存不足 | OutOfMemoryError | Java heap space |
| 权限错误 | AuthorizationException | permission denied |

### Hook 采集

```java
// Hive Hook 配置
<property>
    <name>hive.exec.hooks</name>
    <value>com.AIOps.hooks.DiagnosisHook</value>
</property>

// Hook 实现
public class DiagnosisHook extends AbstractSemanticAnalyzeHook {
    @Override
    public void run(Context context) {
        // 采集查询计划、错误信息
    }
}
```

### 采集字段映射

| Hive Field | JobMeta Field |
|------------|---------------|
| query_id | JobID |
| query_text | JobName (截取前100字符) |
| status | Status |
| start_time | StartTime |
| end_time | EndTime |
| error_message | ErrorMsg |

---

## Spark Plugin

### 采集源

| 数据源 | URL | 采集内容 | 频率 |
|--------|-----|----------|------|
| History Server | `http://spark-history:18080` | Stage/SQL 执行、Executor 日志 | 10s |
| Livy API | `http://livy:8998` | 活跃作业实时状态 | 10s |

### History Server API

```
# 获取所有应用
GET /api/v1/applications

# 获取应用详情
GET /api/v1/applications/{app_id}

# 获取 Stage 列表
GET /api/v1/applications/{app_id}/stages

# 获取 Executor 列表
GET /api/v1/applications/{app_id}/executors

# 获取 Driver 日志
GET /api/v1/applications/{app_id}/logs/driver
```

### Spark 错误分类

| 错误类型 | 检测关键词 |
|----------|------------|
| Executor OOM | OutOfMemoryError,ExecutorLost |
| Shuffle 错误 | ShuffleFetchFailed |
| Stage 失败 | Stage failed |
| Task 失败 | Task failed |
| 数据倾斜 | skewed |

### 采集字段映射

| Spark Field | JobMeta Field |
|-------------|---------------|
| id | JobID |
| name | JobName |
| status | Status |
| startTime | StartTime |
| endTime | EndTime |
| duration | DurationMs |

### Executor OOM 检测

```go
func detectSparkOOM(executorLogs string) bool {
    return strings.Contains(executorLogs, "OutOfMemoryError") ||
           strings.Contains(executorLogs, "ExecutorLost") &&
           strings.Contains(executorLogs, "memory")
}
```

---

## Flink Plugin

### 采集源

| 数据源 | URL | 采集内容 | 频率 |
|--------|-----|----------|------|
| REST API | `http://flink:8081` | JobManager/TaskManager 状态 | 10s |
| Metrics API | `http://flink:8081/metrics` | Checkpoint、内存使用 | 10s |

### REST API 端点

```
# 获取所有作业
GET /jobs

# 获取作业详情
GET /jobs/{job_id}

# 获取 Checkpoint 详情
GET /jobs/{job_id}/checkpoints

# 获取 TaskManager 列表
GET /taskmanagers
```

### Flink 错误分类

| 错误类型 | 检测关键词 |
|----------|------------|
| Checkpoint 超时 | Checkpoint timeout |
| TM 内存 | TaskManager memory |
| Kafka 超时 | Kafka timeout |
| Job 取消 | Job cancelled |
| Task 失败 | Task execution failed |

### Checkpoint 采集

```go
type CheckpointInfo struct {
    JobID         string
    CheckpointID  int64
    Status        string  // COMPLETED/FAILED/TIMED_OUT
    EndToEndDuration int64 // ms
    StateSize     int64   // bytes
    AlignmentBuffered int64 // bytes
}
```

---

## 统一数据格式

所有插件输出统一的 `JobMeta` 结构：

```go
type JobMeta struct {
    // 基础信息
    JobID         string
    Platform      string  // YARN/HIVE/SPARK/FLINK
    JobName       string
    Status        string  // RUNNING/SUCCESS/FAILED/KILLED/ACCEPTED
    StartTime     time.Time
    EndTime       time.Time
    DurationMs    int64

    // YARN 特有
    Queue         string
    Priority      string
    ContainerIDs  []string

    // Hive 特有
    QueryText     string
    QueryPlan     string

    // Spark 特有
    StageIDs      []int
    ExecutorIDs   []string

    // Flink 特有
    CheckpointInfo *CheckpointInfo

    // 通用
    ExitCode      int
    ErrorMsg      string
    Logs          []string
    Metrics       map[string]float64
    DependencyJobIDs []string
}
```

---

## 配置管理

```yaml
plugins:
  yarn:
    enabled: true
    api_url: "http://resourcemanager:8088"
    ats_url: "http://ats:8188"
    user: "hadoop"
    interval: 5s
    timeout: 30s

  hive:
    enabled: true
    hs2_host: "hiveserver2"
    hs2_port: 10000
    database: "default"
    hook_enabled: true
    hook_url: "http://hive-hook:8080"

  spark:
    enabled: true
    history_server: "http://spark-history:18080"
    livy_url: "http://livy:8998"
    user: "spark"
    interval: 10s
    timeout: 30s

  flink:
    enabled: true
    rest_url: "http://flink:8081"
    metrics_enabled: true
    interval: 10s
    timeout: 30s
```

---

**关联文档**
- 02-collector-agent.md - Collector Agent 框架
- 04-diagnosis-engine.md - 诊断引擎（使用采集数据）
