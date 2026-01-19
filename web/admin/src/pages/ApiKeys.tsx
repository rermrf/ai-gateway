import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiKeyApi } from '@/api'
import type { CreateAPIKeyRequest } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Plus, Trash2, Check, X, Copy } from 'lucide-react'

export function ApiKeys() {
    const queryClient = useQueryClient()
    const [showForm, setShowForm] = useState(false)
    const [newKey, setNewKey] = useState<string | null>(null)
    const [formData, setFormData] = useState<CreateAPIKeyRequest>({
        name: '',
    })

    const { data: keys, isLoading } = useQuery({
        queryKey: ['api-keys'],
        queryFn: apiKeyApi.list,
    })

    const createMutation = useMutation({
        mutationFn: apiKeyApi.create,
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['api-keys'] })
            setNewKey(data.key)
            setFormData({ name: '' })
        },
    })

    const deleteMutation = useMutation({
        mutationFn: apiKeyApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['api-keys'] })
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
                <h2 className="text-2xl font-bold">API 密钥管理</h2>
                <Button onClick={() => { setShowForm(true); setNewKey(null) }} disabled={showForm}>
                    <Plus className="mr-2 h-4 w-4" />
                    创建密钥
                </Button>
            </div>

            {/* 新创建的密钥显示 */}
            {newKey && (
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

            {showForm && !newKey && (
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
                                        <th className="pb-3 font-medium">创建时间</th>
                                        <th className="pb-3 font-medium">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {keys.map((key) => (
                                        <tr key={key.ID} className="border-b last:border-0">
                                            <td className="py-3 font-medium">{key.Name}</td>
                                            <td className="py-3 font-mono text-sm text-muted-foreground">
                                                {key.Key}
                                            </td>
                                            <td className="py-3">
                                                <span
                                                    className={`rounded-full px-2 py-1 text-xs ${key.Enabled
                                                            ? 'bg-green-100 text-green-700'
                                                            : 'bg-red-100 text-red-700'
                                                        }`}
                                                >
                                                    {key.Enabled ? '有效' : '已禁用'}
                                                </span>
                                            </td>
                                            <td className="py-3 text-sm text-muted-foreground">
                                                {new Date(key.CreatedAt).toLocaleString()}
                                            </td>
                                            <td className="py-3">
                                                <Button
                                                    size="sm"
                                                    variant="ghost"
                                                    onClick={() => deleteMutation.mutate(key.ID)}
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
