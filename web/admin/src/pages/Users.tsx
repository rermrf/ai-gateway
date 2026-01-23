import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { adminUserApi } from '@/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Trash2, Edit2, Check, X, Coins } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
// import { useToast } from '@/hooks/use-toast'

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

export function Users() {
    const queryClient = useQueryClient()
    // const { toast } = useToast()
    const toast = (opts: any) => console.log('Toast:', opts)
    const [editingUser, setEditingUser] = useState<number | null>(null)
    const [editForm, setEditForm] = useState({
        role: '',
        status: ''
    })

    // Top-up State
    const [topUpUser, setTopUpUser] = useState<any | null>(null)
    const [topUpAmount, setTopUpAmount] = useState('10.0')

    const { data: users, isLoading } = useQuery({
        queryKey: ['users'],
        queryFn: adminUserApi.list,
    })

    const updateMutation = useMutation({
        mutationFn: (data: { id: number, role: string, status: string }) =>
            adminUserApi.update(data.id, { email: '', role: data.role, status: data.status }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
            setEditingUser(null)
            toast({ title: '更新成功', description: '用户信息已更新' })
        },
    })

    const deleteMutation = useMutation({
        mutationFn: adminUserApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
            toast({ title: '删除成功', description: '用户已删除' })
        },
    })

    const topUpMutation = useMutation({
        mutationFn: (data: { id: number, amount: number }) =>
            adminUserApi.topUp(data.id, data.amount),
        onSuccess: () => {
            setTopUpUser(null)
            toast({ title: '充值成功', description: '用户余额已更新' })
        },
        onError: (error: any) => {
            toast({
                variant: 'destructive',
                title: '充值失败',
                description: error.response?.data?.error || '未知错误',
            })
        },
    })

    const handleEdit = (user: any) => {
        setEditingUser(user.id)
        setEditForm({ role: user.role, status: user.status })
    }

    const handleSave = (id: number) => {
        updateMutation.mutate({
            id,
            role: editForm.role,
            status: editForm.status
        })
    }

    const handleTopUpSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        if (!topUpUser) return
        topUpMutation.mutate({
            id: topUpUser.id,
            amount: parseFloat(topUpAmount)
        })
    }

    return (
        <div className="space-y-6">
            <h2 className="text-2xl font-bold">用户管理</h2>

            <Card>
                <CardHeader>
                    <CardTitle>用户列表</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">加载中...</div>
                    ) : !users?.length ? (
                        <div className="text-center py-8 text-muted-foreground">暂无用户</div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="border-b text-left text-muted-foreground">
                                        <th className="pb-3 font-medium">ID</th>
                                        <th className="pb-3 font-medium">用户名</th>
                                        <th className="pb-3 font-medium">邮箱</th>
                                        <th className="pb-3 font-medium">角色</th>
                                        <th className="pb-3 font-medium">状态</th>
                                        <th className="pb-3 font-medium">注册时间</th>
                                        <th className="pb-3 font-medium">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {users.map((user) => (
                                        <tr key={user.id} className="border-b last:border-0 hover:bg-muted/50">
                                            <td className="py-3">{user.id}</td>
                                            <td className="py-3 font-medium">{user.username}</td>
                                            <td className="py-3 text-muted-foreground">{user.email}</td>
                                            <td className="py-3">
                                                {editingUser === user.id ? (
                                                    <select
                                                        className="border rounded p-1"
                                                        value={editForm.role}
                                                        onChange={e => setEditForm({ ...editForm, role: e.target.value })}
                                                    >
                                                        <option value="user">User</option>
                                                        <option value="admin">Admin</option>
                                                    </select>
                                                ) : (
                                                    <span className={`px-2 py-1 rounded-full text-xs ${user.role === 'admin' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-700'
                                                        }`}>
                                                        {user.role}
                                                    </span>
                                                )}
                                            </td>
                                            <td className="py-3">
                                                {editingUser === user.id ? (
                                                    <select
                                                        className="border rounded p-1"
                                                        value={editForm.status}
                                                        onChange={e => setEditForm({ ...editForm, status: e.target.value })}
                                                    >
                                                        <option value="active">Active</option>
                                                        <option value="disabled">Disabled</option>
                                                    </select>
                                                ) : (
                                                    <span className={`px-2 py-1 rounded-full text-xs ${user.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                                                        }`}>
                                                        {user.status}
                                                    </span>
                                                )}
                                            </td>
                                            <td className="py-3 text-muted-foreground">
                                                {new Date(user.createdAt).toLocaleDateString()}
                                            </td>
                                            <td className="py-3">
                                                <div className="flex gap-2">
                                                    {editingUser === user.id ? (
                                                        <>
                                                            <Button size="sm" variant="ghost" onClick={() => handleSave(user.id)}>
                                                                <Check className="h-4 w-4 text-green-600" />
                                                            </Button>
                                                            <Button size="sm" variant="ghost" onClick={() => setEditingUser(null)}>
                                                                <X className="h-4 w-4 text-gray-500" />
                                                            </Button>
                                                        </>
                                                    ) : (
                                                        <>
                                                            <Button
                                                                size="sm"
                                                                variant="ghost"
                                                                onClick={() => setTopUpUser(user)}
                                                                title="充值"
                                                            >
                                                                <Coins className="h-4 w-4 text-yellow-600" />
                                                            </Button>
                                                            <Button size="sm" variant="ghost" onClick={() => handleEdit(user)}>
                                                                <Edit2 className="h-4 w-4 text-blue-600" />
                                                            </Button>
                                                        </>
                                                    )}

                                                    {user.id !== 1 && ( // Prevent deleting first admin
                                                        <Button
                                                            size="sm"
                                                            variant="ghost"
                                                            onClick={() => {
                                                                if (confirm(`Confirm delete user ${user.username}?`)) {
                                                                    deleteMutation.mutate(user.id)
                                                                }
                                                            }}
                                                        >
                                                            <Trash2 className="h-4 w-4 text-destructive" />
                                                        </Button>
                                                    )}
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

            <Modal isOpen={!!topUpUser} onClose={() => setTopUpUser(null)} title={`用户充值 - ${topUpUser?.username}`}>
                <form onSubmit={handleTopUpSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="amount">充值金额 ($)</Label>
                        <Input
                            id="amount"
                            type="number"
                            step="0.01"
                            value={topUpAmount}
                            onChange={(e) => setTopUpAmount(e.target.value)}
                            required
                        />
                    </div>
                    <div className="flex justify-end space-x-2">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => setTopUpUser(null)}
                        >
                            取消
                        </Button>
                        <Button type="submit" disabled={topUpMutation.isPending}>
                            {topUpMutation.isPending ? '充值中...' : '确认充值'}
                        </Button>
                    </div>
                </form>
            </Modal>
        </div>
    )
}
