import { useState, useEffect } from 'react'
import { Search, CheckCircle, XCircle, TrendingUp, TrendingDown } from 'lucide-react'
import clsx from 'clsx'
import { Link } from 'react-router-dom'

// Types
type Platform = 'YARN' | 'HIVE' | 'SPARK' | 'FLINK'
type Status = 'SUCCESS' | 'FAILED' | 'RUNNING'

interface TodayStats {
  totalCount: number
  successCount: number
  failedCount: number
  successRate: number
  trends: {
    totalChange: number
    successChange: number
    failedChange: number
  }
}

interface RecentJob {
  jobId: string
  platform: Platform
  status: Status
  timestamp: string
  rootCause?: string
}

const platformColors: Record<Platform, string> = {
  YARN: 'yarn',
  HIVE: 'hive',
  SPARK: 'spark',
  FLINK: 'flink',
}

const platformNames: Record<Platform, string> = {
  YARN: 'YARN',
  HIVE: 'Hive',
  SPARK: 'Spark',
  FLINK: 'Flink',
}

// Mock data
const mockStats: TodayStats = {
  totalCount: 156,
  successCount: 144,
  failedCount: 12,
  successRate: 92.3,
  trends: {
    totalChange: 12,
    successChange: 8,
    failedChange: -15,
  },
}

const mockRecentJobs: RecentJob[] = [
  { jobId: 'spark_job_001', platform: 'SPARK', status: 'FAILED', timestamp: '10:32', rootCause: 'Executor OOM' },
  { jobId: 'hive_query_042', platform: 'HIVE', status: 'SUCCESS', timestamp: '10:28' },
  { jobId: 'yarn_app_089', platform: 'YARN', status: 'SUCCESS', timestamp: '10:15' },
  { jobId: 'flink_task_023', platform: 'FLINK', status: 'FAILED', timestamp: '10:05', rootCause: 'Checkpoint 超时' },
  { jobId: 'spark_sql_017', platform: 'SPARK', status: 'SUCCESS', timestamp: '09:50' },
]

// Stat Card Component
function StatCard({
  title,
  value,
  trend,
  color,
}: {
  title: string
  value: number | string
  trend?: { direction: 'up' | 'down'; percentage: number }
  color?: string
}) {
  return (
    <div className="card" style={{ padding: '20px' }}>
      <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginBottom: '8px' }}>
        {title}
      </div>
      <div style={{ fontSize: '32px', fontWeight: 700, color: color || 'var(--text-primary)' }}>
        {value}
      </div>
      {trend && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '4px',
            marginTop: '8px',
            fontSize: '12px',
            color: trend.direction === 'up' ? 'var(--success)' : 'var(--danger)',
          }}
        >
          {trend.direction === 'up' ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
          <span>{trend.percentage}%</span>
        </div>
      )}
    </div>
  )
}

// Platform Badge
function PlatformBadge({ platform }: { platform: Platform }) {
  return (
    <span className={clsx('badge', 'badge-platform', platformColors[platform].toLowerCase())}>
      {platformNames[platform]}
    </span>
  )
}

// Status Badge
function StatusBadge({ status }: { status: Status }) {
  const config: Record<Status, { className: string; icon: React.ReactNode; label: string }> = {
    SUCCESS: { className: 'success', icon: <CheckCircle size={12} />, label: '成功' },
    FAILED: { className: 'danger', icon: <XCircle size={12} />, label: '失败' },
    RUNNING: { className: 'warning', icon: <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'currentColor' }} />, label: '运行中' },
  }
  const c = config[status]
  return (
    <span className={clsx('badge', 'badge-status', c.className)} style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
      {c.icon}
      {c.label}
    </span>
  )
}

// Platform Distribution Bar
const platformDist = [
  { platform: 'YARN' as Platform, count: 45, percentage: 28 },
  { platform: 'Spark' as Platform, count: 67, percentage: 42 },
  { platform: 'Hive' as Platform, count: 23, percentage: 14 },
  { platform: 'Flink' as Platform, count: 21, percentage: 13 },
]

