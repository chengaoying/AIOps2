# 09 - 通用组件规范

**创建时间**: 2026-05-06
**引用**: `../DESIGN.md`

---

## 概述

本组件规范基于 `DESIGN.md`，定义 Phase 1 所有 UI 组件的详细规范。

---

## 设计系统

### 色彩系统

#### 平台色彩

| 平台 | 主色 | 用途 |
|------|------|------|
| YARN | #FF6B6B | YARN 相关标识 |
| Hive | #FFE66D | Hive 相关标识 |
| Spark | #4ECDC4 | Spark 相关标识 |
| Flink | #45B7D1 | Flink 相关标识 |

#### 状态色彩

| 状态 | 颜色 | 用途 |
|------|------|------|
| Success | #10B981 | 成功状态 |
| Warning | #F59E0B | 警告状态 |
| Error | #EF4444 | 错误状态 |
| Info | #3B82F6 | 信息状态 |

#### 置信度色彩

| 置信度 | 颜色 |
|--------|------|
| >= 90% | #10B981 (绿色) |
| >= 70% | #F59E0B (橙色) |
| < 70% | #EF4444 (红色) |

#### 风险标签色彩

| 风险 | 颜色 |
|------|------|
| 低风险 | #10B981 (绿色) |
| 中风险 | #F59E0B (橙色) |
| 高风险 | #EF4444 (红色) |

#### 背景与边框

| 用途 | 颜色 |
|------|------|
| 页面背景 | #F5F5F5 |
| 卡片背景 | #FFFFFF |
| 边框 | #E5E5E5 |
| 侧边栏背景 | #FFFFFF |
| 输入框背景 | #FFFFFF |

---

## 基础组件

### Button

```tsx
interface ButtonProps {
    variant: 'primary' | 'secondary' | 'ghost';
    size: 'small' | 'medium' | 'large';
    disabled?: boolean;
    loading?: boolean;
    onClick?: () => void;
    children: React.ReactNode;
}
```

**样式规范**

| 变体 | 背景 | 文字 | 边框 |
|------|------|------|------|
| primary | #2563EB | #FFFFFF | - |
| secondary | #FFFFFF | #374151 | 1px #E5E5E5 |
| ghost | transparent | #374151 | - |

**尺寸规范**

| 尺寸 | 高度 | 内边距 | 字号 |
|------|------|--------|------|
| small | 32px | 8px 12px | 12px |
| medium | 40px | 12px 16px | 14px |
| large | 48px | 16px 24px | 16px |

### Input

```tsx
interface InputProps {
    type: 'text' | 'number' | 'search';
    placeholder?: string;
    value?: string;
    onChange?: (value: string) => void;
    disabled?: boolean;
    error?: string;
}
```

**样式规范**

| 状态 | 边框 | 背景 |
|------|------|------|
| default | #E5E5E5 | #FFFFFF |
| hover | #D1D5DB | #FFFFFF |
| focus | #2563EB | #FFFFFF |
| error | #EF4444 | #FEF2F2 |
| disabled | #E5E5E5 | #F9FAFB |

### Badge

```tsx
interface BadgeProps {
    variant: 'platform' | 'status' | 'confidence' | 'risk';
    value: string;
    platform?: Platform; // platform 类型时需要
    color?: string; // 其他情况自定义颜色
}
```

---

## 平台组件

### PlatformIcon

```tsx
const platformConfig = {
    YARN: { color: '#FF6B6B', label: 'YARN', icon: '🏃' },
    HIVE: { color: '#FFE66D', label: 'Hive', icon: '🐝' },
    SPARK: { color: '#4ECDC4', label: 'Spark', icon: '⚡' },
    FLINK: { color: '#45B7D1', label: 'Flink', icon: '🌊' },
};

interface PlatformIconProps {
    platform: Platform;
    size?: 'small' | 'medium' | 'large';
    showLabel?: boolean;
}
```

**尺寸**

| 尺寸 | 图标大小 | 标签字号 |
|------|----------|----------|
| small | 12px | 10px |
| medium | 16px | 12px |
| large | 20px | 14px |

### PlatformBadge

```tsx
interface PlatformBadgeProps {
    platform: Platform;
    size?: 'small' | 'medium';
}
```

**样式**: 圆角胶囊，平台颜色背景，白色文字

### PlatformSelector

```tsx
interface PlatformSelectorProps {
    value: Platform | null;
    onChange: (platform: Platform) => void;
    options?: Platform[];
}
```

---

## 诊断组件

### DiagnosisCard

```tsx
interface DiagnosisCardProps {
    jobId: string;
    platform: Platform;
    confidence: number;
    rootCause: string;
    suggestions: Suggestion[];
    onSuggestionClick?: (suggestion: Suggestion) => void;
}
```

