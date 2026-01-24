import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiKeyApi, adminApiKeyApi } from '@/api' // Import adminApiKeyApi
import type { CreateAPIKeyRequest } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Plus, Trash2, Check, X, Copy } from 'lucide-react'

interface ApiKeysProps {
    mode?: 'user' | 'admin'
}

export function ApiKeys({ mode = 'user' }: ApiKeysProps) {
    const isAdmin = mode === 'admin'
    const queryClient = useQueryClient()
    const [showForm, setShowForm] = useState(false)
    const [newKey, setNewKey] = useState<string | null>(null)
    const [formData, setFormData] = useState<CreateAPIKeyRequest>({
        name: '',
        enabled: true,
        quota: undefined,
        expiresAt: undefined
    })

    const { data: keys, isLoading } = useQuery({
        queryKey: ['api-keys', mode],
        queryFn: isAdmin ? adminApiKeyApi.list : apiKeyApi.list,
    })

    const createMutation = useMutation({
        mutationFn: apiKeyApi.create,
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['api-keys', mode] })
            setNewKey(data.key)
            setFormData({ name: '', enabled: true, quota: undefined, expiresAt: undefined })
        },
    })

    const deleteMutation = useMutation({
        mutationFn: isAdmin ? adminApiKeyApi.delete : apiKeyApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['api-keys', mode] })
        },
    })

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        createMutation.mutate(formData)
    }

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text)
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold">{isAdmin ? '系统密钥管理' : 'API 密钥管理'}</h2>
                {!isAdmin && (
                    <Button onClick={() => { setShowForm(true); setNewKey(null) }} disabled={showForm}>
                        <Plus className="mr-2 h-4 w-4" />
                        创建密钥
                    </Button>
                )}
            </div>

            {/* 新创建的密钥显示 (User Only) */}
            {newKey && !isAdmin && (
                <Card className="border-green-200 bg-green-50">
                    <CardHeader>
                        <CardTitle className="text-green-700">密钥创建成功</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-green-600 mb-2">
                            请立即复制此密钥，它将不会再次显示：
                        </p>
                        <div className="flex items-center gap-2 p-3 bg-white rounded-md border">
                            <code className="flex-1 font-mono text-sm">{newKey}</code>
                            <Button size="sm" variant="outline" onClick={() => copyToClipboard(newKey)}>
                                <Copy className="h-4 w-4" />
                            </Button>
                        </div>
                        <Button
                            className="mt-4"
                            variant="outline"
                            onClick={() => { setNewKey(null); setShowForm(false) }}
                        >
                            完成
                        </Button>
                    </CardContent>
                </Card>
            )}

            {showForm && !newKey && !isAdmin && (
                <Card>
                    <CardHeader>
                        <CardTitle>创建 API 密钥</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div>
                                <label className="text-sm font-medium">名称</label>
                                <Input
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    placeholder="my-api-key"
                                    required
                                />
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="text-sm font-medium">额度限制 (留空为无限)</label>
                                    <Input
                                        type="number"
                                        min="0"
                                        step="0.01"
                                        value={formData.quota || ''}
                                        onChange={(e) => setFormData({
                                            ...formData,
                                            quota: e.target.value ? parseFloat(e.target.value) : undefined
                                        })}
                                        placeholder="例如: 100.00"
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">过期时间</label>
                                    <Input
                                        type="datetime-local"
                                        value={formData.expiresAt ? new Date(formData.expiresAt).toISOString().slice(0, 16) : ''}
                                        onChange={(e) => setFormData({
                                            ...formData,
                                            expiresAt: e.target.value ? new Date(e.target.value).toISOString() : undefined
                                        })}
                                    />
                                </div>
                            </div>

                            <div className="flex items-center gap-2">
                                <label className="text-sm font-medium">启用状态</label>
                                <input
                                    type="checkbox"
                                    checked={formData.enabled ?? true}
                                    onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                                    className="h-4 w-4"
                                />
                                <span className="text-sm text-muted-foreground">
                                    {(formData.enabled ?? true) ? '启用' : '禁用'}
                                </span>
                            </div>
                            <div className="flex gap-2">
                                <Button type="submit" disabled={createMutation.isPending}>
                                    <Check className="mr-2 h-4 w-4" />
                                    创建
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
                    <CardTitle>密钥列表</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">加载中...</div>
                    ) : !keys?.length ? (
                        <div className="text-center py-8 text-muted-foreground">暂无 API 密钥</div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full">
                                <thead>
                                    <tr className="border-b text-left text-sm text-muted-foreground">
                                        <th className="pb-3 font-medium">名称</th>
                                        <th className="pb-3 font-medium">密钥</th>
                                        <th className="pb-3 font-medium">状态</th>
                                        <th className="pb-3 font-medium">额度使用 (已用 / 总额)</th>
                                        <th className="pb-3 font-medium">有效期</th>
                                        <th className="pb-3 font-medium">创建时间</th>
                                        <th className="pb-3 font-medium">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {keys.map((key) => (
                                        <tr key={key.id} className="border-b last:border-0">
                                            <td className="py-3 font-medium">{key.name}</td>
                                            <td className="py-3 font-mono text-sm text-muted-foreground">
                                                {key.key}
                                            </td>
                                            <td className="py-3">
                                                <span
                                                    className={`rounded-full px-2 py-1 text-xs ${key.enabled
                                                        ? 'bg-green-100 text-green-700'
                                                        : 'bg-red-100 text-red-700'
                                                        }`}
                                                >
                                                    {key.enabled ? '有效' : '已禁用'}
                                                </span>
                                            </td>
                                            <td className="py-3 text-sm">
                                                {key.quota === null ? (
                                                    <span className="text-green-600">无限</span>
                                                ) : (
                                                    <span>
                                                        {key.usedAmount.toFixed(4)} / {key.quota.toFixed(2)}
                                                        {key.usedAmount >= key.quota && (
                                                            <span className="ml-2 text-red-500 text-xs">(已超限)</span>
                                                        )}
                                                    </span>
                                                )}
                                            </td>
                                            <td className="py-3 text-sm text-muted-foreground">
                                                {key.expiresAt ? new Date(key.expiresAt).toLocaleString() : '永久有效'}
                                            </td>
                                            <td className="py-3 text-sm text-muted-foreground">
                                                {new Date(key.createdAt).toLocaleString()}
                                            </td>
                                            <td className="py-3">
                                                <Button
                                                    size="sm"
                                                    variant="ghost"
                                                    onClick={() => {
                                                        if (confirm('确定要删除此密钥吗？')) {
                                                            deleteMutation.mutate(key.id)
                                                        }
                                                    }}
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
