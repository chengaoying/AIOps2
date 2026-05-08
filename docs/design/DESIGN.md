# Design System — AIOps 大数据智能诊断平台

## 产品上下文
- **产品类型**: B端数据诊断Dashboard
- **目标用户**: 大数据平台运维工程师、数据工程师、DevOps团队
- **核心场景**: 凌晨作业失败时，快速诊断问题根因
- **核心价值**: "30秒内给出根因分析和修复建议"
- **设计原则**: 功能优先，简单易用，减少运维压力

## 设计方向

### 美学方向
- **定位**: 实用主义 / 功能优先 (Utilitarian)
- **核心理念**: 视觉噪音归零，所有元素服务功能
- **关键词**: 快速、高效、专业、易用

### 装饰程度
- **极简**: 无渐变、无阴影、无装饰性图形
- **结构清晰**: 边界分明，信息密度高但层次清晰

### 平台色彩系统 (差异化点)
每个大数据组件使用独特色彩，用户一眼识别组件类型:
| 平台 | 颜色 | 用途 |
|------|------|------|
| YARN | #FF6B6B (红) | 资源调度 |
| Hive | #FFE66D (黄) | 数据仓库 |
| Spark | #4ECDC4 (青) | 计算引擎 |
| Flink | #45B7D1 (蓝) | 流处理 |

## 色彩系统

### 设计逻辑
色彩作为信号，不做装饰。状态色用于信息传达，平台色用于组件标识。

### 色彩规范

```css
:root {
  /* 中性色 */
  --bg-primary: #FFFFFF;      /* 主背景 */
  --bg-secondary: #F5F5F5;    /* 卡片/区块背景 */
  --bg-tertiary: #FAFAFA;     /* 斑马条纹 */
  --border: #E5E5E5;          /* 边框/分隔线 */

  --text-primary: #1A1A1A;    /* 主文字 */
  --text-secondary: #666666;  /* 次要文字 */
  --text-muted: #6B7280;      /* 辅助说明 (对比度 ≥4.5:1) */

  /* 状态色 */
  --danger: #DC2626;          /* 危险/失败 */
  --warning: #F59E0B;        /* 警告 */
  --success: #16A34A;        /* 成功/正常 */
  --info: #2563EB;           /* 信息/高亮/主操作 */

  /* 平台色 */
  --yarn: #FF6B6B;
  --hive: #FFE66D;
  --spark: #4ECDC4;
  --flink: #45B7D1;
}
```

### 状态色使用场景
| 状态 | 颜色 | 使用场景 |
|------|------|----------|
| 危险 | #DC2626 | 诊断失败、严重错误 |
| 警告 | #F59E0B | 警告提示、需要关注 |
| 成功 | #16A34A | 诊断成功、正常状态 |
| 信息 | #2563EB | 主操作按钮、链接、重点 |

## 字体系统

### 设计逻辑
网络字体优先加载，中文友好，代码清晰。等宽字体用于代码和日志。

### 字体规范

```css
:root {
  /* 标题字体 - Plus Jakarta Sans (现代、专业) */
  --font-display: "Plus Jakarta Sans", "Noto Sans SC", system-ui, sans-serif;

  /* 正文字体 - Noto Sans SC (中文友好) */
  --font-sans: "Noto Sans SC", "PingFang SC", "Microsoft YaHei", system-ui, -apple-system, sans-serif;

  /* 代码字体 - JetBrains Mono (等宽、高可读) */
  --font-mono: "JetBrains Mono", "Fira Code", "Consolas", monospace;
}
```

### 字体来源

| 字体 | 用途 | 来源 |
|------|------|------|
| Plus Jakarta Sans | 标题、导航 | Google Fonts |
| Noto Sans SC | 中文正文 | Google Fonts |
| JetBrains Mono | 代码、日志 | JetBrains |

### 网络字体 CDN

```html
<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;600&display=swap" rel="stylesheet">
```

### 图标 CDN

```html
<script src="https://unpkg.com/lucide@latest"></script>
```

### 字号层次

