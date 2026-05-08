import { useState } from 'react'
import { Search, AlertCircle, CheckCircle, Loader2 } from 'lucide-react'

type Platform = 'YARN' | 'HIVE' | 'SPARK' | 'FLINK'

export default function Diagnosis() {
  const [jobId, setJobId] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [result, setResult] = useState<any>(null)

  const handleDiagnose = async () => {
    if (!jobId.trim()) return
    setIsLoading(true)
    // Simulate API call
    setTimeout(() => {
      setResult({
        jobId: jobId,
        platform: 'SPARK' as Platform,
        status: 'FAILED',
        rootCause: 'Executor 内存溢出，导致 Task 被 Kill',
        confidence: 0.92,
        suggestions: [
          { action: '增加 executor 内存', risk: '低', detail: '将 spark.executor.memory 从 4g 增加到 6g', command: '--conf spark.executor.memory=6g' },
          { action: '优化数据分区', risk: '中', detail: '使用 salting 策略解决数据倾斜问题', command: null },
        ],
      })
      setIsLoading(false)
    }, 1500)
  }

  return (
    <div style={{ padding: 'var(--space-lg)' }}>
      <h1 style={{ fontSize: '20px', fontWeight: 600, marginBottom: 'var(--space-lg)' }}>作业诊断</h1>

      {/* Diagnosis Input */}
      <div className="card" style={{ marginBottom: 'var(--space-lg)' }}>
        <div className="card-body">
          <div style={{ display: 'flex', gap: 'var(--space-md)' }}>
            <input
              type="text"
              className="form-input"
              placeholder="输入作业 ID 或粘贴错误日志..."
              value={jobId}
              onChange={(e) => setJobId(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleDiagnose()}
              style={{ flex: 1 }}
            />
            <button
              className="btn btn-primary"
              onClick={handleDiagnose}
              disabled={isLoading || !jobId.trim()}
              style={{ minWidth: '100px' }}
            >
              {isLoading ? <Loader2 size={16} className="animate-spin" /> : <Search size={16} />}
              {isLoading ? '诊断中...' : '开始诊断'}
            </button>
          </div>
        </div>
      </div>

      {/* Diagnosis Result */}
      {result && (
        <div className="card">
          <div className="card-header" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-sm)' }}>
            <AlertCircle size={16} color="var(--danger)" />
            <span className="font-mono">{result.jobId}</span>
            <span className="badge badge-platform spark">Spark</span>
            <span className="badge badge-status danger">失败</span>
            <span style={{ marginLeft: 'auto', fontSize: '12px', fontWeight: 400 }}>
              置信度: <strong style={{ color: result.confidence >= 0.9 ? 'var(--success)' : result.confidence >= 0.7 ? 'var(--warning)' : 'var(--danger)' }}>
                {(result.confidence * 100).toFixed(0)}%
              </strong>
            </span>
          </div>
          <div className="card-body">
            <div style={{ marginBottom: 'var(--space-lg)' }}>
              <h3 style={{ fontSize: '14px', color: 'var(--text-muted)', marginBottom: 'var(--space-sm)' }}>根因分析</h3>
              <p style={{ fontSize: '14px' }}>{result.rootCause}</p>
            </div>

            <div>
              <h3 style={{ fontSize: '14px', color: 'var(--text-muted)', marginBottom: 'var(--space-sm)' }}>修复建议</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-md)' }}>
                {result.suggestions.map((s: any, i: number) => (
                  <div key={i} style={{ padding: 'var(--space-md)', background: 'var(--bg-secondary)', borderRadius: 'var(--radius-md)', borderLeft: `3px solid ${s.risk === '低' ? 'var(--success)' : s.risk === '中' ? 'var(--warning)' : 'var(--danger)'}` }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-sm)', marginBottom: '4px' }}>
                      <span style={{ fontWeight: 500 }}>{i + 1}. {s.action}</span>
                      <span className={`badge badge-status ${s.risk === '低' ? 'success' : s.risk === '中' ? 'warning' : 'danger'}`}>
                        {s.risk}风险
                      </span>
                    </div>
                    <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>{s.detail}</p>
                    {s.command && (
                      <code style={{ display: 'block', marginTop: '8px', padding: '6px 10px', background: 'var(--bg-primary)', borderRadius: 'var(--radius-sm)', fontSize: '12px', fontFamily: 'var(--font-mono)' }}>
                        {s.command}
                      </code>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Empty State */}
      {!result && !isLoading && (
        <div className="card" style={{ textAlign: 'center', padding: '60px 20px' }}>
          <Search size={48} style={{ color: 'var(--text-muted)', marginBottom: '16px' }} />
          <h3 style={{ color: 'var(--text-secondary)', marginBottom: '8px' }}>输入作业 ID 开始诊断</h3>
          <p style={{ color: 'var(--text-muted)', fontSize: '13px' }}>支持 Spark、Hive、YARN、Flink 作业</p>
        </div>
      )}

      <style>{`
        .animate-spin {
          animation: spin 1s linear infinite;
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  )
}