// Main Dashboard
export default function Dashboard() {
  const [stats, setStats] = useState<TodayStats | null>(null)
  const [recentJobs, setRecentJobs] = useState<RecentJob[]>([])

  useEffect(() => {
    // Simulate API call
    setTimeout(() => {
      setStats(mockStats)
      setRecentJobs(mockRecentJobs)
    }, 300)
  }, [])

  return (
    <div style={{ padding: 'var(--space-lg)' }}>
      {/* Header */}
      <div style={{ marginBottom: 'var(--space-lg)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: '20px', fontWeight: 600 }}>作业诊断大屏</h1>
          <p style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px' }}>2026-05-08</p>
        </div>
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <select className="form-input" style={{ width: 'auto' }}>
            <option>生产集群</option>
          </select>
          <div style={{ width: 32, height: 32, borderRadius: '50%', background: 'var(--info)', color: 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '12px', fontWeight: 600 }}>
            CY
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 'var(--space-md)', marginBottom: 'var(--space-lg)' }}>
        <StatCard
          title="诊断总数"
          value={stats?.totalCount ?? '-'}
          trend={stats ? { direction: 'up', percentage: stats.trends.totalChange } : undefined}
          color="var(--info)"
        />
        <StatCard
          title="成功数"
          value={stats?.successCount ?? '-'}
          trend={stats ? { direction: 'up', percentage: stats.trends.successChange } : undefined}
          color="var(--success)"
        />
        <StatCard
          title="失败数"
          value={stats?.failedCount ?? '-'}
          trend={stats ? { direction: stats.trends.failedChange >= 0 ? 'up' : 'down', percentage: Math.abs(stats.trends.failedChange) } : undefined}
          color="var(--danger)"
        />
      </div>

      {/* Platform Distribution */}
      <div className="card" style={{ marginBottom: 'var(--space-lg)' }}>
        <div className="card-header">平台分布</div>
        <div className="card-body">
          <div style={{ display: 'flex', gap: 'var(--space-md)', marginBottom: 'var(--space-md)' }}>
            {platformDist.map(({ platform, percentage }) => (
              <div key={platform} style={{ flex: 1 }}>
                <div
                  style={{
                    height: '8px',
                    borderRadius: '4px',
                    background: `var(--${platform.toLowerCase()})`,
                    width: `${percentage}%`,
                    minWidth: '20px',
                  }}
                />
              </div>
            ))}
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: 'var(--text-muted)' }}>
            <span>YARN {platformDist[0].count}</span>
            <span>Spark {platformDist[1].count}</span>
            <span>Hive {platformDist[2].count}</span>
            <span>Flink {platformDist[3].count}</span>
          </div>
        </div>
      </div>

      {/* Recent Diagnoses */}
      <div className="card">
        <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>最近诊断</span>
          <Link to="/diagnosis/history" style={{ fontSize: '12px', fontWeight: 400 }}>查看全部历史 →</Link>
        </div>
        <div>
          {recentJobs.map((job, index) => (
            <div
              key={job.jobId}
              style={{
                display: 'flex',
                alignItems: 'center',
                padding: 'var(--space-md)',
                borderBottom: index < recentJobs.length - 1 ? '1px solid var(--border)' : 'none',
                gap: 'var(--space-md)',
              }}
            >
              <div
                style={{
                  width: '4px',
                  height: '40px',
                  borderRadius: '2px',
                  background: job.status === 'SUCCESS' ? 'var(--success)' : 'var(--danger)',
                }}
              />
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-sm)' }}>
                  <span className="font-mono" style={{ fontSize: '13px' }}>{job.jobId}</span>
                  <PlatformBadge platform={job.platform} />
                  <StatusBadge status={job.status} />
                </div>
                {job.rootCause && (
                  <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px' }}>
                    {job.rootCause}
                  </div>
                )}
              </div>
              <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{job.timestamp}</div>
              <Link to="/diagnosis/job" className="btn btn-secondary" style={{ padding: '6px 12px', fontSize: '12px' }}>
                {job.status === 'FAILED' ? '诊断' : '查看'}
              </Link>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}