| 层级 | 字号 | 字重 | 用途 |
|------|------|------|------|
| 页面标题 | 20px | 600 | 主标题 |
| 区块标题 | 16px | 600 | 卡片标题、表头 |
| 正文 | 14px | 400 | 主要内容 |
| 辅助 | 12px | 400 | 说明文字、标签 |
| 代码 | 13px | 400 | 代码、日志 |

## 间距系统

### 设计逻辑
8px基准单位，紧凑但有呼吸感。

### 间距规范

```css
:root {
  --space-2xs: 2px;
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;
}
```

### 使用场景
| 间距 | 值 | 使用场景 |
|------|---|----------|
| xs | 4px | 紧凑元素间距 |
| sm | 8px | 元素内紧凑间距 |
| md | 16px | 元素间标准间距、卡片内边距 |
| lg | 24px | 区块间间距 |
| xl | 32px | 大区块间间距 |

## 布局规范

### 圆角

```css
:root {
  --radius-sm: 4px;  /* 按钮、输入框 */
  --radius-md: 8px;  /* 卡片 */
  --radius-full: 999px;  /* 标签、小徽章 */
}
```

### 边框

```css
:root {
  --border: 1px solid #E5E5E5;
}
```

### Dashboard布局

```
┌─────────────────────────────────────────────────────────────┐
│  Logo  AIOps                    用户头像                    │
├─────────┬───────────────────────────────────────────────────┤
│         │                                                    │
│ [icon] Dashboard │   作业诊断大屏                            │
│         │   ┌─────────────────────────────────────────┐   │
│ [icon] 元仓    │   │  今日概览                              │   │
│         │   │  诊断数: 156  失败: 12  成功率: 92%   │   │
│ ─────────│   └─────────────────────────────────────────┘   │
│         │   ┌─────────────────────────────────────────┐   │
│ [icon] 作业诊断▼│   │  最近诊断                                │   │
│   ├作业诊断 │   │  [spark_job_001] Spark FAILED     │   │
│   └诊断历史 │   └─────────────────────────────────────────┘   │
│         │                                                    │
│ ─────────│                                                    │
│         │                                                    │
│ [icon] AI助手│                                                    │
│         │                                                    │
│ ─────────│                                                    │
│         │                                                    │
│ [icon] 系统配置▼│                                                    │
│   ├用户管理│                                                    │
│   ├集群配置│                                                    │
│   └系统配置│                                                    │
└─────────┴───────────────────────────────────────────────────┘
```

> 注: [icon] 表示使用 Lucide Icons，详见图标系统章节

### 侧边栏
- 宽度: 200px
- 固定位置
- 当前项高亮: 蓝色背景 #EFF6FF，蓝色文字
- 导航项hover: 灰色背景 #F5F5F5
- 二级菜单: 带箭头指示器，展开时旋转90度
- 告警徽章: 显示在作业诊断/系统配置旁

## 组件规范

### 按钮