**布局**

```
┌─────────────────────────────────────┐
│ ▌ {jobId}  [{Platform}]  {置信度} │  ← 标题栏 (平台色左边框)
├─────────────────────────────────────┤
│ 🔍 根因分析                          │
│ {rootCause}                         │
├─────────────────────────────────────┤
│ 📋 修复建议                          │
│ 1. {suggestion[0]}                  │
│ 2. {suggestion[1]}                  │
└─────────────────────────────────────┘
```

### ConfidenceBadge

```tsx
interface ConfidenceBadgeProps {
    value: number; // 0-1
}
```

**样式**: 根据置信度值显示不同颜色

### SuggestionItem

```tsx
interface SuggestionItemProps {
    action: string;
    risk: 'low' | 'medium' | 'high';
    onClick?: () => void;
}
```

---

## 消息组件

### MessageBubble

```tsx
interface MessageBubbleProps {
    type: 'user' | 'ai';
    content: string;
    timestamp?: Date;
    platform?: Platform; // AI 消息时显示
}
```

**样式**

| 类型 | 对齐 | 背景 | 边框 | 圆角 |
|------|------|------|------|------|
| user | 右 | #2563EB | - | 16px，左下直角 |
| ai | 左 | #FFFFFF | 1px #E5E5E5 | 16px，右下直角 |

### ChatInput

```tsx
interface ChatInputProps {
    placeholder?: string;
    onSend: (message: string) => void;
    disabled?: boolean;
}
```

**样式**: 底部固定，高度自适应，Enter 发送

---

## 导航组件

### Sidebar

```tsx
interface SidebarProps {
    currentPath: string;
    alertCount?: number;
}
```

**新版权导航项**

| 一级路径 | 图标 | 标签 | 二级路径 | 二级标签 |
|----------|------|------|----------|----------|
| /dashboard | 📊 | Dashboard | - | - |
| /metastore | 🗄️ | 元仓 | - | - |
| /diagnosis | 🔍 | 作业诊断 | /diagnosis/job | 作业诊断 |
| | | | /diagnosis/history | 诊断历史 |
| /assistant | 💬 | AI助手 | - | - |
| /settings | ⚙️ | 系统配置 | /settings/users | 用户管理 |
| | | | /settings/clusters | 集群配置 |
| | | | /settings/system | 系统配置 |

**样式规范**

```css
.sidebar {
    width: 200px;
    background: #FFFFFF;
    border-right: 1px solid #E5E5E5;
}

.nav-item {
    padding: 12px 16px;
    cursor: pointer;
    transition: background 0.15s;
}

.nav-item:hover {
    background: #F5F5F5;
}

.nav-item.active {
    background: #EFF6FF;
    color: #2563EB;
}

.nav-item.has-submenu::after {
    content: '▶';
    float: right;
    font-size: 10px;
    transition: transform 0.2s;
}

.nav-item.submenu-expanded::after {
    transform: rotate(90deg);
}

.nav-subitem {
    padding: 8px 24px;
    font-size: 13px;
    color: #666666;
    cursor: pointer;
}

.nav-subitem:hover {
    background: #F5F5F5;
}

.nav-subitem.active {
    color: #2563EB;
    background: #EFF6FF;
}
```

### AlertBadge

```tsx
interface AlertBadgeProps {
    count: number;
    severity?: 'info' | 'warning' | 'critical';
}
```

**样式**: 圆角胶囊，severity对应颜色，显示在父级导航旁

---

## 表格组件

### DataTable

```tsx
interface DataTableProps {
    columns: Column[];
    data: any[];
    loading?: boolean;
    emptyText?: string;
}

interface Column {
    key: string;
    title: string;
    render?: (value: any, row: any) => React.ReactNode;
    width?: string;
}
```

---

## 图表组件

### ChartContainer

```tsx
interface ChartContainerProps {
    title: string;
    type: 'line' | 'bar' | 'pie' | 'table';
    data: any;
}
```

**ECharts 配置**

```tsx
const chartColors = {
    YARN: '#FF6B6B',
    HIVE: '#FFE66D',
    SPARK: '#4ECDC4',
    FLINK: '#45B7D1',
    primary: '#2563EB',
    success: '#10B981',
    warning: '#F59E0B',
    error: '#EF4444',
};
```

---

## 表单组件

### Form

```tsx
interface FormProps {
    onSubmit: (data: any) => void;
    children: React.ReactNode;
}
```

### FormItem

```tsx
interface FormItemProps {
    label: string;
    required?: boolean;
    error?: string;
    children: React.ReactNode;
}
```

---

## 反馈组件

### Loading

