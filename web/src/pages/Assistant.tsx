import { useState } from 'react'
import { Send, MessageCircle } from 'lucide-react'

const mockMessages = [
  { type: 'ai', content: '您好，我是 AIOps AI 助手。有什么关于大数据作业的问题我可以帮您解答？' },
  { type: 'user', content: 'Spark 作业经常 OOM 是怎么回事？' },
  { type: 'ai', content: 'Spark 作业 OOM 通常有以下原因：\n\n1. **Executor 内存不足** - 数据量超过分配的内存\n2. **数据倾斜** - 某个分区数据量过大\n3. **内存泄漏** - 未正确释放资源\n\n建议您：\n- 检查 spark.executor.memory 配置\n- 优化数据分区策略\n- 使用广播变量减少 Shuffle' },
]

export default function Assistant() {
  const [messages, setMessages] = useState(mockMessages)
  const [input, setInput] = useState('')

  const handleSend = () => {
    if (!input.trim()) return
    setMessages([...messages, { type: 'user', content: input }])
    setInput('')
  }

  return (
    <div style={{ padding: 'var(--space-lg)', height: 'calc(100vh - 40px)', display: 'flex', flexDirection: 'column' }}>
      <h1 style={{ fontSize: '20px', fontWeight: 600, marginBottom: 'var(--space-md)' }}>AI 助手</h1>

      <div className="card" style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <div className="card-body" style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 'var(--space-md)' }}>
          {messages.map((msg, i) => (
            <div key={i} style={{ display: 'flex', flexDirection: 'column', alignItems: msg.type === 'user' ? 'flex-end' : 'flex-start' }}>
              <div
                style={{
                  maxWidth: '70%',
                  padding: '12px 16px',
                  borderRadius: msg.type === 'user' ? '16px 16px 0 16px' : '16px 16px 16px 0',
                  background: msg.type === 'user' ? 'var(--info)' : 'var(--bg-primary)',
                  color: msg.type === 'user' ? 'white' : 'var(--text-primary)',
                  border: msg.type === 'ai' ? '1px solid var(--border)' : 'none',
                  whiteSpace: 'pre-wrap',
                  fontSize: '14px',
                  lineHeight: 1.6,
                }}
              >
                {msg.content}
              </div>
            </div>
          ))}
        </div>

        <div style={{ padding: 'var(--space-md)', borderTop: '1px solid var(--border)', display: 'flex', gap: 'var(--space-sm)' }}>
          <input
            type="text"
            className="form-input"
            placeholder="输入您的问题..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          />
          <button className="btn btn-primary" onClick={handleSend}>
            <Send size={16} />
          </button>
        </div>
      </div>
    </div>
  )
}