```css
.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.15s;
}

.btn-primary {
  background: var(--info);
  color: white;
}
.btn-primary:hover {
  background: #1D4ED8;
}

.btn-secondary {
  background: white;
  border: 1px solid var(--border);
}
.btn-secondary:hover {
  border-color: #CCC;
}

.btn-danger {
  background: var(--danger);
  color: white;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

### 输入框

```css
.form-input {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 14px;
}
.form-input:focus {
  outline: none;
  border-color: var(--info);
}
.form-input::placeholder {
  color: var(--text-muted);
}
```

### 诊断卡片

```css
.diagnosis-card {
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
}
.diagnosis-card-header {
  /* 平台色条 */
  border-left: 4px solid var(--平台色);
}
```

### 表格

```css
.table th {
  background: var(--bg-secondary);
  padding: 10px 12px;
  text-align: left;
  font-weight: 500;
  color: var(--text-secondary);
}
.table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}
.table tr:hover td {
  background: var(--bg-tertiary);
}
```

## 状态设计

### 设计原则
- **加载状态**: 显示进度条和当前步骤
- **空状态**: 包含友好文案和主操作按钮
- **错误状态**: 明确错误类型，提供重试操作

### 空状态设计

| 场景 | 设计 |
|------|------|
| 暂无诊断记录 | 文案: "还没有诊断记录，开始第一次诊断吧" + 按钮: "去诊断" |
| 诊断分析中 | 进度条 + 步骤显示: "规则匹配中 → 查询案例 → AI分析中" |
| 诊断失败 | 错误类型 + 重试按钮 + 联系支持选项 |

## 响应式策略

### 设计逻辑
适配不同屏幕尺寸，确保核心功能在移动端可用。

### 断点规范

```css
:root {
  --breakpoint-sm: 576px;   /* 大手机 */
  --breakpoint-md: 768px;   /* 平板 */
  --breakpoint-lg: 1024px;  /* 小屏笔记本 */
  --breakpoint-xl: 1200px;  /* 标准桌面 */
}
```

### 断点

| 屏幕 | 宽度 | 布局调整 |
|------|------|----------|
| Desktop | ≥1200px | 完整侧边栏 (200px) + 主内容 |
| Small Desktop | 1024-1199px | 侧边栏收窄 (160px) |
| Tablet | 768-1023px | 侧边栏折叠为图标模式 (64px) |
| Mobile | <768px | 底部 Tab 导航 + 单列布局 |

### 侧边栏响应式行为

| 断点 | 侧边栏状态 | 展开行为 |
|------|------------|----------|
| ≥1024px | 完整宽度 (200px) | 固定展开 |
| 768-1023px | 图标模式 (64px) | 悬停展开子菜单 |
| <768px | 底部 Tab | 点击切换页面 |

### 移动端优先级
1. 诊断入口
2. 诊断结果查看
3. 历史记录查看

### 表格响应式

| 断点 | 表格列调整 |
|------|------------|
| ≥1200px | 完整列 (平台/集群名/状态/操作) |
| 768-1199px | 隐藏描述列 |
| <768px | 卡片式列表，非关键列折叠 |

## 动效策略

### 设计逻辑
动效最小化，仅用于状态变化反馈和操作确认。

### 动效规范

```css
:root {
  --transition-fast: 0.15s ease;   /* 微交互 */
  --transition-normal: 0.25s ease;  /* 面板展开 */
  --transition-slow: 0.35s ease;   /* 页面过渡 */
}
```

### 微交互定义

| 元素 | 动效 | 时长 | 说明 |
|------|------|------|------|
| 按钮 hover | 背景色变化 | 0.15s | primary: #1D4ED8, secondary: border-color |
| 卡片 hover | 边框颜色加深 | 0.15s | #E5E5E5 → #D1D5DB |
| 输入框 focus | 边框颜色变为 info | 0.15s | 焦点状态反馈 |
| 侧边栏箭头 | 旋转 90° | 0.2s | 二级菜单展开指示 |
| Toast 出现 | 从下往上滑入 + 淡入 | 0.25s | 位置: translateY(10px) → 0 |
| Toast 消失 | 淡出 | 0.15s | opacity: 1 → 0 |
| 模态框 | 淡入 | 0.2s | opacity: 0 → 1, scale: 0.95 → 1 |
| 下拉菜单 | 高度展开 | 0.2s | max-height 过渡 |
| 骨架屏 | 波浪动画 | 1.5s | 从左到右 shimmer |

### 使用场景

```css
/* 按钮 */
.btn:hover {
  transition: background-color var(--transition-fast);
}

/* 卡片 */
.card:hover {
  border-color: #D1D5DB;
  transition: border-color var(--transition-fast);
}

