import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { dashboardApi } from '@/api'
import { Server, GitBranch, Scale, Key } from 'lucide-react'

export function Dashboard() {
    const { data: stats, isLoading } = useQuery({
        queryKey: ['dashboard-stats'],
        queryFn: dashboardApi.getStats,
    })

    const cards = [
        {
            title: '提供商',
            value: stats?.providerCount ?? 0,
            icon: Server,
            color: 'text-blue-500',
            bgColor: 'bg-blue-500/10',
        },
        {
            title: '路由规则',
            value: stats?.routingRuleCount ?? 0,
            icon: GitBranch,
            color: 'text-green-500',
            bgColor: 'bg-green-500/10',
        },
        {
            title: '负载均衡组',
            value: stats?.loadBalanceCount ?? 0,
            icon: Scale,
            color: 'text-purple-500',
            bgColor: 'bg-purple-500/10',
        },
        {
            title: 'API 密钥',
            value: stats?.apiKeyCount ?? 0,
            icon: Key,
            color: 'text-orange-500',
            bgColor: 'bg-orange-500/10',
        },
    ]

    if (isLoading) {
        return (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {[1, 2, 3, 4].map((i) => (
                    <Card key={i} className="animate-pulse">
                        <CardHeader className="flex flex-row items-center justify-between pb-2">
                            <div className="h-4 w-20 rounded bg-muted" />
                            <div className="h-8 w-8 rounded bg-muted" />
                        </CardHeader>
                        <CardContent>
                            <div className="h-8 w-16 rounded bg-muted" />
                        </CardContent>
                    </Card>
                ))}
            </div>
        )
    }

    return (
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {cards.map((card) => {
                    const Icon = card.icon
                    return (
                        <Card key={card.title}>
                            <CardHeader className="flex flex-row items-center justify-between pb-2">
                                <CardTitle className="text-sm font-medium text-muted-foreground">
                                    {card.title}
                                </CardTitle>
                                <div className={`rounded-lg p-2 ${card.bgColor}`}>
                                    <Icon className={`h-4 w-4 ${card.color}`} />
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold">{card.value}</div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>

            {/* 欢迎信息 */}
            <Card>
                <CardHeader>
                    <CardTitle>欢迎使用 AI Gateway 管理后台</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground">
                        在这里您可以管理 AI 提供商、配置路由规则、设置负载均衡策略以及管理 API 密钥。
                        使用左侧导航栏快速访问各个功能模块。
                    </p>
                </CardContent>
            </Card>
        </div>
    )
}
