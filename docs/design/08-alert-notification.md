# 08 - 告警通知详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

告警通知系统负责在检测到异常时及时通知相关人员，支持多种渠道。

### 告警流程

```
异常检测 (Trend Analysis Engine)
    ↓
Alert Manager
    ├── 告警聚合 (去重、分组)
    ├── 告警评估 (严重程度)
    └── 渠道分发
          ├── 钉钉
          ├── 飞书
          ├── 企业微信
          └── 邮件
    ↓
Sidebar Notification (Web UI)
```

---

## 告警规则

### 触发条件

| 规则 | 条件 | 严重程度 |
|------|------|----------|
| 作业失败 | status = FAILED | warning |
| 连续失败 | 同一作业 3 次/小时 | critical |
| 性能退化 | > baseline * 1.5 | warning |
| 资源异常 | > avg * 2.0 | warning |
| LLM 降级 | 规则引擎降级 | info |

### 告警内容模板

```json
{
    "alert_id": "ALT-20260506-001",
    "platform": "SPARK",
    "job_id": "spark_job_001",
    "type": "JOB_FAILURE",
    "severity": "critical",
    "title": "[Critical] Spark 作业连续失败",
    "message": "spark_job_001 在过去 1 小时内失败 3 次",
    "timestamp": "2026-05-06T10:00:00Z",
    "diagnosis": {
        "root_cause": "Executor OOM",
        "suggestions": [
            {"action": "增加 executor memory", "risk": "low"}
        ]
    }
}
```

---

## 渠道配置

### 钉钉

```yaml
dingtalk:
  enabled: true
  webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
  secret: "SECxxx"
  at_mobiles:
    - "13800138000"
  is_at_all: false
```

### 飞书

```yaml
feishu:
  enabled: true
  webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
```

### 企业微信

```yaml
wechat:
  enabled: true
  webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
```

### 邮件

```yaml
email:
  enabled: true
  smtp_host: "smtp.example.com"
  smtp_port: 465
  smtp_user: "alert@example.com"
  smtp_password: "xxx"
  from: "AIOps Alert <alert@example.com>"
  to:
    - "ops@example.com"
```

---

## Alert Manager API

### 创建告警规则

```go
// POST /api/v1/alerts/rules
type CreateAlertRuleRequest struct {
    Name        string      `json:"name"`
    Platform    string      `json:"platform"`   // YARN/HIVE/SPARK/FLINK/ALL
    Condition   AlertCondition `json:"condition"`
    Severity    string      `json:"severity"`   // info/warning/critical
    Channels    []string    `json:"channels"`   // dingtalk/feishu/wechat/email
    Enabled     bool        `json:"enabled"`
}

type AlertCondition struct {
    Type    string      `json:"type"`    // job_failure/consecutive_failure/performance_degradation
    Params  map[string]any `json:"params"`
}
```

### 获取告警列表

```go
// GET /api/v1/alerts
type GetAlertsRequest struct {
    Platform string `json:"platform"`
    Severity string `json:"severity"`
    Status   string `json:"status"`   // active/resolved
    Start    string `json:"start"`
    End      string `json:"end"`
    Limit    int    `json:"limit"`
}
```

### 告警响应

```go
// POST /api/v1/alerts/{alert_id}/acknowledge
type AcknowledgeRequest struct {
    User    string `json:"user"`
    Comment string `json:"comment"`
}
```

---

## Sidebar 告警入口

### 告警徽章

```
┌─────────┐
│ 诊断    │
├─────────┤
│ 历史    │
├─────────┤
│ 作业    │
├─────────┤
│ 集群    │
├─────────┤
│ 告警 🔔 │  ← 告警数量徽章
├─────────┤
│ 设置    │
└─────────┘
```

### 告警徽章显示规则

| 状态 | 颜色 | 数字 |
|------|------|------|
| 无告警 | - | 不显示 |
| info | 蓝色 | 显示数量 |
| warning | 橙色 | 显示数量 |
| critical | 红色 | 显示数量 + 闪烁 |

### 告警列表弹窗

```
┌────────────────────────────────────────────────────┐
│ 告警列表                                    [全部] ▼│
├────────────────────────────────────────────────────┤
│ 🔴 [Critical] Spark 作业连续失败              10:00 │
│     spark_job_001 过去1小时失败3次              [处理]│
├────────────────────────────────────────────────────┤
│ 🟠 [Warning] 性能退化                            09:30 │
│     spark_job_002 执行时间超过基线 150%          [处理]│
├────────────────────────────────────────────────────┤
│ 🔵 [Info] LLM 降级                              09:00 │
│     spark_job_003 使用规则引擎                 [忽略]│
└────────────────────────────────────────────────────┘
```

### 告警处理操作

| 操作 | 说明 |
|------|------|
| 查看详情 | 跳转诊断页面 |
| 处理 | 标记为已处理 |
| 忽略 | 静默此告警 |
| 设置 | 跳转告警设置 |

---

## Webhook 告警格式

### 钉钉格式

```json
{
    "msgtype": "markdown",
    "markdown": {
        "title": "[Critical] Spark 作业连续失败",
        "text": "## [Critical] Spark 作业连续失败\n\n" +
                "- 作业ID: spark_job_001\n" +
                "- 失败次数: 3次/小时\n" +
                "- 根因: Executor OOM\n" +
                "- 建议: 增加 executor memory\n\n" +
                "[查看详情](http://AIOps.example.com/diagnosis/spark_job_001)"
    },
    "at": {
        "atMobiles": ["13800138000"],
        "isAtAll": false
    }
}
```

### 飞书格式

```json
{
    "msg_type": "interactive",
    "card": {
        "header": {
            "title": {
                "tag": "plain_text",
                "content": "[Critical] Spark 作业连续失败"
            },
            "template": "red"
        },
        "elements": [
            {
                "tag": "div",
                "text": {
                    "tag": "lark_md",
                    "content": "**作业ID**: spark_job_001\n" +
                              "**失败次数**: 3次/小时"
                }
            }
        ]
    }
}
```

---

## 告警聚合策略

### 去重规则

| 时间窗口 | 合并规则 |
|----------|----------|
| 5 分钟 | 同一作业 + 同一错误类型 |
| 1 小时 | 同一作业 |
| 24 小时 | 同一作业 + 同一根因 |

### 告警升级

| 条件 | 升级 |
|------|------|
| 30 分钟未处理 | info → warning |
| 1 小时未处理 | warning → critical |
| 24 小时未处理 | 发送升级通知 |

---

**关联文档**
- 04-diagnosis-engine.md - 诊断引擎（触发告警）
- 06-chat-ui.md - Chat UI（展示告警）
