import { useState } from 'react'
import { Bell, AlertCircle, CheckCircle, XCircle, X } from 'lucide-react'

interface Alert {
  id: string
  type: 'warning' | 'error' | 'info'
  title: string
  message: string
  platform?: string
  timestamp: string
  read: boolean
}

const mockAlerts: Alert[] = [
  {
    id: '1',
    type: 'error',
    title: 'Spark Executor OOM',
    message: '作业 spark_job_001 Executor 内存溢出',
    platform: 'SPARK',
    timestamp: '10:32',
    read: false,
  },
  {
    id: '2',
    type: 'warning',
    title: 'YARN 队列满',
    message: '生产队列资源使用率超过 90%',
    platform: 'YARN',
    timestamp: '09:45',
    read: false,
  },
  {
    id: '3',
    type: 'info',
    title: 'Flink Checkpoint 超时',
    message: '作业 flink_job_042 Checkpoint 超时',
    platform: 'FLINK',
    timestamp: '08:20',
    read: true,
  },
]

const platformColors: Record<string, string> = {
  YARN: '#1d4ed8',
  HIVE: '#ca8a04',
  SPARK: '#2563eb',
  FLINK: '#7c3aed',
}

export default function AlertSidebar() {
  const [isOpen, setIsOpen] = useState(false)
  const [alerts, setAlerts] = useState(mockAlerts)

  const unreadCount = alerts.filter((a) => !a.read).length

  const handleMarkRead = (id: string) => {
    setAlerts((prev) =>
      prev.map((a) => (a.id === id ? { ...a, read: true } : a))
    )
  }

  const handleDismiss = (id: string) => {
    setAlerts((prev) => prev.filter((a) => a.id !== id))
  }

  const getIcon = (type: string) => {
    switch (type) {
      case 'error':
        return <XCircle size={16} className="text-error" />
      case 'warning':
        return <AlertCircle size={16} className="text-warning" />
      default:
        return <CheckCircle size={16} className="text-info" />
    }
  }

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        style={{
          position: 'fixed',
          right: 'var(--space-md)',
          bottom: 'var(--space-md)',
          width: '48px',
          height: '48px',
          borderRadius: '50%',
          background: 'var(--info)',
          color: 'white',
          border: 'none',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: 'var(--shadow-md)',
          zIndex: 100,
        }}
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <span
            style={{
              position: 'absolute',
              top: '-4px',
              right: '-4px',
              width: '18px',
              height: '18px',
              borderRadius: '50%',
              background: 'var(--error)',
              color: 'white',
              fontSize: '11px',
              fontWeight: 600,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            right: 0,
            bottom: 0,
            width: '360px',
            background: 'var(--bg-primary)',
            boxShadow: '-4px 0 20px rgba(0,0,0,0.15)',
            zIndex: 200,
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div
            style={{
              padding: 'var(--space-md)',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <h2 style={{ fontSize: '16px', fontWeight: 600 }}>
              告警中心 ({unreadCount} 未读)
            </h2>
            <button
              onClick={() => setIsOpen(false)}
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: 'var(--space-xs)',
              }}
            >
              <X size={20} />
            </button>
          </div>

          <div
            style={{
              flex: 1,
              overflowY: 'auto',
              padding: 'var(--space-md)',
              display: 'flex',
              flexDirection: 'column',
              gap: 'var(--space-sm)',
            }}
          >
            {alerts.length === 0 ? (
              <div
                style={{
                  textAlign: 'center',
                  padding: 'var(--space-xl)',
                  color: 'var(--text-secondary)',
                }}
              >
                暂无告警
              </div>
            ) : (
              alerts.map((alert) => (
                <div
                  key={alert.id}
                  style={{
                    padding: 'var(--space-md)',
                    borderRadius: 'var(--radius-md)',
                    border: '1px solid var(--border)',
                    background: alert.read ? 'transparent' : 'var(--bg-secondary)',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 'var(--space-xs)',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'flex-start',
                      gap: 'var(--space-sm)',
                    }}
                  >
                    {getIcon(alert.type)}
                    <div style={{ flex: 1 }}>
                      <div
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 'var(--space-sm)',
                        }}
                      >
                        <span style={{ fontWeight: 500, fontSize: '14px' }}>
                          {alert.title}
                        </span>
                        {alert.platform && (
                          <span
                            style={{
                              padding: '2px 6px',
                              borderRadius: '4px',
                              fontSize: '10px',
                              fontWeight: 600,
                              color: 'white',
                              background: platformColors[alert.platform] || '#666',
                            }}
                          >
                            {alert.platform}
                          </span>
                        )}
                      </div>
                      <p
                        style={{
                          fontSize: '13px',
                          color: 'var(--text-secondary)',
                          margin: '4px 0 0',
                        }}
                      >
                        {alert.message}
                      </p>
                      <span
                        style={{
                          fontSize: '12px',
                          color: 'var(--text-tertiary)',
                        }}
                      >
                        {alert.timestamp}
                      </span>
                    </div>
                    <div style={{ display: 'flex', gap: '4px' }}>
                      {!alert.read && (
                        <button
                          onClick={() => handleMarkRead(alert.id)}
                          style={{
                            padding: '4px',
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            color: 'var(--text-secondary)',
                          }}
                          title="标记已读"
                        >
                          <CheckCircle size={14} />
                        </button>
                      )}
                      <button
                        onClick={() => handleDismiss(alert.id)}
                        style={{
                          padding: '4px',
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                          color: 'var(--text-secondary)',
                        }}
                        title="关闭"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </>
  )
}
