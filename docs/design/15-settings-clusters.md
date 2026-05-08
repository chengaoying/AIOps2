# 15 - 集群配置页面详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

集群配置页面管理 YARN、Hive、Spark、Flink 四种平台集群的连接参数，用于元仓数据采集和作业诊断元数据采集。

### 页面定位

- **入口**: Sidebar "系统配置 → 集群配置"
- **功能**: 集群配置、连接测试、元仓数据采集
- **用户**: 系统管理员、运维工程师

---

## 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ AIOps                                    [集群: 生产集群 ▼]  [用户] │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│ 📊 Dashboard│  集群配置                                         │
│             │  ┌─────────────────────────────────────────────┐ │
│ 🗄️ 元仓    │  │  [集群列表]                                   │ │
│             │  └─────────────────────────────────────────────┘ │
│ 💬 AI助手 │  ┌─────────────────────────────────────────────┐ │
│             │  │  [添加集群]  [批量导入] [导出]               │ │
│ ⚙️ 系统配置▼│  └─────────────────────────────────────────────┘ │
│   ├ 用户管理 │  ┌─────────────────────────────────────────────┐ │
│   ├ 集群配置 │  │  ┌─────────────────────────────────────┐   │ │
│   └ 系统配置 │  │  │ 平台    │集群名  │状态  │操作       │   │ │
└─────────────┴──│  ├─────────────────────────────────────┤   │ │
                  │  │ YARN   │prod-yarn│在线│编辑测试删除│   │ │
                  │  │ Hive   │prod-hive│在线│编辑测试删除│   │ │
                  │  │ Spark  │prod-spark│警告│编辑测试删除│   │ │
                  │  │ Flink  │prod-flink│离线│编辑测试删除│   │ │
                  │  └─────────────────────────────────────┘   │ │
                  └─────────────────────────────────────────────┘
```

---

## 集群列表

### 平台说明

| 平台 | 采集内容 | 连接方式 |
|------|----------|----------|
| YARN | 作业运行信息、ApplicationMaster日志、容器资源使用 | REST API / RPC |
| Hive | 查询作业、查询计划、执行阶段、日志 | JDBC / REST API |
| Spark | 作业运行信息、Stage DAG、Executor日志、Shuffle数据 | REST API / Spark UI |
| Flink | 作业状态、TaskManager资源、JobManager日志、检查点 | REST API / Flink UI |

### 表格列定义

| 列名 | 字段 | 宽度 | 说明 |
|------|------|------|------|
| 平台 | platform | 100px | YARN/Hive/Spark/Flink |
| 集群名称 | name | 150px | 唯一标识 |
| 连接地址 | endpoint | 200px | WebUI/REST API 地址 |
| 状态 | status | 100px | 在线/离线/警告 |
| 操作 | actions | 180px | 编辑/测试/删除 |

### 状态标识

| 状态 | 颜色 | 说明 |
|------|------|------|
| 在线 | #10B981 (绿色) | 连接正常 |
| 离线 | #EF4444 (红色) | 无法连接 |
| 警告 | #F59E0B (橙色) | 连接延迟>5s |
| 未配置 | #9CA3AF (灰色) | 未添加该平台集群 |

---

## 添加/编辑集群弹窗

### 公共字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 平台 | 下拉选择 | 是 | YARN/Hive/Spark/Flink |
| 集群名称 | 输入框 | 是 | 唯一标识, 2-50字符 |
| 连接超时 | 数字输入 | 否 | 默认30秒 |
| 采集间隔 | 数字输入 | 否 | 默认60秒 |
| 描述 | 文本域 | 否 | 备注信息 |

### YARN 集群配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ResourceManager 地址 | 输入框 | 是 | e.g. rm-host:8088 |
| 历史服务器地址 | 输入框 | 否 | JobHistory Server e.g. hs-host:19888 |
| WebUI 入口 | 输入框 | 否 | 作业查看入口 |

### Hive 集群配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| HiveServer2 地址 | 输入框 | 是 | e.g. hs2-host:10000 |
| Metastore 地址 | 输入框 | 是 | e.g. meta-host:9083 |
| JDBC URL | 输入框 | 是 | jdbc:hive2://hs2-host:10000 |
| 用户名 | 输入框 | 是 | 连接用户名 |
| 密码 | 密码框 | 是 | 连接密码(加密存储) |

### Spark 集群配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Spark Master URL | 输入框 | 是 | e.g. spark://master-host:7077 |
| History Server 地址 | 输入框 | 否 | e.g. hs-host:18080 |
| WebUI 端口 | 数字输入 | 否 | 默认4040 |

### Flink 集群配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| JobManager 地址 | 输入框 | 是 | e.g. jm-host:8081 |
| REST API 地址 | 输入框 | 是 | e.g. http://jm-host:8081 |
| 集群 ID | 输入框 | 否 | YARN/Kubernetes 集群ID |

---

## 连接测试

在添加/编辑弹窗中提供"测试连接"按钮，点击后测试当前配置。

### 测试成功

```
┌─────────────────────────────────────────────────────────────┐
│  连接测试                                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                     │   │
│  │  ✓ ResourceManager 连接成功 (耗时: 0.8s)           │   │
│  │  ✓ JobHistory Server 连接成功 (耗时: 1.2s)          │   │
│  │                                                     │   │
│  │  状态: 正常                                         │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 测试失败

