# AIOps 详细设计文档

**创建时间**: 2026-05-06
**版本**: v1.2
**状态**: 进行中

---

## 文档索引

### 第一层：架构设计
- [01-architecture.md](./01-architecture.md) - 整体架构详细设计（含知识库模块）

### 第二层：功能模块设计
- [02-collector-agent.md](./02-collector-agent.md) - Collector Agent 框架
- [03-data-plugins.md](./03-data-plugins.md) - 4个采集插件（YARN/Hive/Spark/Flink）
- [04-diagnosis-engine.md](./04-diagnosis-engine.md) - 诊断引擎（知识库检索+LLM推理）
- [05-nl-query-parser.md](./05-nl-query-parser.md) - NL Query Parser
- [17-knowledge-base.md](./17-knowledge-base.md) - 知识库模块（文档采集/LLM提炼/混合检索）

### 第三层：页面与交互设计
- [06-chat-ui.md](./06-chat-ui.md) - Chat UI 详细设计
- [07-dashboard.md](./07-dashboard.md) - Dashboard 详细设计（含新版Sidebar）
- [08-alert-notification.md](./08-alert-notification.md) - 告警通知设计
- [09-component-specs.md](./09-component-specs.md) - 通用组件规范（含新版Sidebar组件）
- [10-dashboard-home.md](./10-dashboard-home.md) - Dashboard 首页（作业诊断大屏）
- [11-metastore.md](./11-metastore.md) - 元仓页面（作业列表/关联/趋势）
- [12-diagnosis-job.md](./12-diagnosis-job.md) - 作业诊断页面
- [13-diagnosis-history.md](./13-diagnosis-history.md) - 诊断历史页面
- [14-settings-users.md](./14-settings-users.md) - 用户管理页面
- [15-settings-clusters.md](./15-settings-clusters.md) - 集群配置页面
- [16-settings-system.md](./16-settings-system.md) - 系统配置页面

---

## Sidebar 导航结构（v1.1更新）

| 一级导航 | 二级导航 | 说明 |
|----------|----------|------|
| Dashboard | - | 作业诊断大屏（首页） |
| 元仓 | - | 元数据信息展示 |
| 作业诊断 | 作业诊断 | 单作业诊断表单 |
| | 诊断历史 | 历史诊断记录列表 |
| AI助手 | - | Chat方式诊断 |
| 系统配置 | 用户管理 | 用户权限配置 |
| | 集群配置 | YARN/Hive/Spark/Flink集群配置 |
| | 系统配置 | 告警规则、通知渠道 |

---

## 核心架构决策（v1.2更新）

| 决策 | 原方案 | 新方案 | 理由 |
|------|--------|--------|------|
| 诊断策略 | 规则引擎+Trie树 (70%) | 知识库检索+LLM推理 | 可维护性强，覆盖面广，支持持续学习 |
| 知识存储 | 无 | Vector Store + Full-text Index | 高效相似度匹配 |
| 规则引擎 | 主诊断 | 降级备用 | 仅在知识库无命中时使用 |

---

## 设计原则

1. **Utilitarian** - 功能优先，最小化装饰
2. **Platform Colors** - YARN=#FF6B6B, Hive=#FFE66D, Spark=#4ECDC4, Flink=#45B7D1
3. **Information Dense** - 高信息密度，适合运维工程师
4. **Fast Recovery** - 30秒内给出根因分析和修复建议
5. **Knowledge-Driven** - 基于知识库的智能诊断，支持持续学习

---

## 设计变更记录

| 日期 | 版本 | 变更内容 | 状态 |
|------|------|----------|------|
| 2026-05-06 | v1.0 | 初始版本 | 进行中 |
| 2026-05-06 | v1.1 | Sidebar菜单结构调整，新增元仓、AI助手，作业诊断和系统配置增加二级菜单 | 进行中 |
| 2026-05-06 | v1.2 | 知识库模块：增加17-knowledge-base.md，诊断引擎改为知识库+LLM架构，规则引擎降级为备用 | 进行中 |
