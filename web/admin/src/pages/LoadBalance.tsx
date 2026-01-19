import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { loadBalanceApi } from '@/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Trash2 } from 'lucide-react'

export function LoadBalance() {
    const queryClient = useQueryClient()

    const { data: groups, isLoading } = useQuery({
        queryKey: ['load-balance-groups'],
        queryFn: loadBalanceApi.listGroups,
    })

    const deleteMutation = useMutation({
        mutationFn: loadBalanceApi.deleteGroup,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['load-balance-groups'] })
        },
    })

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold">负载均衡管理</h2>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>负载均衡组列表</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">加载中...</div>
                    ) : !groups?.length ? (
                        <div className="text-center py-8 text-muted-foreground">暂无负载均衡组</div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full">
                                <thead>
                                    <tr className="border-b text-left text-sm text-muted-foreground">
                                        <th className="pb-3 font-medium">名称</th>
                                        <th className="pb-3 font-medium">模型模式</th>
                                        <th className="pb-3 font-medium">策略</th>
                                        <th className="pb-3 font-medium">状态</th>
                                        <th className="pb-3 font-medium">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {groups.map((group) => (
                                        <tr key={group.ID} className="border-b last:border-0">
                                            <td className="py-3 font-medium">{group.Name}</td>
                                            <td className="py-3 font-mono text-sm">{group.ModelPattern}</td>
                                            <td className="py-3">
                                                <span className="rounded-full bg-secondary px-2 py-1 text-xs">
                                                    {group.Strategy}
                                                </span>
                                            </td>
                                            <td className="py-3">
                                                <span
                                                    className={`rounded-full px-2 py-1 text-xs ${group.Enabled
                                                            ? 'bg-green-100 text-green-700'
                                                            : 'bg-red-100 text-red-700'
                                                        }`}
                                                >
                                                    {group.Enabled ? '启用' : '禁用'}
                                                </span>
                                            </td>
                                            <td className="py-3">
                                                <Button
                                                    size="sm"
                                                    variant="ghost"
                                                    onClick={() => deleteMutation.mutate(group.ID)}
                                                    disabled={deleteMutation.isPending}
                                                >
                                                    <Trash2 className="h-4 w-4 text-destructive" />
                                                </Button>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>策略说明</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm text-muted-foreground">
                    <p><strong>round-robin</strong>: 轮询策略，依次选择每个提供商</p>
                    <p><strong>random</strong>: 随机策略，随机选择一个提供商</p>
                    <p><strong>weighted</strong>: 加权策略，按权重随机选择</p>
                    <p><strong>failover</strong>: 故障转移，优先使用主要提供商，失败时切换备用</p>
                </CardContent>
            </Card>
        </div>
    )
}