```tsx
interface LoadingProps {
    size?: 'small' | 'medium' | 'large';
    text?: string;
}
```

### Toast

```tsx
interface ToastProps {
    type: 'success' | 'error' | 'warning' | 'info';
    message: string;
    duration?: number; // 默认 3000ms
}
```

### Modal

```tsx
interface ModalProps {
    open: boolean;
    onClose: () => void;
    title?: string;
    children: React.ReactNode;
}
```

---

## 页面状态

### 状态分类

| 状态 | 触发场景 | 设计要求 |
|------|----------|----------|
| Loading | 数据加载中 | 骨架屏 + 状态文字 |
| Empty | 无数据 | 友好插图 + 主操作入口 |
| Error | 网络/服务端异常 | 错误原因 + 重试按钮 |
| Success | 操作/保存成功 | Toast 提示，1.5s 自动消失 |

### 骨架屏 (Skeleton)

```tsx
interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  variant: 'text' | 'circular' | 'rectangular';
  animation?: 'pulse' | 'wave' | 'none';
}
```

**样式规范**

| 元素 | 背景色 | 动画 |
|------|--------|------|
| 骨架块 | #E5E5E5 | wave 从左到右 1.5s |
| 文本行 | #E5E5E5，2px 高度 | wave |
| 头像/图标 | #E5E5E5，圆形/方形 | pulse |

### 空状态 (Empty State)

**设计原则**
- 居中布局，视觉权重均匀
- 包含：图标 + 主标题 + 描述文案 + 主操作按钮
- 不同场景使用不同图标和文案

**文案规范**

| 场景 | 图标 | 主标题 | 描述 | 主操作 |
|------|------|--------|------|--------|
| 无诊断记录 | 📋 | 暂无诊断记录 | 开始第一次诊断吧 | 去诊断 → |
| 无作业数据 | 🗄️ | 暂无作业数据 | 请先配置集群连接 | 配置集群 → |
| 无用户 | 👥 | 暂无用户 | 添加第一个用户开始使用 | 添加用户 → |
| 无搜索结果 | 🔍 | 未找到匹配结果 | 尝试调整筛选条件 | 清除筛选 |
| AI助手无消息 | 💬 | 开始对话 | 输入您想了解的作业问题 | — |

### 错误状态 (Error State)

**布局**: 居中显示错误信息，背景色 #FEF2F2（浅红）

```tsx
interface ErrorStateProps {
  title: string;       // 错误标题，如"数据加载失败"
  message: string;    // 详细错误信息
  onRetry?: () => void; // 重试回调
}
```

**文案规范**

| 场景 | 标题 | 详情 | 操作 |
|------|------|------|------|
| 网络错误 | 网络连接异常 | 请检查网络后重试 | 重试 |
| 服务端错误 | 服务暂不可用 | 稍后再试或联系管理员 | 重试 |
| 权限不足 | 无访问权限 | 您没有权限查看此内容 | 联系管理员 |

### 加载状态 (Loading State)

**表格加载**: 使用骨架行，模拟真实数据行高，每页 5-10 行占位

**卡片加载**: 统计卡片使用数字骨架，列表使用行骨架

```tsx
// 统计卡片骨架
<CardSkeleton>
  <Skeleton variant="text" width="60%" />
  <Skeleton variant="text" width="40%" height={32} />
  <Skeleton variant="text" width="80%" />
</CardSkeleton>

// 列表行骨架
<ListItemSkeleton>
  <Skeleton variant="circular" width={40} height={40} />
  <div style={{ flex: 1 }}>
    <Skeleton variant="text" width="30%" />
    <Skeleton variant="text" width="50%" />
  </div>
</ListItemSkeleton>
```

### Toast 提示

```tsx
interface ToastProps {
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
  duration?: number; // 默认 3000ms
  action?: { label: string; onClick: () => void };
}
```

**样式规范**

| 类型 | 背景色 | 图标 |
|------|--------|------|
| success | #DCFCE7 | ✓ |
| error | #FEE2E2 | ✗ |
| warning | #FEF3C7 | ⚠ |
| info | #DBEAFE | ℹ |

---

## 字体规范

| 用途 | 字体 | 字号 | 字重 |
|------|------|------|------|
| 页面标题 | Plus Jakarta Sans | 20px | 600 |
| 卡片标题 | Plus Jakarta Sans | 16px | 600 |
| 正文 | Noto Sans SC | 14px | 400 |
| 标签 | Noto Sans SC | 12px | 500 |
| 代码 | JetBrains Mono | 13px | 400 |

**字体来源**: Google Fonts (Plus Jakarta Sans, Noto Sans SC)

---

## 图标规范

