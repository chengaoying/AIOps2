# 14 - 用户管理页面详细设计

**创建时间**: 2026-05-06
**状态**: 待实现

---

## 概述

用户管理页面提供用户账户的增删改查操作，支持角色权限管理。

### 页面定位

- **入口**: Sidebar "系统配置 → 用户管理"
- **功能**: 用户CRUD、角色分配、状态管理
- **用户**: 系统管理员

---

## 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ AIOps                                    [集群: 生产集群 ▼]  [用户] │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│ 📊 Dashboard│  用户管理                                         │
│             │  ┌─────────────────────────────────────────────┐ │
│ 🗄️ 元仓    │  │  [用户列表]                                   │ │
│             │  └─────────────────────────────────────────────┘ │
│ 🔍 作业诊断 ▼│  ┌─────────────────────────────────────────────┐ │
│   ├ 作业诊断 │  │  [添加用户]  [导入] [导出]                  │ │
│   └ 诊断历史 │  └─────────────────────────────────────────────┘ │
│             │  ┌─────────────────────────────────────────────┐ │
│ 💬 AI助手 │  │  ┌─────────────────────────────────────┐   │ │
│             │  │  │ 用户名    │角色  │状态   │操作    │   │ │
│ ⚙️ 系统配置▼│  │  ├─────────────────────────────────────┤   │ │
│   ├ 用户管理 │  │  │ admin    │管理员│启用   │编辑删除│   │ │
│   ├ 集群配置 │  │  │ operator │运维  │启用   │编辑删除│   │ │
│   └ 系统配置 │  │  │ viewer   │查看  │禁用   │编辑删除│   │ │
└─────────────┴──│  └─────────────────────────────────────┘   │ │
                  └─────────────────────────────────────────────┘
```

---

## 用户列表

### 表格列定义

| 列名 | 字段 | 宽度 | 说明 |
|------|------|------|------|
| 用户名 | username | 150px | 唯一标识 |
| 显示名称 | display_name | 150px | 可读名称 |
| 角色 | role | 100px | 管理员/运维/查看 |
| 邮箱 | email | 200px | 联系方式 |
| 状态 | status | 80px | 启用/禁用 |
| 创建时间 | created_at | 140px | 格式: YYYY-MM-DD |
| 操作 | actions | 120px | 编辑/删除 |

### 行样式

- Hover: #FAFAFA 背景
- 状态标识: 启用=绿色, 禁用=灰色

---

## 添加/编辑用户弹窗

### 表单字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 用户名 | 输入框 | 是 | 唯一, 3-20字符 |
| 显示名称 | 输入框 | 是 | 2-50字符 |
| 邮箱 | 输入框 | 是 | 邮箱格式验证 |
| 角色 | 下拉选择 | 是 | 管理员/运维/查看 |
| 密码 | 密码框 | 新建时必填 | 8位以上, 含大小写和数字 |
| 确认密码 | 密码框 | 是 | 与密码一致 |
| 状态 | 开关 | 否 | 默认启用 |

### 密码验证规则

- 最少8位
- 必须包含大小写字母
- 必须包含数字
- 不能与用户名相同

### 错误提示

| 错误类型 | 提示文案 |
|----------|----------|
| 用户名已存在 | 该用户名已被注册 |
| 邮箱格式错误 | 请输入有效的邮箱地址 |
| 密码强度不足 | 密码必须包含大小写字母和数字 |
| 密码不匹配 | 两次输入的密码不一致 |

---

## 角色权限

### 角色定义

| 角色 | 权限说明 |
|------|----------|
| 管理员 | 全部功能, 包括用户管理和系统配置 |
| 运维 | 作业诊断, 告警配置, 查看所有数据 |
| 查看 | 仅查看作业和诊断历史, 无写操作 |

### 权限矩阵

| 功能 | 管理员 | 运维 | 查看 |
|------|--------|------|------|
| 首页仪表盘 | ✓ | ✓ | ✓ |
| 元仓浏览 | ✓ | ✓ | ✓ |
| 作业诊断 | ✓ | ✓ | ✗ |
| 诊断历史 | ✓ | ✓ | ✓ |
| AI助手 | ✓ | ✓ | ✗ |
| 用户管理 | ✓ | ✗ | ✗ |
| 集群配置 | ✓ | ✓ | ✗ |
| 系统配置 | ✓ | ✗ | ✗ |

---

## 批量操作

```
┌─────────────────────────────────────────────────────────────┐
│  ✓ 已选择 3 项                        [批量启用] [批量禁用] [批量删除]  │
└─────────────────────────────────────────────────────────────┘
```

---

## API 接口

### 获取用户列表

```go
// GET /api/v1/settings/users
type GetUsersRequest struct {
    Search  string `form:"search"`    // 搜索用户名/邮箱
    Role   string `form:"role"`      // 角色筛选
    Status string `form:"status"`    // enabled/disabled
    Page   int    `form:"page"`
    PageSize int   `form:"page_size"`
}

type GetUsersResponse struct {
    Users      []User      `json:"users"`
    Pagination Pagination  `json:"pagination"`
}

type User struct {
    ID          int64     `json:"id"`
    Username    string    `json:"username"`
    DisplayName string    `json:"display_name"`
    Email       string    `json:"email"`
    Role        string    `json:"role"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 创建用户

```go
// POST /api/v1/settings/users
type CreateUserRequest struct {
    Username    string `json:"username"`
    DisplayName string `json:"display_name"`
    Email       string `json:"email"`
    Role        string `json:"role"`
    Password    string `json:"password"`
}

type CreateUserResponse struct {
    User User `json:"user"`
}
```

### 更新用户

```go
// PUT /api/v1/settings/users/{user_id}
type UpdateUserRequest struct {
    DisplayName string `json:"display_name,omitempty"`
    Email       string `json:"email,omitempty"`
    Role        string `json:"role,omitempty"`
    Status      string `json:"status,omitempty"`
    Password    string `json:"password,omitempty"` // 可选, 仅在需要重置时传递
}
```

### 删除用户

```go
// DELETE /api/v1/settings/users/{user_id}
type DeleteUserResponse struct {
    Success bool `json:"success"`
}
```

### 批量操作

```go
// POST /api/v1/settings/users/batch
type BatchUserRequest struct {
    Action      string   `json:"action"` // enable/disable/delete
    UserIDs     []int64  `json:"user_ids"`
}

type BatchUserResponse struct {
    Success int `json:"success"` // 成功数量
    Failed  int `json:"failed"`  // 失败数量
}
```

---

## 状态设计

### 加载状态

表格骨架屏 + 分页占位

### 空状态

```
┌─────────────────────────────────────────────────────────────┐
│  筛选: [全部角色▼] [全部状态▼]                         │
│  搜索: [用户名/邮箱...                                ]       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│              暂无用户数据                                    │
│                                                             │
│              [添加第一个用户 →]                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 删除确认

```
┌─────────────────────────────────────────────────────────────┐
│  ⚠️ 确认删除                                               │
│                                                             │
│  确定要删除用户 "operator" 吗？                              │
│  此操作不可恢复。                                           │
│                                                             │
│                          [取消] [确认删除]                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 响应式设计

| 断点 | 表格列变化 |
|------|-------------|
| ≥1200px | 完整8列 |
| 768-1199px | 隐藏"邮箱"列 |
| <768px | 卡片式列表, 折叠非关键信息 |

---

**关联文档**
- 07-dashboard.md - Dashboard 框架
- 09-component-specs.md - 通用组件规范
- 16-settings-system.md - 系统配置页面
