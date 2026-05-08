import { useState, useEffect, useRef } from 'react'
import { Send, MessageCircle, BarChart2, TrendingUp, AlertTriangle } from 'lucide-react'
import * as echarts from 'echarts'

interface Message {
  id: string
  type: 'user' | 'ai'
  content: string
  timestamp: Date
  intent?: string
  chartType?: string
  chartData?: any
}

interface ChatResponse {
  reply: string
  intent?: string
  chart_type?: string
  chart_data?: any
}

const platformIcons: Record<string, string> = {
  YARN: '🟢',
  HIVE: '🟡',
  SPARK: '🔵',
  FLINK: '🟣',
}

const mockResponses: Record<string, ChatResponse> = {
  oom: {
    reply: 'Spark作业OOM通常有以下原因：\n\n1. **Executor内存不足** - 数据量超过分配的内存\n2. **数据倾斜** - 某个分区数据量过大\n3. **内存泄漏** - 未正确释放资源\n\n建议您：\n- 检查spark.executor.memory配置\n- 优化数据分区策略\n- 使用广播变量减少Shuffle',
    intent: 'FAILURE_ANALYSIS',
    chart_type: 'table',
  },
  slow: {
    reply: '性能分析结果：\n\n该作业执行时间为 **245秒**，相比baseline (180秒) 增长了 **36%**。\n\n可能原因：\n1. 数据量增长导致处理时间增加\n2. 资源竞争导致排队\n3. GC停顿影响',
    intent: 'PERFORMANCE_ANALYSIS',
    chart_type: 'line',
    chart_data: {
      xAxis: ['10:00', '10:15', '10:30', '10:45', '11:00'],
      series: [{ name: '执行时间', data: [180, 195, 210, 230, 245] }],
    },
  },
  resource: {
    reply: '资源使用分析：\n\n近1小时内存使用最高的作业：\n\n| 作业 | 内存使用 | CPU使用 |\n|------|----------|--------|\n| spark_job_001 | 8.5GB | 4核 |\n| hive_query_042 | 6.2GB | 3核 |\n| yarn_app_089 | 4.1GB | 2核 |',
    intent: 'RESOURCE_ANALYSIS',
    chart_type: 'bar',
    chart_data: {
      xAxis: ['spark_job_001', 'hive_query_042', 'yarn_app_089'],
      series: [{ name: '内存(GB)', data: [8.5, 6.2, 4.1] }],
    },
  },
}

export default function Assistant() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      type: 'ai',
      content: '您好，我是 AIOps AI 助手。有什么关于大数据作业的问题我可以帮您解答？',
      timestamp: new Date(),
    },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const chartRefs = useRef<Record<string, echarts.ECharts>>({})

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const getMockResponse = (query: string): ChatResponse => {
    const q = query.toLowerCase()
    if (q.includes('oom') || q.includes('内存')) return mockResponses.oom
    if (q.includes('慢') || q.includes('性能')) return mockResponses.slow
    if (q.includes('资源') || q.includes('内存') || q.includes('cpu')) return mockResponses.resource
    return {
      reply: `关于"${query}"，我可以帮您分析作业性能、资源使用和故障原因。请告诉我更多细节。`,
      intent: 'JOB_QUERY',
      chart_type: 'table',
    }
  }

  const handleSend = async () => {
    if (!input.trim() || loading) return

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: input,
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, userMessage])
    setInput('')
    setLoading(true)

    await new Promise((resolve) => setTimeout(resolve, 800))

    const response = getMockResponse(input)
    const aiMessage: Message = {
      id: (Date.now() + 1).toString(),
      type: 'ai',
      content: response.reply,
      timestamp: new Date(),
      intent: response.intent,
      chartType: response.chart_type,
      chartData: response.chart_data,
    }

    setMessages((prev) => [...prev, aiMessage])
    setLoading(false)
  }

  const renderChart = (chatId: string, chartType: string, chartData: any) => {
    if (!chartData) return null

    const chartRef = (el: HTMLDivElement | null) => {
      if (!el) return
      if (chartRefs.current[chatId]) {
        chartRefs.current[chatId].dispose()
      }

      const chart = echarts.init(el)
      chartRefs.current[chatId] = chart

      let option: echarts.EChartsOption
      switch (chartType) {
        case 'line':
          option = {
            tooltip: { trigger: 'axis' },
            xAxis: { type: 'category', data: chartData.xAxis },
            yAxis: { type: 'value', name: '执行时间(ms)' },
            series: [
              {
                name: chartData.series[0].name,
                type: 'line',
                data: chartData.series[0].data,
                smooth: true,
                areaStyle: { opacity: 0.3 },
              },
            ],
          }
          break
        case 'bar':
          option = {
            tooltip: { trigger: 'axis' },
            xAxis: { type: 'category', data: chartData.xAxis },
            yAxis: { type: 'value', name: '内存(GB)' },
            series: [
              {
                name: chartData.series[0].name,
                type: 'bar',
                data: chartData.series[0].data,
                itemStyle: { color: '#3b82f6' },
              },
            ],
          }
          break
        default:
          return null
      }

      chart.setOption(option)
    }

    return (
      <div
        ref={chartRef}
        style={{
          width: '100%',
          height: '200px',
          marginTop: 'var(--space-sm)',
        }}
      />
    )
  }

  return (
    <div
      style={{
        padding: 'var(--space-lg)',
        height: 'calc(100vh - 40px)',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <h1 style={{ fontSize: '20px', fontWeight: 600, marginBottom: 'var(--space-md)', display: 'flex', alignItems: 'center', gap: 'var(--space-sm)' }}>
        <MessageCircle size={20} />
        AI 助手
      </h1>

      <div
        className="card"
        style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}
      >
        <div
          className="card-body"
          style={{
            flex: 1,
            overflowY: 'auto',
            display: 'flex',
            flexDirection: 'column',
            gap: 'var(--space-md)',
          }}
        >
          {messages.map((msg) => (
            <div
              key={msg.id}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: msg.type === 'user' ? 'flex-end' : 'flex-start',
              }}
            >
              <div
                style={{
                  maxWidth: '70%',
                  padding: '12px 16px',
                  borderRadius:
                    msg.type === 'user' ? '16px 16px 0 16px' : '16px 16px 16px 0',
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

              {msg.type === 'ai' && msg.chartType && msg.chartData && renderChart(msg.id, msg.chartType, msg.chartData)}
            </div>
          ))}

          {loading && (
            <div style={{ display: 'flex', alignItems: 'flex-start' }}>
              <div
                style={{
                  padding: '12px 16px',
                  borderRadius: '16px 16px 16px 0',
                  background: 'var(--bg-primary)',
                  border: '1px solid var(--border)',
                  fontSize: '14px',
                  color: 'var(--text-secondary)',
                }}
              >
                分析中...
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        <div
          style={{
            padding: 'var(--space-md)',
            borderTop: '1px solid var(--border)',
            display: 'flex',
            gap: 'var(--space-sm)',
          }}
        >
          <input
            type="text"
            className="form-input"
            placeholder="输入您的问题..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          />
          <button className="btn btn-primary" onClick={handleSend} disabled={loading}>
            <Send size={16} />
          </button>
        </div>
      </div>
    </div>
  )
}
