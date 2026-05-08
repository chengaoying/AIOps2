# AIOps 项目计划 v2.0

**创建时间**: 2026-05-05
**最后更新**: 2026-05-08
**状态**: 设计阶段完成，准备进入开发

---

## 项目阶段总览

```
已完成 (2026-05-03 ~ 2026-05-08)
├── ✅ 产品设计 (PRD)
├── ✅ 架构设计
├── ✅ 技术架构评审
├── ✅ 设计系统评审
└── ✅ 产品设计评审

进行中 / 待启动
├── 🔄 Phase 1 开发 (6-8周)
└── 📋 Phase 2 规划 (TBD)
```

---

## 已完成阶段

### 产品设计阶段 (2026-05-03 ~ 2026-05-08)

| 里程碑 | 完成日期 | 产出文档 |
|--------|----------|----------|
| 产品愿景 & 核心价值 | 2026-05-04 | PRD.md |
| 用户故事 & 痛点分析 | 2026-05-04 | PRD.md |
| 功能规格 & 优先级 | 2026-05-04 | PRD.md |
| 技术架构设计 | 2026-05-04 | 01-architecture.md |
| 设计系统 | 2026-05-04 | DESIGN.md |
| 组件规范 | 2026-05-06 | 09-component-specs.md |
| Dashboard 设计 | 2026-05-06 | 10-dashboard-home.md |
| 深色模式 & 无障碍 | 2026-05-07 | DESIGN.md (更新) |
| Lucide Icons 迁移 | 2026-05-07 | DESIGN.md (更新) |
| PRD 演示稿 | 2026-05-08 | PRD-slides.html |

### 评审记录

| 评审类型 | 日期 | 状态 |
|----------|------|------|
| 架构评审 (plan-eng-review) | 2026-05-04 | ✅ CLEARED |
| 设计评审 (plan-design-review) | 2026-05-07 | ✅ CLEARED |
| 产品评审 (office-hours) | 2026-05-04 | ✅ DONE |

---

## Phase 1 开发计划 (6-8周)

**预计周期**: 2026-05-09 ~ 2026-06-26

### 关键技术决策

| 决策 | 选择 |
|------|------|
| 消息队列 | 去掉 Kafka，Collector 直接写入 StarRocks |
| 诊断策略 | 知识库+LLM 混合（知识库检索+LLM 推理） |
| LLM 安全 | 三层防护（输入验证+结构化 Prompt+输出验证） |
| 高可用 | Collector/API 无状态微服务 |
| 知识库 | Vector 存储+全文检索（Milvus） |
| 自愈系统 | Phase 2 再做，Phase 1 专注诊断 |
| 向量数据库 | Qdrant（性能优、易部署） |

### Week 1-2: 基础设施 (2026-05-09 ~ 2026-05-22)

#### 1.1 项目脚手架
- [ ] Go 项目初始化，DDD 目录结构
- [ ] 统一采集接口 (Collector interface) 定义
- [ ] 配置管理系统
- [ ] 日志系统

#### 1.2 Collector Agent 框架
- [ ] Plugin Registry 实现
- [ ] WAL 缓冲机制（内存队列 10000 条 + 磁盘 WAL）
- [ ] 批量写入器（每 5 秒或 1000 条）
- [ ] 背压控制（WAL 超过 1GB 时丢弃最旧数据）

#### 1.3 StarRocks 环境
- [ ] 搭建开发环境
- [ ] 创建作业元数据表 (job_meta)
- [ ] 创建诊断事件表 (incidents)
- [ ] 创建物化视图 (mv_job_failure_stats, mv_top_error_codes)

**交付物**:
- Git 仓库初始化
- Collector Agent 框架代码
- StarRocks 表结构 SQL

---

### Week 3-4: 数据采集 + 元仓 (2026-05-23 ~ 2026-06-05)

#### 2.1 数据采集插件

| 插件 | 接入方式 | 采集内容 | 采集频率 |
|------|----------|----------|----------|
| YARN | REST API + ATS | Application 状态、Container 日志、作业耗时 | 5s |
| Hive | HS2 API + Hook | HiveQL 执行日志、查询计划、错误信息 | 事件触发 |
| Spark | History Server + Livy | Stage/SQL 执行、Executor 日志 | 10s |
| Flink | REST API + Metrics | JobManager/TaskManager 状态、Checkpoint | 10s |

#### 2.2 元仓建设
- [ ] 统一元数据视图 (unified_job_view)
- [ ] 作业级关联表 (job_dependency)
- [ ] 资源使用趋势视图 (mv_resource_trend)

**交付物**:
- 4 个采集插件完整实现
- StarRocks 元仓表结构（含视图）

---

### Week 5-6: 诊断引擎 + 知识库 (2026-06-06 ~ 2026-06-19)

#### 3.1 知识库构建
- [ ] Qdrant 部署
- [ ] 知识卡片结构设计
- [ ] 文档采集器 (Doc Collector)
- [ ] LLM 分析提炼器 (LLM Analyzer)
- [ ] 知识检索器 (Retriever)

#### 3.2 诊断 API
- [ ] Context Builder 实现
- [ ] NL Query Parser
- [ ] Prompt 安全防护（三层防护）
- [ ] 通义千问 API 集成
- [ ] 输出解析器

