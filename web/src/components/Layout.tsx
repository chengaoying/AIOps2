import { Outlet, NavLink, useLocation } from 'react-router-dom'
import { useState } from 'react'
import {
  LayoutDashboard,
  Database,
  Search,
  History,
  MessageCircle,
  Settings,
  Users,
  Server,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import clsx from 'clsx'

interface NavItem {
  path?: string
  label: string
  icon: React.ReactNode
  children?: { path: string; label: string }[]
}

const navItems: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard', icon: <LayoutDashboard size={16} /> },
  { path: '/metastore', label: '元仓', icon: <Database size={16} /> },
  {
    label: '作业诊断',
    icon: <Search size={16} />,
    children: [
      { path: '/diagnosis/job', label: '作业诊断' },
      { path: '/diagnosis/history', label: '诊断历史' },
    ],
  },
  { path: '/assistant', label: 'AI 助手', icon: <MessageCircle size={16} /> },
  {
    label: '系统配置',
    icon: <Settings size={16} />,
    children: [
      { path: '/settings/users', label: '用户管理' },
      { path: '/settings/clusters', label: '集群配置' },
      { path: '/settings/system', label: '系统配置' },
    ],
  },
]

export default function Layout() {
  const location = useLocation()
  const [expandedMenus, setExpandedMenus] = useState<string[]>(['作业诊断', '系统配置'])

  const toggleMenu = (label: string) => {
    setExpandedMenus((prev) =>
      prev.includes(label) ? prev.filter((l) => l !== label) : [...prev, label]
    )
  }

  const isActive = (path?: string) => {
    if (!path) return false
    return location.pathname === path
  }

  const isChildActive = (children?: { path: string; label: string }[]) => {
    if (!children) return false
    return children.some((child) => location.pathname === child.path)
  }

  return (
    <div className="app-layout">
      <aside className="sidebar">
        <div style={{ padding: '16px', borderBottom: '1px solid var(--border)' }}>
          <span style={{ fontSize: '18px', fontWeight: 700, color: 'var(--info)' }}>
            AIOps
          </span>
        </div>

        <nav style={{ padding: '8px 0' }}>
          {navItems.map((item) => (
            <div key={item.label}>
              {item.path ? (
                <NavLink
                  to={item.path}
                  className={clsx('nav-item', isActive(item.path) && 'active')}
                >
                  {item.icon}
                  <span style={{ marginLeft: '8px' }}>{item.label}</span>
                </NavLink>
              ) : (
                <>
                  <div
                    className={clsx('nav-item', 'has-submenu', isChildActive(item.children) && 'active')}
                    onClick={() => toggleMenu(item.label)}
                  >
                    {item.icon}
                    <span style={{ marginLeft: '8px', flex: 1 }}>{item.label}</span>
                    {expandedMenus.includes(item.label) ? (
                      <ChevronDown size={14} />
                    ) : (
                      <ChevronRight size={14} />
                    )}
                  </div>
                  {expandedMenus.includes(item.label) && item.children && (
                    <div className="nav-submenu">
                      {item.children.map((child) => (
                        <NavLink
                          key={child.path}
                          to={child.path}
                          className={clsx('nav-subitem', isActive(child.path) && 'active')}
                        >
                          {child.label}
                        </NavLink>
                      ))}
                    </div>
                  )}
                </>
              )}
            </div>
          ))}
        </nav>

        <style>{`
          .nav-item {
            display: flex;
            align-items: center;
            padding: 10px 16px;
            cursor: pointer;
            transition: background 0.15s;
            color: var(--text-secondary);
          }
          .nav-item:hover {
            background: var(--bg-secondary);
          }
          .nav-item.active {
            background: #EFF6FF;
            color: var(--info);
          }
          .nav-item.has-submenu {
            justify-content: flex-start;
          }
          .nav-submenu {
            background: var(--bg-tertiary);
          }
          .nav-subitem {
            display: block;
            padding: 8px 24px;
            font-size: 13px;
            color: var(--text-secondary);
            cursor: pointer;
            transition: background 0.15s;
          }
          .nav-subitem:hover {
            background: var(--bg-secondary);
          }
          .nav-subitem.active {
            color: var(--info);
            background: #EFF6FF;
          }
        `}</style>
      </aside>

      <main className="main-content">
        <Outlet />
      </main>
    </div>
  )
}