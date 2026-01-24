import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Loader2, Trophy } from 'lucide-react'
import apiClient from '@/api/client'

interface LeaderboardEntry {
    dimension: string
    value: string
    requestCount: number
    inputTokens: number
    outputTokens: number
    totalTokens: number
}

interface ApiResponse<T> {
    data: T
}

export function UsageLeaderboard() {
    const [dimension, setDimension] = useState('user_id')
    const [days, setDays] = useState(30)
    const [limit, setLimit] = useState(10)

    const { data, isLoading, isError } = useQuery<LeaderboardEntry[]>({
        queryKey: ['usage-leaderboard', dimension, days, limit],
        queryFn: async () => {
            const res = await apiClient.get<ApiResponse<LeaderboardEntry[]>>(`/admin/usage/leaderboard`, {
                params: { dimension, days, limit }
            })
            return res.data.data
        },
    })

    const dimensionLabels: Record<string, string> = {
        'user_id': '用户 ID',
        'api_key_id': 'API Key ID',
        'client_ip': 'Client IP'
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">使用量排行榜</h2>
            </div>

            <div className="flex items-center gap-4 bg-card p-4 rounded-lg border">
                <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">统计维度:</span>
                    <select
                        className="h-8 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        value={dimension}
                        onChange={(e) => setDimension(e.target.value)}
                    >
                        <option value="user_id">用户</option>
                        <option value="api_key_id">API Key</option>
                        <option value="client_ip">IP 地址</option>
                    </select>
                </div>

                <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">时间范围:</span>
                    <select
                        className="h-8 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        value={days}
                        onChange={(e) => setDays(Number(e.target.value))}
                    >
                        <option value={1}>最近 24 小时</option>
                        <option value={7}>最近 7 天</option>
                        <option value={30}>最近 30 天</option>
                        <option value={90}>最近 3 个月</option>
                    </select>
                </div>

                <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">显示数量:</span>
                    <select
                        className="h-8 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        value={limit}
                        onChange={(e) => setLimit(Number(e.target.value))}
                    >
                        <option value={5}>Top 5</option>
                        <option value={10}>Top 10</option>
                        <option value={20}>Top 20</option>
                        <option value={50}>Top 50</option>
                    </select>
                </div>
            </div>

            <div className="rounded-md border bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[100px]">排名</TableHead>
                            <TableHead>{dimensionLabels[dimension]}</TableHead>
                            <TableHead>请求数</TableHead>
                            <TableHead>输入 Tokens</TableHead>
                            <TableHead>输出 Tokens</TableHead>
                            <TableHead>总 Tokens</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center">
                                    <div className="flex items-center justify-center gap-2">
                                        <Loader2 className="h-4 w-4 animate-spin" />
                                        加载中...
                                    </div>
                                </TableCell>
                            </TableRow>
                        ) : isError ? (
                            <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center text-destructive">
                                    加载失败
                                </TableCell>
                            </TableRow>
                        ) : data?.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                                    暂无数据
                                </TableCell>
                            </TableRow>
                        ) : (
                            data?.map((entry, index) => (
                                <TableRow key={index}>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            {index < 3 ? (
                                                <Trophy className={`h-4 w-4 ${index === 0 ? 'text-yellow-500' :
                                                        index === 1 ? 'text-gray-400' :
                                                            'text-amber-600'
                                                    }`} />
                                            ) : (
                                                <span className="w-4 text-center font-mono text-muted-foreground">{index + 1}</span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell className="font-mono">{entry.value || '-'}</TableCell>
                                    <TableCell>
                                        <Badge variant="secondary" className="font-mono">
                                            {entry.requestCount}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="font-mono text-muted-foreground">{entry.inputTokens.toLocaleString()}</TableCell>
                                    <TableCell className="font-mono text-muted-foreground">{entry.outputTokens.toLocaleString()}</TableCell>
                                    <TableCell className="font-bold font-mono">{entry.totalTokens.toLocaleString()}</TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    )
}
