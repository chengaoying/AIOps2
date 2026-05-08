# AIOps2 系统测试报告

**项目**: AIOps2 大数据智能诊断平台
**日期**: 2026-05-08
**分支**: master
**状态**: Phase 1 完成待验证

---

## 1. 测试范围

本报告覆盖以下测试类型：

| 测试类型 | 状态 | 说明 |
|---------|------|------|
| 单元测试 | 部分完成 | 65+ 测试用例 |
| 集成测试 | 待实施 | Collector→StarRocks, Diagnosis API→LLM |
| E2E测试 | 待实施 | 故障诊断全流程 |
| UI测试 | N/A | 后端项目，无 Web UI |

**注意**: 本项目为后端服务，无运行中的 Web UI，因此 `/qa` 浏览器测试不适用，改用 API 级测试和代码审查。

---

## 2. 单元测试现状

### 2.1 已实现的测试

| 模块 | 测试文件 | 用例数 | 覆盖内容 |
|------|---------|--------|----------|
| Collector Registry | `registry_test.go` | 13 | Register/Get/List/CollectAll/并发 |
| MemoryQueue | `queue_test.go` | 17 | Push/Pop/FIFO/Overflow/并发 |
| Rule Engine | `rule_test.go` | 20 | 平台匹配/Pattern匹配/Confidence |
| LLM Reasoner | `reasoner_test.go` | 15 | Input验证/Output验证/Prompt构建 |

**总计**: 65 个测试用例

### 2.2 测试覆盖明细

#### Collector Registry Tests (`registry_test.go`)

| 测试用例 | 描述 |
|---------|------|
| TestRegistry_Register | 正常注册 |
| TestRegistry_Get | 获取已注册插件 |
| TestRegistry_List | 列出所有插件 |
| TestRegistry_CollectAll | 收集所有插件数据 |
| TestRegistry_InitAll | 初始化所有启用的插件 |
| TestRegistry_InitAll_Errors | 初始化失败处理 |
| TestRegistry_CollectAll_MultipleJobs | 多作业收集 |
| TestRegistry_Unregister | 注销插件 |
| TestRegistry_Empty | 空注册表 |
| TestRegistry_NilPlugin | 空插件检测 |
| TestRegistry_DoubleRegister | 重复注册检测 |

#### MemoryQueue Tests (`queue_test.go`)

| 测试用例 | 描述 |
|---------|------|
| TestMemoryQueue_New | 新队列初始化 |
| TestMemoryQueue_Push | 入队操作 |
| TestMemoryQueue_Pop | 出队操作 |
| TestMemoryQueue_FIFO | FIFO 顺序验证 |
| TestMemoryQueue_Overflow | 溢出驱逐 |
| TestMemoryQueue_PopEmpty | 空队列 Pop |
| TestMemoryQueue_Full | 队列满检测 |
| TestMemoryQueue_Empty | 队列空检测 |
| TestMemoryQueue_Concurrent | 并发 Push |
| TestMemoryQueue_ConcurrentPop | 并发 Pop |
| TestMemoryQueue_Peek | 查看队首 |
| TestMemoryQueue_Clear | 清空队列 |
| TestMemoryQueue_BatchPush | 批量入队 |
| TestMemoryQueue_BatchPop | 批量出队 |
| TestMemoryQueue_BatchPopMoreThanAvailable | 批量溢出 |
| TestMemoryQueue_Stats | 统计数据 |
| TestMemoryQueue_EvictionCount | 驱逐计数 |
| TestMemoryQueue_PushWithTimestamp | 时间戳 |
| TestMemoryQueue_RingBufferWrap | 环形缓冲环绕 |

#### Rule Engine Tests (`rule_test.go`)

| 测试用例 | 描述 |
|---------|------|
| TestRuleEngine_Match_YARN_OOM | YARN OOM 检测 |
| TestRuleEngine_Match_Spark_OOM | Spark OOM 检测 |
| TestRuleEngine_Match_Hive_Memory | Hive 内存检测 |
| TestRuleEngine_Match_Flink_Checkpoint | Flink Checkpoint 超时 |
| TestRuleEngine_Match_NoMatch | 无匹配 |
| TestRuleEngine_Match_PlatformMismatch | 平台不匹配 |
| TestRuleEngine_Match_MultiplePatterns | 多 Pattern 匹配 |
| TestRuleEngine_Match_CaseInsensitive | 大小写不敏感 |
| TestRuleEngine_Match_Confidence | Confidence 评分 |
| TestRuleEngine_Match_WithKB | Knowledge Base 联动 |
| TestRuleEngine_Match_EmptyRules | 空规则集 |
| TestRuleEngine_Match_EmptyErrorMsg | 空错误信息 |
| TestRuleEngine_Match_AllPlatforms | 全平台覆盖 |
| TestRuleEngine_AddRule | 动态添加规则 |
| TestRuleEngine_RuleCount | 规则计数 |
| TestRuleEngine_GetRulesByPlatform | 按平台获取规则 |

