# 06 - Chat UI 详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

Chat UI 是用户与 AIOps 交互的主要界面，支持自然语言查询和诊断结果的展示。

### 设计方向（从 design-shotgun 生成 3 个变体）

| 变体 | 风格 | 特点 |
|------|------|------|
| B1 | 专业消息风格 | 清晰的对话气泡，平台标识 |
| B2 | 紧凑Telegram风格 | 时间戳，密集高效 |
| B3 | 分屏布局 | 左侧对话，右侧作业信息 |

---

## B1: 专业消息风格

### 布局结构（推荐用于AI助手）

```
┌─────────────────────────────────────────────────────────────────┐
│ AIOps AI                                      [用户头像]        │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│ 📊 Dashboard│  AI助手                                       │
│             │  ┌───────────────────────────────────────────┐ │
│ 🗄️ 元仓    │  │                                           │ │
│             │  │  ┌─────────────────────────────────────┐│ │
│ 🔍 作业诊断 ▼│  │ │ User: 为什么作业变慢了？            ││ │
│   ├ 作业诊断 │  │ └─────────────────────────────────────┘│ │
│   └ 诊断历史 │  │                                           │ │
│             │  │  ┌─────────────────────────────────────┐│ │
│ 💬 AI助手 ●│  │ │ AI: 分析完成。spark_job_001 执行  ││ │
│             │  │ │ 时间从5分钟增加到15分钟。            ││ │
│ ⚙️ 系统配置▼│  │ │ [Spark ●]                          ││ │
│   ├ 用户管理 │  │ └─────────────────────────────────────┘│ │
│   ├ 集群配置 │  │                                           │ │
│   └ 系统配置 │  │  ┌─────────────────────────────────────┐│ │
│             │  │ │ Executor内存不足导致OOM              ││ │
│  Sidebar    │  │ ├─────────────────────────────────────┤│ │
│             │  │ │ 1. 增加executor内存 4g→6g [低]    ││ │
│             │  │ │ 2. 解决数据倾斜 [中]               ││ │
│             │  │ └─────────────────────────────────────┘│ │
│             │  │                                           │ │
│             │  ├───────────────────────────────────────────┤ │
│             │  │ [问我任何关于作业的问题...        ] [发送]│ │
│             │  └───────────────────────────────────────────┘ │
└─────────────┴─────────────────────────────────────────────────┘
```

### 组件规范

#### 消息气泡

