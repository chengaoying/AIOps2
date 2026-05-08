-- AIOps StarRocks Schema
-- Version: 0.0.0.1

CREATE DATABASE IF NOT EXISTS aiops;
USE aiops;

-- 作业元数据表
CREATE TABLE IF NOT EXISTS job_meta (
    job_id            VARCHAR(128)     NOT NULL COMMENT '作业ID',
    platform          VARCHAR(32)     NOT NULL COMMENT '平台类型: YARN/HIVE/SPARK/FLINK',
    job_name          VARCHAR(256)    NOT NULL COMMENT '作业名称',
    status            VARCHAR(32)     NOT NULL COMMENT '状态: RUNNING/SUCCESS/FAILED/KILLED',
    start_time        DATETIME        NOT NULL COMMENT '开始时间',
    end_time          DATETIME        COMMENT '结束时间',
    duration_ms       BIGINT          DEFAULT 0 COMMENT '执行时长(毫秒)',
    submit_time       DATETIME        COMMENT '提交时间',
    priority          VARCHAR(32)     DEFAULT 'NORMAL' COMMENT '优先级',
    user              VARCHAR(64)     NOT NULL COMMENT '提交用户',
    queue             VARCHAR(64)     DEFAULT 'default' COMMENT '队列',
    exit_code         INT             DEFAULT 0 COMMENT '退出码',
    error_msg         TEXT            COMMENT '错误信息',
    logs              TEXT            COMMENT '日志片段',
    metrics           JSON            COMMENT '指标JSON',
    dependency_job_ids VARCHAR(1024)   COMMENT '依赖作业ID列表',
    raw_data          JSON            COMMENT '原始数据',
    create_time       DATETIME        DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time       DATETIME        DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (job_id),
    INDEX idx_platform (platform),
    INDEX idx_status (status),
    INDEX idx_user (user),
    INDEX idx_queue (queue),
    INDEX idx_start_time (start_time),
    INDEX idx_create_time (create_time)
) ENGINE=OLAP
DUPLICATE KEY(job_id, platform)
COMMENT '作业元数据表'
DISTRIBUTED BY HASH(job_id) BUCKETS 10;

-- 诊断事件表
CREATE TABLE IF NOT EXISTS incidents (
    incident_id       VARCHAR(64)      NOT NULL COMMENT '诊断事件ID',
    job_id            VARCHAR(128)     NOT NULL COMMENT '作业ID',
    platform          VARCHAR(32)      NOT NULL COMMENT '平台类型',
    status            VARCHAR(32)      NOT NULL COMMENT '作业状态',
    error_msg         TEXT             COMMENT '错误信息',
    root_cause        TEXT             COMMENT '根因分析',
    confidence        DOUBLE           DEFAULT 0.0 COMMENT '置信度',
    suggestions       JSON             COMMENT '修复建议',
    diagnosis_type    VARCHAR(32)      DEFAULT 'LLM' COMMENT '诊断类型: LLM/RULE/KB',
    create_time       DATETIME         DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    PRIMARY KEY (incident_id),
    INDEX idx_job_id (job_id),
    INDEX idx_platform (platform),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=OLAP
DUPLICATE KEY(incident_id, job_id)
COMMENT '诊断事件表'
DISTRIBUTED BY HASH(incident_id) BUCKETS 10;

-- 作业基线表 (用于性能基准)
CREATE TABLE IF NOT EXISTS job_baseline (
    id                BIGINT          AUTO_INCREMENT COMMENT '自增ID',
    platform          VARCHAR(32)     NOT NULL COMMENT '平台类型',
    job_name          VARCHAR(256)    NOT NULL COMMENT '作业名称(模糊匹配)',
    p50_duration_ms  BIGINT          DEFAULT 0 COMMENT 'P50执行时长',
    p95_duration_ms  BIGINT          DEFAULT 0 COMMENT 'P95执行时长',
    p99_duration_ms  BIGINT          DEFAULT 0 COMMENT 'P99执行时长',
    avg_duration_ms  BIGINT          DEFAULT 0 COMMENT '平均执行时长',
    sample_count      INT             DEFAULT 0 COMMENT '样本数量',
    update_time       DATETIME        DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (id),
    UNIQUE KEY uk_platform_job (platform, job_name),
    INDEX idx_update_time (update_time)
) ENGINE=OLAP
DUPLICATE KEY(id)
COMMENT '作业性能基线表'
DISTRIBUTED BY HASH(id) BUCKETS 10;

-- 告警记录表
CREATE TABLE IF NOT EXISTS alerts (
    alert_id          VARCHAR(64)      NOT NULL COMMENT '告警ID',
    platform          VARCHAR(32)      NOT NULL COMMENT '平台类型',
    alert_type        VARCHAR(64)      NOT NULL COMMENT '告警类型',
    severity          VARCHAR(32)      NOT NULL COMMENT '严重程度: CRITICAL/WARNING/INFO',
    title             VARCHAR(256)     NOT NULL COMMENT '告警标题',
    content           TEXT             COMMENT '告警内容',
    job_id            VARCHAR(128)     COMMENT '关联作业ID',
    status            VARCHAR(32)      DEFAULT 'OPEN' COMMENT '状态: OPEN/ACKNOWLEDGED/RESOLVED',
    channel          VARCHAR(32)      COMMENT '通知渠道: DINGTALK/FEISHU/WECHAT/EMAIL',
    notified_at       DATETIME         COMMENT '通知时间',
    resolved_at       DATETIME         COMMENT '解决时间',
    create_time       DATETIME         DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    PRIMARY KEY (alert_id),
    INDEX idx_platform (platform),
    INDEX idx_alert_type (alert_type),
    INDEX idx_severity (severity),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=OLAP
DUPLICATE KEY(alert_id)
COMMENT '告警记录表'
DISTRIBUTED BY HASH(alert_id) BUCKETS 10;

-- 知识卡片表
CREATE TABLE IF NOT EXISTS knowledge_cards (
    card_id           VARCHAR(64)      NOT NULL COMMENT '知识卡片ID',
    platform          VARCHAR(32)      NOT NULL COMMENT '平台类型',
    error_type        VARCHAR(128)     NOT NULL COMMENT '错误类型',
    error_patterns    JSON             COMMENT '错误模式(多种表述)',
    root_cause        TEXT             NOT NULL COMMENT '根因分析',
    suggestions       JSON             NOT NULL COMMENT '修复建议',
    source            VARCHAR(256)     COMMENT '来源文档',
    confidence        DOUBLE           DEFAULT 0.0 COMMENT '置信度',
    tags              JSON             COMMENT '标签',
    embedding         JSON             COMMENT '向量表示',
    status            VARCHAR(32)      DEFAULT 'ACTIVE' COMMENT '状态: ACTIVE/DEPRECATED',
    create_time       DATETIME         DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time       DATETIME         DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (card_id),
    INDEX idx_platform (platform),
    INDEX idx_error_type (error_type),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=OLAP
DUPLICATE KEY(card_id)
COMMENT '知识卡片表'
DISTRIBUTED BY HASH(card_id) BUCKETS 10;
