import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { adminUserApi } from '@/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Trash2, Edit2, Check, X } from 'lucide-react'

// Since Select component might also be missing, I'll use standard select if needed, 
// but assuming Shadcn Select might be there or I should check.
// Checking components dir: I only saw button, card, input.
// I will just use standard HTML select to be safe and save tool calls.

export function Users() {
    const queryClient = useQueryClient()
    const [editingUser, setEditingUser] = useState<number | null>(null)
    const [editForm, setEditForm] = useState({
        role: '',
        status: ''
    })

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
        },
    })

    const deleteMutation = useMutation({
        mutationFn: adminUserApi.delete,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
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
                                                        <Button size="sm" variant="ghost" onClick={() => handleEdit(user)}>
                                                            <Edit2 className="h-4 w-4 text-blue-600" />
                                                        </Button>
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
        </div>
    )
}