#### 3.3 缓存与降级
- [ ] Redis 缓存（TTL=1 小时）
- [ ] LLM 失败→知识库降级
- [ ] 令牌桶限流

**交付物**:
- 知识库模块
- 诊断 API 代码
- Redis 缓存集成

---

### Week 7-8: UI + 告警 + 测试 (2026-06-20 ~ 2026-07-03)

#### 4.1 Dashboard UI
- [ ] React 组件开发
- [ ] 诊断卡片设计
- [ ] 平台分布可视化
- [ ] 诊断历史列表

#### 4.2 Chat UI
- [ ] 对话界面
- [ ] 对话历史管理
- [ ] 图表展示 (ECharts)

#### 4.3 告警通知
- [ ] 钉钉通知
- [ ] 飞书通知
- [ ] 企业微信通知
- [ ] 邮件通知

#### 4.4 测试
- [ ] 单元测试（200+ 用例，覆盖 80%+）
- [ ] 集成测试（20+ 场景）
- [ ] E2E 测试（5 个核心场景）

**交付物**:
- Dashboard + Chat UI 完整实现
- 告警通知渠道
- 测试报告

---

## Phase 2 规划 (TBD)

| 功能 | 理由 |
|------|------|
| 自愈系统 | Phase 1 聚焦诊断核心价值 |
| 预测性告警 | 基于趋势分析提前预警 |
| 多集群管理 | 支持多环境统一管理 |
| 多租户隔离 | 企业级需求 |
| 语音输入 | 提升操作效率 |

---

## 项目结构

```
AIOps/
├── cmd/
│   ├── collector/          # Collector Agent 入口
│   ├── diagnosis-api/       # Diagnosis API 入口
│   ├── knowledge-base-api/  # Knowledge Base API 入口
│   └── alert-manager/       # Alert Manager 入口
├── internal/
│   ├── collector/
│   │   ├── api/            # HTTP/gRPC API
│   │   ├── service/         # 业务逻辑
│   │   ├── repository/      # 数据访问
│   │   ├── model/           # 领域模型
│   │   └── plugin/          # 插件（YARN/Hive/Spark/Flink）
│   ├── diagnosis/
│   │   ├── api/
│   │   ├── service/
│   │   │   ├── kb/          # 知识库检索
│   │   │   ├── llm/         # LLM 调用
│   │   │   └── context/     # 上下文构建
│   │   ├── repository/
│   │   └── model/
│   ├── knowledge/
│   │   ├── api/
│   │   ├── service/
│   │   ├── collector/       # 文档采集
│   │   ├── analyzer/       # LLM 分析
│   │   └── retriever/      # 检索器
│   ├── alert/
│   │   ├── api/
│   │   ├── service/
│   │   └── notifier/        # 通知渠道
│   └── shared/
│       ├── config/
│       ├── logger/
│       └── errors/
├── web/                     # React 前端
│   ├── src/
│   │   ├── components/     # 组件库
│   │   ├── pages/          # 页面
│   │   ├── hooks/          # 自定义 Hooks
│   │   └── styles/         # 样式
│   └── public/
├── deployments/             # 部署配置
│   ├── helm/
│   └── k8s/
├── docs/                    # 文档
│   ├── design/             # 设计文档
│   ├── phase1-plan.md      # Phase 1 计划
│   └── *.md                # 其他文档
└── tests/                   # 测试
    ├── unit/
    ├── integration/
    └── e2e/
```

---

## 成功指标

| 指标 | 目标 |
|------|------|
| AI 诊断准确率 | > 80% |
| 诊断响应时间（LLM） | < 30s |
| 诊断响应时间（规则） | < 1s |
| AI 交互查询响应时间 | < 5s |
| 性能基线覆盖率 | > 90% |
| 用户满意度 | > 4.0/5.0 |

---

## 外部依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| Go | 1.21+ | 后端语言 |
| React | 18+ | 前端框架 |
| StarRocks | 2.5+ | 元数据仓库 |
| Redis | 7+ | 缓存 |
| Milvus | latest | 向量数据库 |
| Elasticsearch | latest | 全文检索 |
| 通义千问 | latest | LLM |

---

## 文档索引

| 文档 | 路径 |
|------|------|
| 产品设计 | `docs/design/PRD.md` |
| 设计系统 | `docs/design/DESIGN.md` |
| 组件规范 | `docs/design/09-component-specs.md` |
| Dashboard 设计 | `docs/design/10-dashboard-home.md` |
| 技术架构 | `docs/design/01-architecture.md` |
| 架构评审 | `docs/architecture-review-report.md` |
| 设计预览 | `.gstack/projects/AIOps2/designs/design-system-preview-20260507/preview.html` |
| PRD 演示稿 | `docs/design/PRD-slides.html` |

---

**评审状态**

| 评审类型 | 状态 | 日期 |
|----------|------|------|
| CEO Review | ✅ CLEARED | 2026-05-04 |
| Engineering Review | ✅ CLEARED | 2026-05-04 |
| Design Review | ✅ CLEARED | 2026-05-07 |
| Product Review | ✅ DONE | 2026-05-04 |

**下次评审**: Phase 1 中期评审 (Week 4 结束后)