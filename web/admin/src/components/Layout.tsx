import { Link, useLocation } from 'react-router-dom'
import {
    LayoutDashboard,
    Server,
    GitBranch,
    Scale,
    Coins,
    Key,
    Settings,
    Zap,
    Users,
    LogOut,
    User,
    Shield,
    Trophy
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/contexts/AuthContext'
import { Button } from '@/components/ui/button'

export function Layout({ children }: { children: React.ReactNode }) {
    const location = useLocation()
    const { user, logout } = useAuth()

    const navItems = [
        { path: '/', label: '仪表盘', icon: LayoutDashboard, roles: ['admin', 'user'] },
        { path: '/providers', label: '提供商', icon: Server, roles: ['admin'] },
        { path: '/routing-rules', label: '路由规则', icon: GitBranch, roles: ['admin'] },
        { path: '/load-balance', label: '负载均衡', icon: Scale, roles: ['admin'] },
        { path: '/model-rates', label: '模型费率', icon: Coins, roles: ['admin'] },
        { path: '/api-keys', label: '我的密钥', icon: Key, roles: ['admin', 'user'] },
        { path: '/admin/api-keys', label: '系统密钥', icon: Key, roles: ['admin'] },
        { path: '/admin/users', label: '用户管理', icon: Users, roles: ['admin'] },
        { path: '/admin/audit-logs', label: '审计日志', icon: Shield, roles: ['admin'] },
        { path: '/admin/leaderboard', label: '排行榜', icon: Trophy, roles: ['admin'] },
        { path: '/settings', label: '设置', icon: Settings, roles: ['admin', 'user'] },
    ]

    const filteredNavItems = navItems.filter(item =>
        user && item.roles.includes(user.role)
    )

    return (
        <div className="flex min-h-screen bg-background">
            {/* 侧边栏 */}
            <aside className="w-64 border-r bg-card flex flex-col">
                <div className="flex h-16 items-center gap-2 border-b px-6">
                    <Zap className="h-6 w-6 text-primary" />
                    <span className="text-lg font-bold">AI Gateway</span>
                </div>
                <nav className="flex-1 space-y-1 p-4">
                    {filteredNavItems.map((item) => {
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
                <div className="p-4 border-t">
                    <div className="flex items-center gap-3 px-3 py-2 mb-2">
                        <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                            <User className="h-4 w-4 text-primary" />
                        </div>
                        <div className="overflow-hidden">
                            <p className="text-sm font-medium truncate">{user?.username}</p>
                            <p className="text-xs text-muted-foreground truncate">{user?.role}</p>
                        </div>
                    </div>
                    <Button variant="outline" className="w-full justify-start text-muted-foreground" onClick={logout}>
                        <LogOut className="mr-2 h-4 w-4" />
                        退出登录
                    </Button>
                </div>
            </aside>

            {/* 主内容区 */}
            <main className="flex-1 overflow-auto">
                <header className="flex h-16 items-center border-b px-6 justify-between">
                    <h1 className="text-lg font-semibold">
                        {navItems.find(item => item.path === location.pathname)?.label || 'AI Gateway'}
                    </h1>
                </header>
                <div className="p-6">
                    {children}
                </div>
            </main>
        </div>
    )
}
