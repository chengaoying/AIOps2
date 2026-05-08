-- AIOps StarRocks Materialized Views
-- Version: 0.0.0.1

USE aiops;

-- 作业失败统计物化视图
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_job_failure_stats
AS SELECT
    platform,
    DATE_FORMAT(start_time, '%Y-%m-%d') as stat_date,
    status,
    COUNT(*) as job_count,
    COUNT(DISTINCT user) as user_count
FROM job_meta
WHERE status IN ('FAILED', 'KILLED')
GROUP BY platform, DATE_FORMAT(start_time, '%Y-%m-%d'), status;

-- Top错误码统计
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_top_error_codes
AS SELECT
    platform,
    exit_code,
    error_msg,
    COUNT(*) as error_count,
    MAX(create_time) as last_occurrence
FROM job_meta
WHERE status = 'FAILED' AND exit_code != 0
GROUP BY platform, exit_code, error_msg;

-- 作业执行时长趋势
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_job_duration_trend
AS SELECT
    platform,
    DATE_FORMAT(start_time, '%Y-%m-%d %H:00:00') as stat_hour,
    COUNT(*) as total_jobs,
    AVG(duration_ms) as avg_duration_ms,
    MAX(duration_ms) as max_duration_ms,
    MIN(duration_ms) as min_duration_ms
FROM job_meta
GROUP BY platform, DATE_FORMAT(start_time, '%Y-%m-%d %H:00:00');

-- 队列资源使用统计
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_queue_usage
AS SELECT
    platform,
    queue,
    DATE_FORMAT(start_time, '%Y-%m-%d') as stat_date,
    COUNT(*) as job_count,
    AVG(duration_ms) as avg_duration_ms,
    COUNT(DISTINCT user) as user_count
FROM job_meta
GROUP BY platform, queue, DATE_FORMAT(start_time, '%Y-%m-%d');

-- 统一作业视图 (关联作业链)
CREATE VIEW IF NOT EXISTS unified_job_view AS
SELECT
    j.job_id,
    j.platform,
    j.job_name,
    j.status,
    j.start_time,
    j.end_time,
    j.duration_ms,
    j.user,
    j.queue,
    j.error_msg,
    i.root_cause,
    i.confidence,
    i.diagnosis_type,
    j.create_time
FROM job_meta j
LEFT JOIN incidents i ON j.job_id = i.job_id;

-- 作业依赖关系
CREATE TABLE IF NOT EXISTS job_dependency (
    job_id            VARCHAR(128)     NOT NULL COMMENT '作业ID',
    dependency_job_id VARCHAR(128)     NOT NULL COMMENT '依赖作业ID',
    dependency_type   VARCHAR(32)      DEFAULT 'DATA' COMMENT '依赖类型: DATA/TIME/RESOURCE',
    create_time       DATETIME        DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    PRIMARY KEY (job_id, dependency_job_id),
    INDEX idx_dependency_job_id (dependency_job_id)
) ENGINE=OLAP
DUPLICATE KEY(job_id)
COMMENT '作业依赖关系表'
DISTRIBUTED BY HASH(job_id) BUCKETS 10;

-- 资源使用趋势视图
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_resource_trend
AS SELECT
    platform,
    queue,
    DATE_FORMAT(start_time, '%Y-%m-%d') as stat_date,
    COUNT(*) as total_jobs,
    SUM(CASE WHEN status = 'RUNNING' THEN 1 ELSE 0 END) as running_jobs,
    SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) as success_jobs,
    SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) as failed_jobs,
    AVG(duration_ms) as avg_duration_ms
FROM job_meta
GROUP BY platform, queue, DATE_FORMAT(start_time, '%Y-%m-%d');