**用户消息**
- 右对齐
- 蓝色背景 (#2563EB)
- 白色文字
- 圆角 (16px，左下角直角)
- 最大宽度 70%

**AI 消息**
- 左对齐
- 白色背景 (#FFFFFF)
- 灰色边框 (#E5E5E5)
- 圆角 (16px，右下角直角)
- 显示平台图标 (● Spark #4ECDC4)
- 最大宽度 80%

#### 诊断卡片

嵌入在 AI 消息中：
- 白色卡片
- 平台颜色左边框 (4px)
- 根因分析 + 修复建议

### 交互行为

| 交互 | 行为 |
|------|------|
| 输入框 | Textarea，Enter 发送，Shift+Enter 换行 |
| 发送按钮 | 蓝色，发送时显示 loading |
| 消息发送 | 滚动到底部，400ms 动画 |
| 建议点击 | 复制到输入框或直接执行 |

---

## B2: 紧凑Telegram风格

### 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ ┌───┐                                                           │
│ │ 👤│ AIOps                                      [用户头像]  │
│ └───┘                                                           │
├─────┬───────────────────────────────────────────────────────────┤
│     │                                                             │
│ [i] │ 14:32 User: 为什么变慢了                          [蓝色]  │
│ [i] │                                                        │
│ [i] │ 14:33 AI: Spark作业执行时间增加150%        [白色]         │
│ [i] │        ● Spark                                          │
│     │                                                        │
│ [i] │ 14:33 AI: 根因: Executor内存不足...          [白色]      │
│     │        [增加内存] [解决倾斜]                              │
│     │                                                        │
│     ├──────────────────────────────────────────────────────────┤
│     │ [输入问题...  ] [➤]                                     │
│     └──────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────────┘
```

### 组件规范

- 侧边栏 48px，只显示图标
- 消息紧凑排列
- 所有消息显示时间戳
- 平台标识用小圆点 ●

### 交互行为

| 交互 | 行为 |
|------|------|
| 发送 | Enter 发送 |
| 建议按钮 | 点击直接显示详情 |

---

## B3: 分屏布局

### 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ ┌─────────┐                                                     │
│ │ Logo    │ AIOps AI                              [用户头像] │
│ └─────────┘                                                     │
├───────────────┬──────────────────────┬────────────────────────────┤
│               │                      │                            │
│  ┌─────────┐ │ User: 为什么变慢了   │  spark_job_001           │
│  │ 诊断    │ │                      │  ┌──────────────────┐    │
│  ├─────────┤ │ AI: 分析完成...      │  │ Platform: Spark  │    │
│  │ 历史    │ │                      │  │ Duration: 15min  │    │
│  ├─────────┤ │ ┌────────────────┐   │  │ Memory: 8GB      │    │
│  │ 作业    │ │ │ 根因分析       │   │  │ Status: FAILED   │    │
│  ├─────────┤ │ │ Executor OOM   │   │  ├──────────────────┤    │
│  │ 集群    │ │ └────────────────┘   │  │ 最近失败: 3次      │    │
│  ├─────────┤ │                      │  │ 相似作业: spark_002│   │
│  │ 告警    │ │ [输入问题...  ]      │  │                    │    │
│  └─────────┘ │                      │  └────────────────────┘    │
│               │                      │                            │
│  Sidebar      │   Chat Area (60%)    │    Info Panel (40%)       │
└───────────────┴──────────────────────┴────────────────────────────┘
```

### 组件规范

- 分屏比例 60:40
- 右侧 Info Panel 显示当前作业详情
- 切换作业时右侧同步更新

### 交互行为

| 交互 | 行为 |
|------|------|
| 点击消息中的作业ID | 右侧 Info Panel 跳转到该作业 |
| 发送消息 | 发送后保持 Info Panel 不变 |

---

## 通用组件

### ChatInput

```tsx
interface ChatInputProps {
    placeholder?: string;
    onSend: (message: string) => void;
    disabled?: boolean;
}
```

### MessageBubble

```tsx
interface MessageBubbleProps {
    type: 'user' | 'ai';
    content: string;
    timestamp?: Date;
    platform?: Platform; // AI 消息显示
    diagnosisCard?: DiagnosisCard; // AI 消息嵌入
}
```

### DiagnosisCard

```tsx
interface DiagnosisCardProps {
    jobId: string;
    platform: Platform;
    confidence: number;
    rootCause: string;
    suggestions: Suggestion[];
    expanded?: boolean;
    onSuggestionClick?: (suggestion: Suggestion) => void;
}
```

### PlatformIcon

```tsx
const platformColors = {
    YARN: '#FF6B6B',
    HIVE: '#FFE66D',
    SPARK: '#4ECDC4',
    FLINK: '#45B7D1',
};

interface PlatformIconProps {
    platform: Platform;
    size?: 'small' | 'medium' | 'large';
}
```

---

## 状态管理

```tsx
interface ChatState {
    messages: Message[];
    currentJobId?: string;
    isLoading: boolean;
    conversationId: string;
}

interface Message {
    id: string;
    type: 'user' | 'ai';
    content: string;
    timestamp: Date;
    platform?: Platform;
    diagnosisCard?: DiagnosisCard;
}
```

---

## API 接口

### 发送消息

```tsx
// POST /api/v1/chat/send
interface ChatSendRequest {
    message: string;
    conversationId: string;
    context?: {
        jobId?: string;
        platform?: Platform;
    };
}

interface ChatSendResponse {
    messageId: string;
    content: string;
    platform?: Platform;
    diagnosisCard?: DiagnosisCard;
    timestamp: Date;
}
```

---

## 设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 推荐方案 | B1 专业消息风格 | 清晰易读，适合工程师 |
| 消息对齐 | 用户右/AI左 | 符合常见聊天习惯 |
| 平台标识 | AI消息内显示 | 上下文清晰 |
| 诊断卡片 | 嵌入AI消息 | 保持对话连贯性 |

---

**关联文档**
- 07-dashboard.md - Dashboard 设计
- 09-component-specs.md - 通用组件规范
