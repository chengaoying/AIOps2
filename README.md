# AIOps 大数据智能诊断平台

AIOps2 项目 - 基于AI的大数据平台智能运维诊断系统

## 核心价值

- **30秒内诊断**：从告警到根因分析只需30秒
- **统一视图**：一站式查看 YARN/Hive/Spark/Flink 健康状态
- **AI 增强**：知识库 + LLM 混合诊断，90%+ 准确率
- **容器化部署**：docker-compose 一键启动

## 快速开始

```bash
# 启动所有服务
docker-compose up -d

# 访问 Dashboard
open http://localhost:3000
```

## 技术架构

### 混血方案：Go + Java

| 服务 | 语言 | 端口 | 职责 |
|------|------|------|------|
| **collector** | Go | 8081 | 数据采集、WAL 缓冲、批量写入 |
| **diagnosis-api** | Go | 8080 | 知识库检索、LLM 调用、缓存 |
| **knowledge-base-api** | Java | 8082 | 文档采集、LLM 分析提炼 |
| **alert-manager** | Java | 8083 | 告警聚合、渠道分发 |
| **web** | React | 3000 | 前端 Dashboard |

### 服务架构

```
┌─────────────────────────────────────────────────────────────┐
│  Web (Port 3000) - React SPA                                 │
└────────────────────────┬────────────────────────────────────┘
                         │ /api/*
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Go Services                                                 │
│  ├── collector (8081) - 数据采集                            │
│  └── diagnosis-api (8080) - 诊断 API                       │
└──────────┬──────────────────────────────┬───────────────────┘
           │                              │
           ▼                              ▼
┌─────────────────┐          ┌─────────────────────────────────┐
│    StarRocks    │          │  Java Services                    │
│   (元数据存储)   │          │  ├── knowledge-base-api (8082)   │
└─────────────────┘          │  └── alert-manager (8083)        │
                             └─────────────────────────────────┘
           │                              │
           ▼                              ▼
┌─────────────────┐          ┌─────────────────┐
│     Redis       │          │    Milvus       │
│   (缓存)        │          │   (向量存储)     │
└─────────────────┘          └─────────────────┘
                                       │
                                       ▼
                              ┌─────────────────┐
                              │ Elasticsearch   │
                              │   (全文检索)    │
                              └─────────────────┘
```

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | Vite + React + TypeScript | 轻量级生产级前端 |
| Go 后端 | Go + Gin | 高并发 API |
| Java 后端 | Java 17 + Spring Boot 3.2 | 企业级服务 |
| 元数据 | StarRocks 2.5+ | 分析型数据库 |
| 缓存 | Redis 7+ | 热数据缓存 |
| 向量库 | Milvus 3.1 | 知识库向量检索 |
| 全文检索 | Elasticsearch 8.11 | 知识库全文检索 |
| 部署 | Docker Compose | 容器化编排 |

## 项目结构

```
AIOps/
├── go/                           # Go 服务
│   ├── cmd/
│   │   ├── collector/            # Collector Agent (8081)
│   │   └── diagnosis-api/       # Diagnosis API (8080)
│   └── internal/
│       ├── collector/
│       └── diagnosis/
├── java/                          # Java 服务
│   ├── knowledge-base-api/       # Knowledge Base API (8082)
│   │   └── src/main/java/com/AIOps/
│   └── alert-manager/           # Alert Manager (8083)
│       └── src/main/java/com/AIOps/
├── web/                          # React 前端
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── styles/
│   └── Dockerfile
├── docker-compose.yml             # 容器编排
└── docs/                          # 设计文档
    ├── design/
    ├── phase1-plan.md
    └── architecture-review-report.md
```

## API 端点

### Go Services

| 端点 | 方法 | 服务 | 描述 |
|------|------|------|------|
| `/health` | GET | collector | 健康检查 |
| `/api/v1/collect` | POST | collector | 采集作业数据 |
| `/health` | GET | diagnosis-api | 健康检查 |
| `/api/v1/dashboard/home` | GET | diagnosis-api | 首页数据 |
| `/api/v1/diagnosis` | POST | diagnosis-api | 作业诊断 |
| `/api/v1/diagnosis/history` | GET | diagnosis-api | 诊断历史 |
| `/api/v1/knowledge/retrieve` | POST | diagnosis-api | 知识库检索 |
| `/api/v1/assistant/chat` | POST | diagnosis-api | AI 对话 |

### Java Services

| 端点 | 方法 | 服务 | 描述 |
|------|------|------|------|
| `/health` | GET | knowledge-base-api | 健康检查 |
| `/api/v1/knowledge/cards` | GET/POST | knowledge-base-api | 知识卡片管理 |
| `/api/v1/knowledge/collect` | POST | knowledge-base-api | 文档采集 |
| `/health` | GET | alert-manager | 健康检查 |
| `/api/v1/alerts` | GET/POST | alert-manager | 告警管理 |
| `/api/v1/channels` | GET/POST | alert-manager | 通知渠道 |

## 环境变量

### Collector

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8081 | 监听端口 |
| STARROCKS_HOST | starrocks | StarRocks 主机 |
| STARROCKS_PORT | 9030 | StarRocks 端口 |

### Diagnosis API

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 监听端口 |
| STARROCKS_HOST | starrocks | StarRocks 主机 |
| REDIS_HOST | redis | Redis 主机 |

### Knowledge Base API

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8082 | 监听端口 |
| MILVUS_HOST | milvus | Milvus 主机 |
| ES_HOST | elasticsearch | Elasticsearch 主机 |

### Alert Manager

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8083 | 监听端口 |
| STARROCKS_HOST | starrocks | StarRocks 主机 |

## 设计文档

- [产品设计 PRD](docs/design/PRD.md)
- [设计系统](docs/design/DESIGN.md)
- [组件规范](docs/design/09-component-specs.md)
- [Dashboard 设计](docs/design/10-dashboard-home.md)
- [知识库设计](docs/design/17-knowledge-base.md)
- [Phase 1 计划](docs/phase1-plan.md)

## 项目状态

| 阶段 | 状态 | 完成日期 |
|------|------|----------|
| 产品设计 | ✅ 完成 | 2026-05-08 |
| 架构设计 | ✅ 完成 | 2026-05-04 |
| 设计系统 | ✅ 完成 | 2026-05-07 |
| Phase 1 开发 | 🔄 进行中 | - |