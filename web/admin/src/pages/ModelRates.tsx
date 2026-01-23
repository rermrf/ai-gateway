import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
// import { useToast } from '@/hooks/use-toast' 
import { modelRateApi } from '@/api'
import type { ModelRate, CreateModelRateRequest } from '@/types'

// Simple Modal Component
function Modal({ isOpen, onClose, title, children }: { isOpen: boolean; onClose: () => void; title: string; children: React.ReactNode }) {
    if (!isOpen) return null
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div className="bg-background rounded-lg shadow-lg w-full max-w-md p-6 relative">
                <h3 className="text-lg font-semibold mb-4">{title}</h3>
                {children}
                <button onClick={onClose} className="absolute top-4 right-4 text-gray-500 hover:text-gray-700">
                    ✕
                </button>
            </div>
        </div>
    )
}

export function ModelRates() {
    const [isCreateOpen, setIsCreateOpen] = useState(false)
    const [editingRate, setEditingRate] = useState<ModelRate | null>(null)
    // const { toast } = useToast() 
    // Fallback toast if missing
    const toast = (opts: any) => console.log('Toast:', opts)

    const queryClient = useQueryClient()

    const { data: rates, isLoading } = useQuery({
        queryKey: ['modelRates'],
        queryFn: modelRateApi.list,
    })

    const createMutation = useMutation({
        mutationFn: modelRateApi.create,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['modelRates'] })
            setIsCreateOpen(false)
            toast({ title: '创建成功', description: '模型费率已创建' })
        },
        onError: (error: any) => {
            alert('创建失败: ' + (error.response?.data?.error || '未知错误'))
        },
    })

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: number; data: CreateModelRateRequest }) =>
            modelRateApi.update(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['modelRates'] })
            setEditingRate(null)
            toast({ title: '更新成功', description: '模型费率已更新' })
        },
        onError: (error: any) => {
            alert('更新失败: ' + (error.response?.data?.error || '未知错误'))
        },
    })

    const deleteMutation = useMutation({
        mutationFn: modelRateApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['modelRates'] })
            toast({ title: '删除成功', description: '模型费率已删除' })
        },
        onError: (error: any) => {
            alert('删除失败: ' + (error.response?.data?.error || '未知错误'))
        },
    })

    const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        const formData = new FormData(e.currentTarget)
        createMutation.mutate({
            modelPattern: formData.get('modelPattern') as string,
            promptPrice: parseFloat(formData.get('promptPrice') as string),
            completionPrice: parseFloat(formData.get('completionPrice') as string),
            enabled: formData.get('enabled') === 'on',
        })
    }

    const handleUpdate = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        if (!editingRate) return
        const formData = new FormData(e.currentTarget)
        updateMutation.mutate({
            id: editingRate.id,
            data: {
                modelPattern: formData.get('modelPattern') as string,
                promptPrice: parseFloat(formData.get('promptPrice') as string),
                completionPrice: parseFloat(formData.get('completionPrice') as string),
                enabled: formData.get('enabled') === 'on',
            },
        })
    }

    if (isLoading) return <div>Loading...</div>

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">模型费率</h2>
                    <p className="text-muted-foreground">
                        配置不同模型的 Token 价格（$/1M Tokens，支持通配符 *）
                    </p>
                </div>
                <Button onClick={() => setIsCreateOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    添加费率
                </Button>
            </div>

            <Modal isOpen={isCreateOpen} onClose={() => setIsCreateOpen(false)} title="添加模型费率">
                <form onSubmit={handleCreate} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="modelPattern">模型模式</Label>
                        <Input
                            id="modelPattern"
                            name="modelPattern"
                            placeholder="例如: gpt-4* 或 claude-3-opus"
                            required
                        />
                        <p className="text-xs text-muted-foreground">
                            支持通配符 * 匹配后缀，例如 gpt-4* 匹配所有 gpt-4 开头的模型
                        </p>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="promptPrice">输入价格 ($/1M)</Label>
                            <Input
                                id="promptPrice"
                                name="promptPrice"
                                type="number"
                                step="0.000001"
                                defaultValue="0.0"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="completionPrice">输出价格 ($/1M)</Label>
                            <Input
                                id="completionPrice"
                                name="completionPrice"
                                type="number"
                                step="0.000001"
                                defaultValue="0.0"
                                required
                            />
                        </div>
                    </div>
                    <div className="flex items-center space-x-2">
                        <input type="checkbox" id="enabled" name="enabled" defaultChecked className="h-4 w-4" />
                        <Label htmlFor="enabled">启用</Label>
                    </div>
                    <div className="flex justify-end space-x-2">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => setIsCreateOpen(false)}
                        >
                            取消
                        </Button>
                        <Button type="submit" disabled={createMutation.isPending}>
                            {createMutation.isPending ? '创建中...' : '创建'}
                        </Button>
                    </div>
                </form>
            </Modal>

            <div className="rounded-md border">
                <table className="w-full text-sm">
                    <thead>
                        <tr className="border-b bg-muted/50 text-left">
                            <th className="p-3 font-medium">模型模式</th>
                            <th className="p-3 font-medium">输入价格 ($/1M)</th>
                            <th className="p-3 font-medium">输出价格 ($/1M)</th>
                            <th className="p-3 font-medium">状态</th>
                            <th className="p-3 font-medium w-[100px]">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        {rates?.map((rate) => (
                            <tr key={rate.id} className="border-b last:border-0">
                                <td className="p-3 font-mono">{rate.modelPattern}</td>
                                <td className="p-3 font-mono text-muted-foreground">${rate.promptPrice}</td>
                                <td className="p-3 font-mono text-muted-foreground">${rate.completionPrice}</td>
                                <td className="p-3">
                                    <div className={`flex items-center gap-2 ${rate.enabled ? 'text-green-600' : 'text-gray-400'}`}>
                                        <div className={`h-2 w-2 rounded-full ${rate.enabled ? 'bg-green-600' : 'bg-gray-400'}`} />
                                        {rate.enabled ? '已启用' : '已禁用'}
                                    </div>
                                </td>
                                <td className="p-3">
                                    <div className="flex items-center gap-2">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => setEditingRate(rate)}
                                        >
                                            <Pencil className="h-4 w-4" />
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="text-destructive"
                                            onClick={() => {
                                                if (confirm('确定要删除这个费率配置吗？')) {
                                                    deleteMutation.mutate(rate.id)
                                                }
                                            }}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </td>
                            </tr>
                        ))}
                        {rates?.length === 0 && (
                            <tr>
                                <td colSpan={5} className="p-4 text-center text-muted-foreground h-24">
                                    暂无配置，默认免费
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>

            <Modal isOpen={!!editingRate} onClose={() => setEditingRate(null)} title="编辑模型费率">
                {editingRate && (
                    <form onSubmit={handleUpdate} className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="edit-modelPattern">模型模式</Label>
                            <Input
                                id="edit-modelPattern"
                                name="modelPattern"
                                defaultValue={editingRate.modelPattern}
                                required
                            />
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="edit-promptPrice">输入价格 ($/1M)</Label>
                                <Input
                                    id="edit-promptPrice"
                                    name="promptPrice"
                                    type="number"
                                    step="0.000001"
                                    defaultValue={editingRate.promptPrice}
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="edit-completionPrice">输出价格 ($/1M)</Label>
                                <Input
                                    id="edit-completionPrice"
                                    name="completionPrice"
                                    type="number"
                                    step="0.000001"
                                    defaultValue={editingRate.completionPrice}
                                    required
                                />
                            </div>
                        </div>
                        <div className="flex items-center space-x-2">
                            <input type="checkbox" id="edit-enabled" name="enabled" defaultChecked={editingRate.enabled} className="h-4 w-4" />
                            <Label htmlFor="edit-enabled">启用</Label>
                        </div>
                        <div className="flex justify-end space-x-2">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setEditingRate(null)}
                            >
                                取消
                            </Button>
                            <Button type="submit" disabled={updateMutation.isPending}>
                                {updateMutation.isPending ? '更新中...' : '更新'}
                            </Button>
                        </div>
                    </form>
                )}
            </Modal>
        </div>
    )
}
