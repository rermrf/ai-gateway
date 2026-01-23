import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { dashboardApi, userApi, modelApi } from '@/api'
import { Users, Key, Activity, DollarSign, Database } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'

export function Dashboard() {
    const { user } = useAuth()
    const isAdmin = user?.role === 'admin'

    // Admin query
    const { data: adminStats, isLoading: isAdminLoading, error: adminError } = useQuery({
        queryKey: ['admin-stats'],
        queryFn: dashboardApi.getStats,
        enabled: isAdmin,
    })

    // User query
    const { data: userStats, isLoading: isUserLoading, error: userError } = useQuery({
        queryKey: ['user-stats'],
        queryFn: userApi.getUsage,
        enabled: !isAdmin,
    })

    const isLoading = isAdmin ? isAdminLoading : isUserLoading
    const error = isAdmin ? adminError : userError

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

    if (error) {
        return (
            <div className="p-4 text-red-500 bg-red-50 border border-red-200 rounded-md">
                加载统计数据失败: {(error as Error).message}
            </div>
        )
    }

    const cards = isAdmin ? [
        {
            title: '总用户数',
            value: adminStats?.userCount ?? 0,
            icon: Users,
            color: 'text-blue-500',
            bgColor: 'bg-blue-500/10',
        },
        {
            title: 'API 密钥数',
            value: adminStats?.apiKeyCount ?? 0,
            icon: Key,
            color: 'text-orange-500',
            bgColor: 'bg-orange-500/10',
        },
        {
            title: '总请求数',
            value: adminStats?.usage?.totalRequests ?? 0,
            icon: Activity,
            color: 'text-green-500',
            bgColor: 'bg-green-500/10',
        },
        {
            title: '总消耗 Token',
            value: ((adminStats?.usage?.totalInputTokens ?? 0) + (adminStats?.usage?.totalOutputTokens ?? 0)).toLocaleString(),
            icon: Database,
            color: 'text-purple-500',
            bgColor: 'bg-purple-500/10',
        },
    ] : [
        {
            title: '总请求数',
            value: userStats?.totalRequests ?? 0,
            icon: Activity,
            color: 'text-green-500',
            bgColor: 'bg-green-500/10',
        },
        {
            title: '消耗 Token',
            value: ((userStats?.totalInputTokens ?? 0) + (userStats?.totalOutputTokens ?? 0)).toLocaleString(),
            icon: Database,
            color: 'text-purple-500',
            bgColor: 'bg-purple-500/10',
        },
        {
            title: '平均延迟',
            value: `${userStats?.avgLatencyMs ?? 0}ms`,
            icon: Activity,
            color: 'text-yellow-500', // Changed icon/color since cost isn't ready
            bgColor: 'bg-yellow-500/10',
        },
    ]

    return (
        <div className="space-y-6">
            <h2 className="text-3xl font-bold tracking-tight">仪表盘</h2>

            {/* 统计卡片 */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {cards.map((card) => {
                    const Icon = card.icon
                    return (
                        <Card key={card.title}>
                            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                                <CardTitle className="text-sm font-medium text-muted-foreground">
                                    {card.title}
                                </CardTitle>
                                <div className={`rounded-full p-2 ${card.bgColor}`}>
                                    <Icon className={`h-4 w-4 ${card.color}`} />
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{card.value}</div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>

            {/* 欢迎信息 */}
            <Card>
                <CardHeader>
                    <CardTitle>欢迎回来, {user?.username}</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground">
                        {isAdmin
                            ? "作为管理员，您可以管理系统用户、配置 AI 提供商和路由规则，并查看全局使用情况。"
                            : "在这里您可以查看 API 使用情况，管理您的 API 密钥。"}
                    </p>
                </CardContent>
            </Card>

            {/* 用户可用模型列表 (仅限普通用户) */}
            {!isAdmin && (
                <AvailableModels />
            )}
        </div>
    )
}

function AvailableModels() {
    const { data: models, isLoading, error } = useQuery({
        queryKey: ['available-models'],
        queryFn: modelApi.listAvailable,
    })

    const { data: wallet } = useQuery({
        queryKey: ['my-wallet'],
        queryFn: userApi.getWallet,
    })

    if (isLoading) return <div className="text-muted-foreground">加载模型中...</div>
    if (error) return <div className="text-red-500">加载模型失败</div>

    return (
        <div className="grid gap-6 md:grid-cols-2">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <DollarSign className="h-5 w-5 text-green-600" />
                        当前余额
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="text-3xl font-bold text-green-700">
                        ${wallet?.balance?.toFixed(4) || '0.0000'}
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Database className="h-5 w-5 text-blue-600" />
                        可用模型列表
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-wrap gap-2">
                        {models?.map(model => (
                            <div key={model} className="px-3 py-1 bg-secondary rounded-full text-sm font-medium">
                                {model}
                            </div>
                        ))}
                        {!models?.length && <div className="text-muted-foreground">暂无可用模型</div>}
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}

