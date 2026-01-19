import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { routingRuleApi } from '@/api'
import type { RoutingRule, CreateRoutingRuleRequest } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Plus, Trash2, Check, X } from 'lucide-react'

export function RoutingRules() {
    const queryClient = useQueryClient()
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState<CreateRoutingRuleRequest>({
        ruleType: 'exact',
        pattern: '',
        providerName: '',
        actualModel: '',
        priority: 0,
        enabled: true,
    })

    const { data: rules, isLoading } = useQuery({
        queryKey: ['routing-rules'],
        queryFn: routingRuleApi.list,
    })

    const createMutation = useMutation({
        mutationFn: routingRuleApi.create,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['routing-rules'] })
            setShowForm(false)
            setFormData({ ruleType: 'exact', pattern: '', providerName: '', actualModel: '', priority: 0, enabled: true })
        },
    })

    const deleteMutation = useMutation({
        mutationFn: routingRuleApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['routing-rules'] })
        },
    })

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        createMutation.mutate(formData)
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold">路由规则管理</h2>
                <Button onClick={() => setShowForm(true)} disabled={showForm}>
                    <Plus className="mr-2 h-4 w-4" />
                    添加规则
                </Button>
            </div>

            {showForm && (
                <Card>
                    <CardHeader>
                        <CardTitle>添加路由规则</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="grid gap-4 md:grid-cols-2">
                                <div>
                                    <label className="text-sm font-medium">规则类型</label>
                                    <select
                                        className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                                        value={formData.ruleType}
                                        onChange={(e) => setFormData({ ...formData, ruleType: e.target.value })}
                                    >
                                        <option value="exact">精确匹配</option>
                                        <option value="prefix">前缀匹配</option>
                                        <option value="wildcard">通配符</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="text-sm font-medium">模式</label>
                                    <Input
                                        value={formData.pattern}
                                        onChange={(e) => setFormData({ ...formData, pattern: e.target.value })}
                                        placeholder="gpt-4 或 deepseek-*"
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">提供商名称</label>
                                    <Input
                                        value={formData.providerName}
                                        onChange={(e) => setFormData({ ...formData, providerName: e.target.value })}
                                        placeholder="openai"
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">实际模型（可选）</label>
                                    <Input
                                        value={formData.actualModel}
                                        onChange={(e) => setFormData({ ...formData, actualModel: e.target.value })}
                                        placeholder="gpt-4-turbo"
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">优先级</label>
                                    <Input
                                        type="number"
                                        value={formData.priority}
                                        onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                                    />
                                </div>
                                <div className="flex items-center gap-4 pt-6">
                                    <label className="flex items-center gap-2">
                                        <input
                                            type="checkbox"
                                            checked={formData.enabled}
                                            onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                                        />
                                        <span className="text-sm">启用</span>
                                    </label>
                                </div>
                            </div>
                            <div className="flex gap-2">
                                <Button type="submit" disabled={createMutation.isPending}>
                                    <Check className="mr-2 h-4 w-4" />
                                    保存
                                </Button>
                                <Button type="button" variant="outline" onClick={() => setShowForm(false)}>
                                    <X className="mr-2 h-4 w-4" />
                                    取消
                                </Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            )}

            <Card>
                <CardHeader>
                    <CardTitle>规则列表</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">加载中...</div>
                    ) : !rules?.length ? (
                        <div className="text-center py-8 text-muted-foreground">暂无路由规则</div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full">
                                <thead>
                                    <tr className="border-b text-left text-sm text-muted-foreground">
                                        <th className="pb-3 font-medium">类型</th>
                                        <th className="pb-3 font-medium">模式</th>
                                        <th className="pb-3 font-medium">提供商</th>
                                        <th className="pb-3 font-medium">实际模型</th>
                                        <th className="pb-3 font-medium">优先级</th>
                                        <th className="pb-3 font-medium">状态</th>
                                        <th className="pb-3 font-medium">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {rules.map((rule: RoutingRule) => (
                                        <tr key={rule.ID} className="border-b last:border-0">
                                            <td className="py-3">
                                                <span className="rounded-full bg-secondary px-2 py-1 text-xs">
                                                    {rule.RuleType}
                                                </span>
                                            </td>
                                            <td className="py-3 font-mono text-sm">{rule.Pattern}</td>
                                            <td className="py-3">{rule.ProviderName}</td>
                                            <td className="py-3 text-sm text-muted-foreground">
                                                {rule.ActualModel || '-'}
                                            </td>
                                            <td className="py-3">{rule.Priority}</td>
                                            <td className="py-3">
                                                <span
                                                    className={`rounded-full px-2 py-1 text-xs ${rule.Enabled
                                                            ? 'bg-green-100 text-green-700'
                                                            : 'bg-red-100 text-red-700'
                                                        }`}
                                                >
                                                    {rule.Enabled ? '启用' : '禁用'}
                                                </span>
                                            </td>
                                            <td className="py-3">
                                                <Button
                                                    size="sm"
                                                    variant="ghost"
                                                    onClick={() => deleteMutation.mutate(rule.ID)}
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
        </div>
    )
}
