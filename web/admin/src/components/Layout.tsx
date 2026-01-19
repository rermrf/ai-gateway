import { Link, useLocation } from 'react-router-dom'
import {
    LayoutDashboard,
    Server,
    GitBranch,
    Scale,
    Key,
    Settings,
    Zap
} from 'lucide-react'
import { cn } from '@/lib/utils'

const navItems = [
    { path: '/', label: '仪表盘', icon: LayoutDashboard },
    { path: '/providers', label: '提供商', icon: Server },
    { path: '/routing-rules', label: '路由规则', icon: GitBranch },
    { path: '/load-balance', label: '负载均衡', icon: Scale },
    { path: '/api-keys', label: 'API 密钥', icon: Key },
    { path: '/settings', label: '设置', icon: Settings },
]

interface LayoutProps {
    children: React.ReactNode
}

export function Layout({ children }: LayoutProps) {
    const location = useLocation()

    return (
        <div className="flex min-h-screen bg-background">
            {/* 侧边栏 */}
            <aside className="w-64 border-r bg-card">
                <div className="flex h-16 items-center gap-2 border-b px-6">
                    <Zap className="h-6 w-6 text-primary" />
                    <span className="text-lg font-bold">AI Gateway</span>
                </div>
                <nav className="space-y-1 p-4">
                    {navItems.map((item) => {
                        const Icon = item.icon
                        const isActive = location.pathname === item.path
                        return (
                            <Link
                                key={item.path}
                                to={item.path}
                                className={cn(
                                    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
                                    isActive
                                        ? "bg-primary text-primary-foreground"
                                        : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                                )}
                            >
                                <Icon className="h-4 w-4" />
                                {item.label}
                            </Link>
                        )
                    })}
                </nav>
            </aside>

            {/* 主内容区 */}
            <main className="flex-1">
                <header className="flex h-16 items-center border-b px-6">
                    <h1 className="text-lg font-semibold">
                        {navItems.find(item => item.path === location.pathname)?.label || 'AI Gateway Admin'}
                    </h1>
                </header>
                <div className="p-6">
                    {children}
                </div>
            </main>
        </div>
    )
}
