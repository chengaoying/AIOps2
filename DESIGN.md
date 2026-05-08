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
  --text-muted: #999999;      /* 辅助说明 */

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
系统字体优先，中文友好，代码清晰。等宽字体用于代码和日志。

### 字体规范

```css
:root {
  --font-sans: "PingFang SC", "Microsoft YaHei", system-ui, -apple-system, sans-serif;
  --font-mono: "JetBrains Mono", "Fira Code", "Consolas", monospace;
}
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
│ 📊 Dashboard │   作业诊断大屏                                 │
│         │   ┌─────────────────────────────────────────┐   │
│ 🗄️ 元仓  │   │  今日概览                                  │   │
│         │   │  诊断数: 156  失败: 12  成功率: 92%   │   │
│ ─────────│   └─────────────────────────────────────────┘   │
│         │   ┌─────────────────────────────────────────┐   │
│ 🔍 作业诊断▼│   │  最近诊断                                │   │
│   ├作业诊断 │   │  [spark_job_001] Spark FAILED     │   │
│   └诊断历史 │   └─────────────────────────────────────────┘   │
│         │                                                    │
│ ─────────│                                                    │
│         │                                                    │
│ 💬 诊断助手│                                                    │
│         │                                                    │
│ ─────────│                                                    │
│         │                                                    │
│ ⚙️ 系统配置▼│                                                    │
│   ├用户管理│                                                    │
│   ├集群配置│                                                    │
│   └系统配置│                                                    │
└─────────┴───────────────────────────────────────────────────┘
```

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

### 断点

| 屏幕 | 宽度 | 布局调整 |
|------|------|----------|
| Desktop | ≥1200px | 完整侧边栏 + 主内容 |
| Tablet | 768-1199px | 侧边栏收起 |
| Mobile | <768px | 底部Tab + 单列布局 |

### 移动端优先级
1. 诊断入口
2. 诊断结果查看
3. 历史记录查看

## 动效策略

### 设计逻辑
动效最小化，仅用于状态变化反馈。

### 动效规范

```css
:root {
  --transition-fast: 0.15s ease;
  --transition-normal: 0.25s ease;
}
```

### 使用场景
- 按钮hover: 背景/边框变化，0.15s
- 卡片hover: 边框颜色变化，0.15s
- 页面切换: 无动画，直接显示

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

## 设计决策日志

| 日期 | 决策 | 理由 |
|------|------|------|
| 2026-05-04 | 初始设计系统创建 | 基于"功能优先，简单易用"原则 |
| 2026-05-04 | 采用平台色彩系统 | YARN/Hive/Spark/Flink各用独特色，便于识别 |
| 2026-05-04 | 系统字体优先 | 中文友好，无加载延迟 |
