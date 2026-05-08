# TODOS.md — AIOps Phase 1 实施进度跟踪

**创建时间**: 2026-05-05
**预计周期**: 6-8周
**状态**: IN_PROGRESS

---

## 状态说明

- [ ] 未开始
- [ ] 已完成
- [x] 进行中

---

## Week 1-2: 基础设施

### 1.1 项目脚手架

- [x] 初始化Go项目，DDD目录结构
- [x] 定义统一采集接口(Collector interface)
- [x] 实现配置管理系统
- [x] 实现日志系统

### 1.2 Collector Agent框架

- [x] 实现Plugin Registry
- [x] 实现WAL缓冲机制（内存队列10000条 + 磁盘WAL）
- [x] 实现批量写入器（每5秒或1000条）
- [x] 实现背压控制（WAL超过1GB时丢弃最旧数据）

### 1.3 StarRocks环境

- [x] 搭建StarRocks开发环境
- [x] 创建作业元数据表(job_meta)
- [x] 创建诊断事件表(incidents)
- [x] 创建日志表(logs)
- [x] 创建物化视图(mv_job_failure_stats, mv_top_error_codes)

---

## Week 3-4: 数据采集+元仓

### 2.1 数据采集插件

#### YARN Plugin
- [x] 实现REST API采集
- [x] 实现ATS采集
- [x] 实现Application状态、Container日志、作业耗时采集

#### Hive Plugin
- [x] 实现HS2 API采集
- [x] 实现Hook采集
- [x] 实现HiveQL执行日志、查询计划、错误信息采集

#### Spark Plugin
- [x] 实现History Server采集
- [x] 实现Livy API采集
- [x] 实现Stage/SQL执行、Executor日志采集

#### Flink Plugin
- [x] 实现REST API采集
- [x] 实现Metrics采集
- [x] 实现JobManager/TaskManager状态、Checkpoint采集

### 2.2 元仓建设

- [x] 创建统一元数据视图(unified_job_view)
- [x] 创建作业级关联表(job_dependency)
- [x] 创建资源使用趋势视图(mv_resource_trend)

---

## Week 5-6: 诊断引擎

### 3.1 规则引擎

- [x] 实现前缀树索引
- [x] 编写YARN规则（OOM、队列满、Container启动失败等）
- [x] 编写Hive规则（语义错误、SerDe冲突、内存不足等）
- [x] 编写Spark规则（Executor OOM、Stage失败、Shuffle错误等）
- [x] 编写Flink规则（Checkpoint超时、TM内存、Kafka超时等）

### 3.2 LLM调用

- [x] 实现Prompt安全防护（三层防护）
- [x] 集成通义千问API
- [x] 实现输出解析器

### 3.3 上下文构建

- [x] 实现Context Builder
- [x] 实现作业链查询
- [x] 实现相似案例检索
- [x] 实现日志上下文获取

### 3.4 缓存与降级

- [x] 实现Redis缓存（TTL=1小时）
- [x] 实现LLM失败→规则引擎降级
- [x] 实现令牌桶限流（10 QPS）

---

## Week 7-8: AI交互+UI

### 4.1 NL Query Parser

- [x] 实现自然语言解析为SQL/API调用
- [x] 实现性能分析查询
- [x] 实现资源分析查询
- [x] 实现故障分析查询
- [x] 实现趋势分析查询

### 4.2 Performance Baseline Service

- [x] 实现定时任务（每小时）
- [x] 实现p50/p95执行时间计算
- [x] 创建job_baseline表

### 4.3 Trend Analysis Engine

- [x] 实现异常检测逻辑
- [x] 实现性能退化检测（>baseline*1.5）
- [x] 实现资源异常检测（>avg*2.0）

### 4.4 Chat UI

- [x] 实现React组件（ChatInterface、ChatHistory、InputBox、ChartDisplay）
- [x] 实现对话历史管理
- [x] 实现图表展示（ECharts）
- [x] 实现用户消息右对齐、AI消息左对齐
- [x] 实现平台图标（YARN/Hive/Spark/Flink）
- [x] 实现Chat组件扩展现有设计系统
- [x] 实现Tablet侧边栏图标模式

### 4.5 告警通知

- [x] 实现钉钉通知渠道
- [x] 实现飞书通知渠道
- [x] 实现企业微信通知渠道
- [x] 实现邮件通知渠道
- [x] 实现侧边栏告警入口

### 4.6 测试

#### 单元测试
- [ ] Collector Agent单元测试（200+用例，覆盖80%+）
- [ ] Prompt构建器测试（输入验证+转义）
- [ ] 规则引擎测试（前缀树索引匹配）

#### 集成测试
- [ ] Collector→StarRocks集成测试
- [ ] Diagnosis API→LLM集成测试
- [ ] 规则引擎匹配测试（40+条规则）
- [ ] 告警通知集成测试（钉钉/飞书/企业微信/邮件）

#### E2E测试
- [ ] Spark OOM故障诊断全流程
- [ ] Hive语义错误诊断全流程
- [ ] 告警触发诊断流程
- [ ] LLM降级到规则引擎
- [ ] StarRocks不可用时WAL缓冲

---

## Phase 2 技术债务（暂不实施）

以下功能推迟到Phase 2:

- [ ] 多租户隔离
- [ ] 内置可观测性（Grafana Dashboard）
- [ ] 自愈/自动化修复
- [ ] SQLScan规则扫描
- [ ] Kafka/HDFS/HBase扩展组件采集
- [ ] 语音输入

---

## 进度统计

### 当前进度

| 阶段 | 总任务数 | 已完成 | 进行中 | 未开始 |
|------|----------|--------|--------|--------|
| Week 1-2: 基础设施 | 12 | 12 | 0 | 0 |
| Week 3-4: 数据采集+元仓 | 13 | 13 | 0 | 0 |
| Week 5-6: 诊断引擎 | 13 | 13 | 0 | 0 |
| Week 7-8: AI交互+UI | 20 | 17 | 0 | 3 |
| **总计** | **58** | **55** | **0** | **3** |

### 完成进度

```
Week 1-2: [████████████] 100%
Week 3-4: [████████████] 100%
Week 5-6: [██████████] 100%
Week 7-8: [████████░░░░] 85%
总体:     [█████████░░░] 95%
```

---

## 成功指标

| 指标 | 目标 | 当前 | 状态 |
|------|------|------|------|
| AI诊断准确率 | > 80% | — | 待验证 |
| 诊断响应时间（LLM） | < 30秒 | — | 待验证 |
| 诊断响应时间（规则） | < 1秒 | — | 待验证 |
| AI交互查询响应时间 | < 5秒 | — | 待验证 |
| 性能基线覆盖率 | > 90% | — | 待验证 |
| 用户满意度 | > 4.0/5.0 | — | 待验证 |

---

**最后更新**: 2026-05-05
**文件位置**: `TODOS.md`