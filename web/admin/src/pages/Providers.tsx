import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { providerApi } from '@/api'
import type { Provider, CreateProviderRequest } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Plus, Trash2, Edit, Check, X } from 'lucide-react'

export function Providers() {
    const queryClient = useQueryClient()
    const [showForm, setShowForm] = useState(false)
    const [editingId, setEditingId] = useState<number | null>(null)
    const [formData, setFormData] = useState<CreateProviderRequest>({
        name: '',
        type: 'openai',
        apiKey: '',
        baseURL: '',
        timeoutMs: 60000,
        isDefault: false,
        enabled: true,
    })

    const { data: providers, isLoading } = useQuery({
        queryKey: ['providers'],
        queryFn: providerApi.list,
    })

    const createMutation = useMutation({
        mutationFn: providerApi.create,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['providers'] })
            setShowForm(false)
            resetForm()
        },
    })

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: number; data: CreateProviderRequest }) =>
            providerApi.update(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['providers'] })
            setEditingId(null)
            resetForm()
        },
    })

    const deleteMutation = useMutation({
        mutationFn: providerApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['providers'] })
        },
    })

    const resetForm = () => {
        setFormData({
            name: '',
            type: 'openai',
            apiKey: '',
            baseURL: '',
            timeoutMs: 60000,
            isDefault: false,
            enabled: true,
        })
    }

    const startEdit = (provider: Provider) => {
        setEditingId(provider.id)
        setFormData({
            name: provider.name,
            type: provider.type,
            apiKey: provider.apiKey,
            baseURL: provider.baseURL,
            timeoutMs: provider.timeoutMs,
            isDefault: provider.isDefault,
            enabled: provider.enabled,
        })
    }

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        if (editingId) {
            updateMutation.mutate({ id: editingId, data: formData })
        } else {
            createMutation.mutate(formData)
        }
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold">提供商管理</h2>
                <Button onClick={() => setShowForm(true)} disabled={showForm}>
                    <Plus className="mr-2 h-4 w-4" />
                    添加提供商
                </Button>
            </div>

            {/* 创建/编辑表单 */}
            {(showForm || editingId) && (
                <Card>
                    <CardHeader>
                        <CardTitle>{editingId ? '编辑提供商' : '添加提供商'}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="grid gap-4 md:grid-cols-2">
                                <div>
                                    <label className="text-sm font-medium">名称</label>
                                    <Input
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        placeholder="my-openai"
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">类型</label>
                                    <select
                                        className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                                        value={formData.type}
                                        onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                                    >
                                        <option value="openai">OpenAI</option>
                                        <option value="anthropic">Anthropic</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="text-sm font-medium">API Key</label>
                                    <Input
                                        type="password"
                                        value={formData.apiKey}
                                        onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
                                        placeholder="sk-..."
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">Base URL</label>
                                    <Input
                                        value={formData.baseURL}
                                        onChange={(e) => setFormData({ ...formData, baseURL: e.target.value })}
                                        placeholder="https://api.openai.com/v1"
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">超时时间 (ms)</label>
                                    <Input
                                        type="number"
                                        value={formData.timeoutMs}
                                        onChange={(e) => setFormData({ ...formData, timeoutMs: parseInt(e.target.value) })}
                                    />
                                </div>
                                <div className="flex items-center gap-4 pt-6">
                                    <label className="flex items-center gap-2">
                                        <input
                                            type="checkbox"
                                            checked={formData.isDefault}
                                            onChange={(e) => setFormData({ ...formData, isDefault: e.target.checked })}
                                        />
                                        <span className="text-sm">默认</span>
                                    </label>
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
                                <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                                    <Check className="mr-2 h-4 w-4" />
                                    保存
                                </Button>
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        setShowForm(false)
                                        setEditingId(null)
                                        resetForm()
                                    }}
                                >
                                    <X className="mr-2 h-4 w-4" />
                                    取消
                                </Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            )}

            {/* 提供商列表 */}
            <Card>
                <CardHeader>
                    <CardTitle>提供商列表</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">加载中...</div>
                    ) : !providers?.length ? (
                        <div className="text-center py-8 text-muted-foreground">暂无提供商</div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full">
                                <thead>
                                    <tr className="border-b text-left text-sm text-muted-foreground">
                                        <th className="pb-3 font-medium">名称</th>
                                        <th className="pb-3 font-medium">类型</th>
                                        <th className="pb-3 font-medium">Base URL</th>
                                        <th className="pb-3 font-medium">状态</th>
                                        <th className="pb-3 font-medium">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {providers.map((provider) => (
                                        <tr key={provider.id} className="border-b last:border-0">
                                            <td className="py-3 font-medium">{provider.name}</td>
                                            <td className="py-3">
                                                <span className="rounded-full bg-secondary px-2 py-1 text-xs">
                                                    {provider.type}
                                                </span>
                                            </td>
                                            <td className="py-3 text-sm text-muted-foreground">
                                                {provider.baseURL}
                                            </td>
                                            <td className="py-3">
                                                <span
                                                    className={`rounded-full px-2 py-1 text-xs ${provider.enabled
                                                        ? 'bg-green-100 text-green-700'
                                                        : 'bg-red-100 text-red-700'
                                                        }`}
                                                >
                                                    {provider.enabled ? '启用' : '禁用'}
                                                </span>
                                                {provider.isDefault && (
                                                    <span className="ml-2 rounded-full bg-blue-100 px-2 py-1 text-xs text-blue-700">
                                                        默认
                                                    </span>
                                                )}
                                            </td>
                                            <td className="py-3">
                                                <div className="flex gap-2">
                                                    <Button
                                                        size="sm"
                                                        variant="ghost"
                                                        onClick={() => startEdit(provider)}
                                                    >
                                                        <Edit className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="ghost"
                                                        onClick={() => deleteMutation.mutate(provider.id)}
                                                        disabled={deleteMutation.isPending}
                                                    >
                                                        <Trash2 className="h-4 w-4 text-destructive" />
                                                    </Button>
                                                </div>
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