#### LLM Reasoner Tests (`reasoner_test.go`)

| 测试用例 | 描述 |
|---------|------|
| TestInputValidator_Validate | 输入验证 |
| TestInputValidator_CheckSQLInjection | SQL 注入检测 |
| TestInputValidator_CheckXSS | XSS 检测 |
| TestInputValidator_CheckLength | 长度限制 |
| TestOutputValidator_Validate | 输出验证 |
| TestOutputValidator_CheckJSON | JSON 解析 |
| TestOutputValidator_CheckStructure | 结构检查 |
| TestOutputValidator_RemoveSensitive | 敏感信息移除 |
| TestPromptBuilder_Build | Prompt 构建 |
| TestPromptBuilder_Escape | 转义处理 |
| TestPromptBuilder_BuildWithContext | 带上下文构建 |
| TestPromptBuilder_EmptyJob | 空 Job 处理 |
| TestInputValidator_MultipleValidation | 多次验证 |
| TestOutputValidator_EmptyOutput | 空输出检测 |

---

## 3. 集成测试设计

### 3.1 Collector → StarRocks 集成测试

**测试目标**: 验证数据从 Collector 经 WAL 缓冲写入 StarRocks

```
测试场景:
1. YARN Plugin 采集作业数据
2. 数据进入 MemoryQueue (容量 10000)
3. BatchWriter 每5秒或1000条批量写入
4. StarRocks 存储并可查询
```

**测试用例**:

| ID | 测试场景 | 预期结果 |
|----|----------|----------|
| INT-001 | 正常采集写入 | 数据正确写入 job_meta 表 |
| INT-002 | WAL 缓冲Overflow | 丢弃最旧数据，新数据入队 |
| INT-003 | StarRocks 不可用 | 数据写入 WAL 文件，恢复后重放 |
| INT-004 | 批量写入500条 | 每5秒批量写入成功 |
| INT-005 | 多插件同时采集 | 4个插件数据正确合并写入 |

**SQL 验证**:

```sql
-- 验证数据写入
SELECT job_id, platform, status, COUNT(*) FROM job_meta
WHERE start_time >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
GROUP BY job_id, platform, status;

-- 验证批量写入计数
SELECT COUNT(*) FROM job_meta WHERE create_time >= DATE_SUB(NOW(), INTERVAL 1 MINUTE);
```

### 3.2 Diagnosis API → LLM 集成测试

**测试目标**: 验证诊断引擎完整调用链路

```
测试场景:
1. 接收诊断请求 (job_id, platform, error_msg)
2. 检查 Redis 缓存
3. 限流检查 (10 QPS)
4. 构建上下文 (Context Builder)
5. 检索知识库 (Hybrid KB)
6. 调用 LLM (通义千问)
7. 结果缓存或降级
```

**测试用例**:

| ID | 测试场景 | 预期结果 |
|----|----------|----------|
| INT-011 | 正常诊断流程 | 返回根因、置信度、建议 |
| INT-012 | 缓存命中 | 直接返回缓存结果 |
| INT-013 | 限流触发 | 返回 429 Too Many Requests |
| INT-014 | LLM 调用失败 | 降级到规则引擎 |
| INT-015 | LLM 超时 (>30s) | 降级到规则引擎 |
| INT-016 | 空 error_msg | 返回参数错误 |
| INT-017 | 未知平台 | 使用默认规则 |

**API 测试**:

```bash
# 健康检查
curl -s http://localhost:8080/health | jq .

# 诊断请求
curl -s -X POST http://localhost:8080/api/v1/diagnosis \
  -H "Content-Type: application/json" \
  -d '{"job_id":"spark_001","platform":"SPARK","error_msg":"Executor OOM"}' | jq .

# 知识库检索
curl -s -X POST http://localhost:8080/api/v1/knowledge/retrieve \
  -H "Content-Type: application/json" \
  -d '{"platform":"SPARK","error_msg":"OutOfMemoryError","top_k":5}' | jq .
```

### 3.3 告警通知集成测试

**测试用例**:

| ID | 测试场景 | 预期结果 |
|----|----------|----------|
| INT-021 | 钉钉 Webhook | 消息发送成功 |
| INT-022 | 飞书 Webhook | 消息发送成功 |
| INT-023 | 企业微信 Webhook | 消息发送成功 |
| INT-024 | 邮件发送 | SMTP 发送成功 |
| INT-025 | 通知渠道故障 | 降级到备用渠道 |

---

## 4. E2E 测试场景

### 4.1 Spark OOM 故障诊断全流程

```
用户流程:
1. 用户提交 Spark 作业
2. 作业执行中 Executor OOM
3. Collector 采集错误日志
4. 写入 StarRocks job_meta 表
5. 用户调用诊断 API
6. 诊断引擎:
   a. 检索知识库 → 匹配 "Executor OOM" 规则
   b. 查询相似案例 → 找到历史案例
   c. 调用 LLM → 生成诊断报告
7. 返回根因: "Executor 内存不足，建议增加 executor.memory"
```

