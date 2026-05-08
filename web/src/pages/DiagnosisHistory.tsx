import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, CheckCircle, XCircle } from 'lucide-react'

const mockHistory = [
  { id: 'spark_job_001', platform: 'SPARK', status: 'FAILED', time: '10:32', rootCause: 'Executor OOM' },
  { id: 'hive_query_042', platform: 'HIVE', status: 'SUCCESS', time: '10:28' },
  { id: 'yarn_app_089', platform: 'YARN', status: 'SUCCESS', time: '10:15' },
  { id: 'flink_task_023', platform: 'FLINK', status: 'FAILED', time: '10:05', rootCause: 'Checkpoint 超时' },
  { id: 'spark_sql_017', platform: 'SPARK', status: 'SUCCESS', time: '09:50' },
]

export default function DiagnosisHistory() {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('all')

  return (
    <div style={{ padding: 'var(--space-lg)' }}>
      <h1 style={{ fontSize: '20px', fontWeight: 600, marginBottom: 'var(--space-lg)' }}>诊断历史</h1>

      <div style={{ display: 'flex', gap: 'var(--space-md)', marginBottom: 'var(--space-lg)' }}>
        <input
          type="text"
          className="form-input"
          placeholder="搜索作业 ID..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{ maxWidth: '300px' }}
        />
        <select className="form-input" value={filter} onChange={(e) => setFilter(e.target.value)} style={{ width: 'auto' }}>
          <option value="all">全部</option>
          <option value="success">成功</option>
          <option value="failed">失败</option>
        </select>
      </div>

      <div className="card">
        <table className="table">
          <thead>
            <tr>
              <th>作业 ID</th>
              <th>平台</th>
              <th>状态</th>
              <th>时间</th>
              <th>根因/说明</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {mockHistory.map((item) => (
              <tr key={item.id}>
                <td className="font-mono">{item.id}</td>
                <td><span className={`badge badge-platform ${item.platform.toLowerCase()}`}>{item.platform}</span></td>
                <td>
                  {item.status === 'SUCCESS' ? (
                    <span className="badge badge-status success"><CheckCircle size={12} /> 成功</span>
                  ) : (
                    <span className="badge badge-status danger"><XCircle size={12} /> 失败</span>
                  )}
                </td>
                <td>{item.time}</td>
                <td style={{ color: 'var(--text-muted)', fontSize: '13px' }}>{item.rootCause || '-'}</td>
                <td>
                  <Link to="/diagnosis/job" className="btn btn-secondary" style={{ padding: '4px 8px', fontSize: '12px' }}>
                    {item.status === 'FAILED' ? '重新诊断' : '查看'}
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}