```
┌─────────────────────────────────────────────────────────────┐
│  连接测试                                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                     │   │
│  │  ✗ ResourceManager 连接失败: Connection timeout    │   │
│  │  ✗ JobHistory Server 连接成功                       │   │
│  │                                                     │   │
│  │  状态: 部分异常                                     │   │
│  │  建议: 检查 ResourceManager 网络连通性              │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 状态设计

### 加载状态

表格骨架屏 + 状态指示器占位

### 空状态

```
┌─────────────────────────────────────────────────────────────┐
│  筛选: [全部平台▼] [全部状态▼]                         │
│  搜索: [集群名称...                                   ]       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│              暂无集群配置                                    │
│                                                             │
│              请先添加集群以启动元仓数据采集                   │
│                                                             │
│              [添加第一个集群 →]                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 删除确认

```
┌─────────────────────────────────────────────────────────────┐
│  ⚠️ 确认删除                                               │
│                                                             │
│  确定要删除集群 "prod-yarn" 吗？                            │
│  该操作只会删除集群配置, 不会删除已采集的元仓数据。          │
│                                                             │
│                          [取消] [确认删除]                   │
└─────────────────────────────────────────────────────────────┘
```

---

## API 接口

### 获取集群列表

```go
// GET /api/v1/settings/clusters
type GetClustersRequest struct {
    Platform string `form:"platform"`  // YARN/HIVE/SPARK/FLINK
    Status   string `form:"status"`    // online/offline/warning
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
}

type GetClustersResponse struct {
    Clusters   []Cluster   `json:"clusters"`
    Pagination Pagination  `json:"pagination"`
}

type Cluster struct {
    ID           int64     `json:"id"`
    Platform     string    `json:"platform"`     // YARN/HIVE/SPARK/FLINK
    Name         string    `json:"name"`
    Endpoint     string    `json:"endpoint"`     // 连接地址
    Status       string    `json:"status"`      // online/offline/warning
    Description  string    `json:"description"`
}
```

### 创建集群

```go
// POST /api/v1/settings/clusters
type CreateClusterRequest struct {
    Platform    string `json:"platform"`
    Name        string `json:"name"`
    Endpoint    string `json:"endpoint"`
    Config      map[string]string `json:"config"` // 平台特定配置
    Timeout     int    `json:"timeout"`     // 秒
    Interval    int    `json:"interval"`    // 采集间隔(秒)
    Description string `json:"description"`
}

// YARN 配置
type YARNConfig struct {
    ResourceManager string `json:"resource_manager"` // e.g. rm-host:8088
    JobHistoryServer string `json:"job_history_server"` // e.g. hs-host:19888
    WebUI           string `json:"web_ui"`
}

// Hive 配置
type HiveConfig struct {
    HS2Address     string `json:"hs2_address"`     // e.g. hs2-host:10000
    Metastore      string `json:"metastore"`      // e.g. meta-host:9083
    JDBCURL        string `json:"jdbc_url"`
    Username       string `json:"username"`
    Password       string `json:"password"`        // 加密存储
}

// Spark 配置
type SparkConfig struct {
    Master         string `json:"master"`         // e.g. spark://master-host:7077
    HistoryServer  string `json:"history_server"`  // e.g. hs-host:18080
    WebUIPort      int    `json:"web_ui_port"`     // 默认4040
}

// Flink 配置
type FlinkConfig struct {
    JobManager     string `json:"job_manager"`    // e.g. jm-host:8081
    RESTAPI        string `json:"rest_api"`        // e.g. http://jm-host:8081
    ClusterID      string `json:"cluster_id"`     // YARN/K8s集群ID
}
```

### 测试连接

```go
// POST /api/v1/settings/clusters/test
type TestConnectionRequest struct {
    Platform string `json:"platform"`
    Endpoint string `json:"endpoint"`
    Config   map[string]string `json:"config"`
    Timeout  int    `json:"timeout"`
}

type TestConnectionResponse struct {
    Success     bool    `json:"success"`
    LatencyMs   int64   `json:"latency_ms"`
    Message     string  `json:"message,omitempty"`   // 成功/失败描述
    Suggestions []string `json:"suggestions,omitempty"` // 修复建议
    Error       string   `json:"error,omitempty"`    // 错误信息
}
```

### 更新集群

```go
// PUT /api/v1/settings/clusters/{cluster_id}
type UpdateClusterRequest struct {
    Name        string `json:"name"`
    Endpoint    string `json:"endpoint"`
    Config      map[string]string `json:"config"`
    Timeout     int    `json:"timeout"`
    Interval    int    `json:"interval"`
    Description string `json:"description"`
}
```

### 删除集群

```go
// DELETE /api/v1/settings/clusters/{cluster_id}
```

---

## 响应式设计

| 断点 | 表格列变化 |
|------|-------------|
| ≥1200px | 完整5列 |
| 768-1199px | 隐藏"描述"列 |
| <768px | 卡片式列表, 折叠非关键信息 |

---

**关联文档**
- 07-dashboard.md - Dashboard 框架
- 09-component-specs.md - 通用组件规范
- 11-metastore.md - 元仓页面
- 16-settings-system.md - 系统配置页面
- 02-collector-agent.md - Collector Agent 框架
- 03-data-plugins.md - 数据采集插件