### 图标库
使用 Lucide Icons (https://lucide.dev/)

### 样式规范
- stroke-width: 2
- 尺寸: 16px (默认), 20px (大), 24px (特大)
- 颜色: 继承 currentColor

### 导航图标映射

| 导航项 | 图标名 | 说明 |
|--------|--------|------|
| Dashboard | layout-dashboard | 仪表盘 |
| 元仓 | database | 数据库/元数据 |
| 作业诊断 | search | 搜索 |
| AI助手 | message-circle | 对话 |
| 系统配置 | settings | 设置 |
| 用户管理 | users | 用户组 |
| 集群配置 | server | 服务器 |
| 诊断历史 | history | 历史 |

### 状态图标

| 状态 | 图标名 |
|------|--------|
| 成功 | check-circle |
| 错误 | x-circle |
| 警告 | alert-triangle |
| 信息 | info |
| 加载中 | loader (旋转动画) |

---

## 动效规范

### 设计原则
- 微交互反馈优先，装饰性动画禁止
- 时长控制在 0.15s - 0.35s
- 避免自动播放的动画

### 过渡时长

| 类型 | 时长 | 使用场景 |
|------|------|----------|
| micro | 0.15s | 按钮 hover, 输入框 focus |
| normal | 0.25s | 面板展开, Toast 出现 |
| slow | 0.35s | 模态框, 下拉菜单 |

### CSS 变量

```css
:root {
  --transition-micro: 0.15s ease;
  --transition-normal: 0.25s ease;
  --transition-slow: 0.35s ease;
}
```

### 组件动效

#### Button
```css
.btn {
  transition: background-color var(--transition-micro),
              border-color var(--transition-micro),
              opacity var(--transition-micro);
}
.btn:hover {
  /* 背景/边框变化 */
}
.btn:active {
  transform: scale(0.98);
}
```

#### Card
```css
.card {
  transition: border-color var(--transition-micro),
              box-shadow var(--transition-micro);
}
.card:hover {
  border-color: #D1D5DB;
}
```

#### Toast
```css
.toast-enter {
  animation: slideUp 0.25s ease;
}
.toast-exit {
  animation: fadeOut 0.15s ease forwards;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes fadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}
```

#### Modal
```css
.modal-overlay {
  transition: opacity var(--transition-normal);
}
.modal-content {
  transition: transform var(--transition-normal),
              opacity var(--transition-normal);
}
.modal-enter {
  opacity: 0;
  transform: scale(0.95);
}
.modal-enter-active {
  opacity: 1;
  transform: scale(1);
}
```

#### Skeleton
```css
.skeleton-wave {
  background: linear-gradient(
    90deg,
    #E5E5E5 0%,
    #F5F5F5 50%,
    #E5E5E5 100%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}
@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

#### Sidebar Arrow
```css
.nav-arrow {
  transition: transform var(--transition-fast);
}
.nav-arrow.expanded {
  transform: rotate(90deg);
}
```

---

## 响应式断点

### 断点定义

```css
:root {
  --breakpoint-sm: 576px;
  --breakpoint-md: 768px;
  --breakpoint-lg: 1024px;
  --breakpoint-xl: 1200px;
}
```

### 响应式布局

| 断点 | 屏幕宽度 | Sidebar | 主内容 |
|------|----------|---------|--------|
| xl | ≥1200px | 200px 完整 | 自适应 |
| lg | 1024-1199px | 160px 收窄 | 自适应 |
| md | 768-1023px | 64px 图标模式 | 自适应 |
| sm | <768px | 底部 Tab | 单列布局 |

### Sidebar 响应式行为

| 断点 | 展开方式 | 图标尺寸 |
|------|----------|----------|
| ≥1024px | 固定展开 | 16px |
| 768-1023px | 悬停展开子菜单 | 20px |
| <768px | 无侧边栏 | - |

### 表格响应式

| 断点 | 列显示 |
|------|--------|
| ≥1200px | 完整列 (平台/集群名/状态/操作/描述) |
| 768-1199px | 隐藏描述列 |
| <768px | 卡片式列表 |

### 栅格系统

```css
.grid {
  display: grid;
  gap: 16px;
}
.grid-cols-3 { grid-template-columns: repeat(3, 1fr); }
.grid-cols-2 { grid-template-columns: repeat(2, 1fr); }
.grid-cols-1 { grid-template-columns: 1fr; }

@media (max-width: 768px) {
  .grid-cols-3 { grid-template-columns: 1fr; }
  .grid-cols-2 { grid-template-columns: 1fr; }
}
```

---

**关联文档**
- `../DESIGN.md` - 设计系统基础
- 06-chat-ui.md - Chat UI
- 07-dashboard.md - Dashboard