**验证点**:

| 步骤 | 验证内容 |
|------|----------|
| 1 | Spark 作业提交到集群 |
| 2 | YARN ResourceManager 或 Spark History Server 获取作业状态 |
| 3 | Collector 日志包含 "OutOfMemoryError" |
| 4 | job_meta 表有对应记录，status='FAILED' |
| 5 | API 返回诊断结果 |
| 6a | 知识库检索返回相关卡片 |
| 6b | 相似案例查询返回历史诊断记录 |
| 6c | LLM 返回结构化诊断 |
| 7 | 响应包含 root_cause、confidence、suggestions |

### 4.2 Hive 语义错误诊断全流程

```
验证点:
1. Hive 作业提交失败
2. 错误信息: "SemanticException"
3. 诊断引擎匹配 Hive 规则
4. 返回建议: "检查 JOIN 条件或数据类型"
```

### 4.3 LLM 降级到规则引擎

```
验证点:
1. 模拟 LLM API 不可用 (设置错误的 API Key)
2. 发送诊断请求
3. 验证降级到规则引擎
4. 验证返回结果包含 Fallback=true
```

### 4.4 StarRocks 不可用时 WAL 缓冲

```
验证点:
1. 停止 StarRocks 容器
2. Collector 继续采集数据
3. 验证数据写入 WAL 文件
4. 重启 StarRocks
5. 验证数据从 WAL 恢复写入
```

---

## 5. 代码质量分析

### 5.1 关键路径

| 模块 | 文件 | 关键逻辑 |
|------|------|----------|
| Collector | `registry/registry.go` | 插件注册、CollectAll |
| Collector | `queue/queue.go` | FIFO、Overflow 驱逐 |
| Collector | `writer/writer.go` | 批量写入、ON DUPLICATE KEY |
| Diagnosis API | `engine/engine.go` | 诊断引擎主流程 |
| Diagnosis API | `kb/knowledge_base.go` | RRF 融合检索 |
| Diagnosis API | `llm/reasoner.go` | 三层防护 |
| Diagnosis API | `limiter/limiter.go` | 令牌桶限流 |

### 5.2 风险点

| 风险 | 描述 | 缓解措施 |
|------|------|----------|
| WAL 文件损坏 | 1GB 写入后 CRC32 校验 | 定期重建 WAL |
| LLM API 限流 | 通义千问 QPS 限制 | 令牌桶降级 |
| StarRocks 连接池 | 高并发时连接不足 | 连接池调优 |
| 内存队列满 | 10000 条限制 | WAL 溢出处理 |

---

## 6. 测试环境

### 6.1 Docker Compose 环境

```bash
# 启动所有服务
docker-compose up -d

# 验证服务状态
docker-compose ps

# 查看日志
docker-compose logs -f diagnosis-api
docker-compose logs -f collector
```

### 6.2 服务端口映射

| 服务 | 端口 | 用途 |
|------|------|------|
| diagnosis-api | 8080 | 诊断 API |
| collector | 8081 | 采集 Agent |
| knowledge-base-api | 8082 | 知识库 API |
| alert-manager | 8083 | 告警管理 |
| starrocks | 9030 | StarRocks MySQL |
| redis | 6379 | Redis 缓存 |
| milvus | 19530 | 向量数据库 |
| elasticsearch | 9200 | 全文搜索 |

---

## 7. 下一步建议

### 7.1 短期 (1-2周)

1. **实施集成测试**: 在测试环境部署并运行上述 INT 系列用例
2. **补充缺失测试**: 为 writer、cache、context 模块添加单元测试
3. **CI 集成**: 在 GitHub Actions 中运行测试

### 7.2 中期 (1个月)

1. **E2E 测试自动化**: 使用 Docker Compose + Testcontainers
2. **性能测试**: 1000 QPS 压测、内存 profiling
3. **监控告警**: 添加测试覆盖率监控

### 7.3 长期

1. **混沌工程**: 模拟 StarRocks 宕机、网络分区
2. **A/B 测试**: LLM vs 规则引擎准确率对比
3. **用户验收测试**: 真实用户反馈驱动的测试

---

## 8. 附录

### 8.1 相关文档

- `TODOS.md` - 项目进度跟踪
- `docs/design/04-diagnosis-engine.md` - 诊断引擎设计
- `docs/design/05-nl-query-parser.md` - NL 查询解析器设计
- `sql/starrocks/01_init.sql` - StarRocks 表结构

### 8.2 测试命令

```bash
# 运行所有单元测试
cd go/cmd/collector && go test ./... -v
cd go/cmd/diagnosis-api && go test ./... -v

# 运行指定模块测试
go test -v ./internal/registry/...
go test -v ./internal/queue/...
go test -v ./internal/engine/...
go test -v ./internal/llm/...

# 查看测试覆盖率
go test -cover ./...
```

---

**报告生成时间**: 2026-05-08
**报告版本**: v1.0
**下次更新**: 集成测试实施后