/* Toast */
.toast-enter {
  animation: slideUp 0.25s ease;
}
.toast-exit {
  animation: fadeOut 0.15s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes fadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* 模态框 */
.modal-enter {
  animation: modalIn 0.2s ease;
}

@keyframes modalIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
```

### 禁止项
- 页面切换动画
- 装饰性动画
- 自动播放的轮播

## 代码字体

```css
/* 代码和日志使用等宽字体 */
.code, pre, code {
  font-family: "JetBrains Mono", "Fira Code", "Consolas", monospace;
  font-size: 13px;
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}
```

## 深色模式

### 设计逻辑
支持运维工程师在低光环境下工作，减少视觉疲劳。

### 深色模式色板

```css
:root {
  /* 深色模式 - 背景 */
  --bg-primary-dark: #1E293B;      /* 主背景 (参考 page1.png) */
  --bg-secondary-dark: #0F172A;    /* 卡片/区块背景 */
  --bg-tertiary-dark: #334155;     /* 斑马条纹/hover */
  --border-dark: #334155;          /* 边框/分隔线 */

  /* 深色模式 - 文字 */
  --text-primary-dark: #F5F5F5;     /* 主文字 */
  --text-secondary-dark: #A0A0A0;   /* 次要文字 */
  --text-muted-dark: #6B7280;      /* 辅助说明 */
}
```

### 深色模式使用场景

| 元素 | 深色模式 |
|------|----------|
| 页面背景 | #0F1117 |
| 卡片背景 | #1A1D27 |
| 边框 | #2D3344 |
| 文字 | #F5F5F5 / #A0A0A0 |

## 无障碍设计

### 设计逻辑
遵循 WCAG 2.1 AA 标准，确保所有用户可访问。

### 对比度要求

| 类型 | 要求 | 示例 |
|------|------|------|
| 正文文字 | ≥4.5:1 | #1A1A1A on #FFFFFF |
| 大文本 (≥18px/粗体≥14px) | ≥3:1 | #666666 on #FFFFFF |
| UI 组件边界 | ≥3:1 | 输入框边框 |

### 修复项

| 问题 | 修复方案 |
|------|----------|
| #999999 文字对比度不足 | 改为 #6B7280 (深色模式) / #666666 (浅色模式) |
| 链接颜色不足 | 确保 ≥4.5:1，当前 #2563EB 符合 |

### 焦点状态

```css
:focus-visible {
  outline: 2px solid #2563EB;
  outline-offset: 2px;
}

/* 深色模式 */
@media (prefers-color-scheme: dark) {
  :focus-visible {
    outline-color: #60A5FA;
  }
}
```

### ARIA 地标

```html
<nav aria-label="主导航">...</nav>
<main>...</main>
<aside aria-live="polite">告警通知区域</aside>
<footer>...</footer>
```

### 键盘导航

- 所有交互元素可通过 Tab 聚焦
- 使用 `role` 属性标识组件类型
- 模态框聚焦陷阱，Esc 关闭

## 图标系统

### 设计逻辑
使用统一风格的图标库，替换所有 emoji，保持视觉一致性。

### 图标规范

| 平台 | 图标 | 来源 |
|------|------|------|
| Dashboard | LayoutDashboard | Lucide Icons |
| 元仓 | Database | Lucide Icons |
| 作业诊断 | Search | Lucide Icons |
| AI助手 | MessageCircle | Lucide Icons |
| 系统配置 | Settings | Lucide Icons |
| 用户管理 | Users | Lucide Icons |
| 集群配置 | Server | Lucide Icons |
| 告警 | AlertTriangle | Lucide Icons |
| 成功 | CheckCircle | Lucide Icons |
| 错误 | XCircle | Lucide Icons |
| 警告 | AlertCircle | Lucide Icons |

### 图标样式

```css
.icon {
  width: 16px;
  height: 16px;
  stroke-width: 2;
  stroke: currentColor;
  fill: none;
}
```

### CDN 引入

```html
<script src="https://unpkg.com/lucide@latest"></script>
```

## 设计决策日志

| 日期 | 决策 | 理由 |
|------|------|------|
| 2026-05-04 | 初始设计系统创建 | 基于"功能优先，简单易用"原则 |
| 2026-05-04 | 采用平台色彩系统 | YARN/Hive/Spark/Flink各用独特色，便于识别 |
| 2026-05-04 | 系统字体优先 | 中文友好，无加载延迟 |
| 2026-05-07 | 添加深色模式 | 支持低光环境工作 |
| 2026-05-07 | 添加无障碍规范 | WCAG 2.1 AA 合规 |
| 2026-05-07 | 升级图标系统 | Lucide Icons 替换 emoji |
