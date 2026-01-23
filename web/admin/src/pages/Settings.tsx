import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/contexts/AuthContext'
import { userApi } from '@/api'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ModelPricingList } from '@/components/ModelPricingList'

export function Settings() {
    const { user, updateUser } = useAuth()
    const [activeTab, setActiveTab] = useState('profile')

    return (
        <div className="space-y-6">
            <h2 className="text-2xl font-bold">系统设置</h2>

            <div className="flex gap-4 border-b">
                <button
                    className={`px-4 py-2 border-b-2 ${activeTab === 'profile' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
                    onClick={() => setActiveTab('profile')}
                >
                    个人资料
                </button>
                <button
                    className={`px-4 py-2 border-b-2 ${activeTab === 'password' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
                    onClick={() => setActiveTab('password')}
                >
                    修改密码
                </button>
                <button
                    className={`px-4 py-2 border-b-2 ${activeTab === 'models' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
                    onClick={() => setActiveTab('models')}
                >
                    模型及费率
                </button>
                <button
                    className={`px-4 py-2 border-b-2 ${activeTab === 'about' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
                    onClick={() => setActiveTab('about')}
                >
                    关于系统
                </button>
            </div>

            {activeTab === 'profile' && <ProfileSettings user={user} updateUser={updateUser} />}
            {activeTab === 'password' && <PasswordSettings />}
            {activeTab === 'models' && <ModelPricingList />}
            {activeTab === 'about' && <AboutSection />}
        </div>
    )
}

function ProfileSettings({ user, updateUser }: { user: any, updateUser: any }) {
    const [email, setEmail] = useState('')
    const [message, setMessage] = useState('')
    const [error, setError] = useState('')

    useEffect(() => {
        if (user) setEmail(user.email || '')
    }, [user])

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault()
        setMessage('')
        setError('')
        try {
            const updated = await userApi.updateProfile({ email })
            updateUser(updated)
            setMessage('Profile updated successfully')
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to update profile')
        }
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>个人信息</CardTitle>
                <CardDescription>更新您的基本信息</CardDescription>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSave} className="space-y-4 max-w-md">
                    {message && <Alert className="bg-green-50 text-green-700 border-green-200"><AlertDescription>{message}</AlertDescription></Alert>}
                    {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}

                    <div className="space-y-2">
                        <Label>用户名</Label>
                        <Input value={user?.username} disabled />
                    </div>
                    <div className="space-y-2">
                        <Label>邮箱</Label>
                        <Input value={email} onChange={e => setEmail(e.target.value)} type="email" />
                    </div>
                    <div className="space-y-2">
                        <Label>角色</Label>
                        <Input value={user?.role} disabled />
                    </div>
                    <Button type="submit">保存更改</Button>
                </form>
            </CardContent>
        </Card>
    )
}

function PasswordSettings() {
    const [form, setForm] = useState({ oldPassword: '', newPassword: '', confirmPassword: '' })
    const [message, setMessage] = useState('')
    const [error, setError] = useState('')

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault()
        setMessage('')
        setError('')

        if (form.newPassword !== form.confirmPassword) {
            setError('New passwords do not match')
            return
        }

        try {
            await userApi.changePassword({ oldPassword: form.oldPassword, newPassword: form.newPassword })
            setMessage('Password changed successfully')
            setForm({ oldPassword: '', newPassword: '', confirmPassword: '' })
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to change password')
        }
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>修改密码</CardTitle>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSave} className="space-y-4 max-w-md">
                    {message && <Alert className="bg-green-50 text-green-700 border-green-200"><AlertDescription>{message}</AlertDescription></Alert>}
                    {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}

                    <div className="space-y-2">
                        <Label>当前密码</Label>
                        <Input type="password" value={form.oldPassword} onChange={e => setForm({ ...form, oldPassword: e.target.value })} required />
                    </div>
                    <div className="space-y-2">
                        <Label>新密码</Label>
                        <Input type="password" value={form.newPassword} onChange={e => setForm({ ...form, newPassword: e.target.value })} required />
                    </div>
                    <div className="space-y-2">
                        <Label>确认新密码</Label>
                        <Input type="password" value={form.confirmPassword} onChange={e => setForm({ ...form, confirmPassword: e.target.value })} required />
                    </div>
                    <Button type="submit">修改密码</Button>
                </form>
            </CardContent>
        </Card>
    )
}

function AboutSection() {
    return (
        <Card>
            <CardHeader>
                <CardTitle>关于</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
                <div>
                    <h3 className="font-medium">AI Gateway</h3>
                    <p className="text-sm text-muted-foreground">
                        一个通用的 AI 网关服务，提供标准的 OpenAI 和 Anthropic 兼容接口，
                        实现不同 LLM 提供商之间的协议转换。
                    </p>
                </div>
                <div className="space-y-2 text-sm">
                    <p><strong>功能特性：</strong></p>
                    <ul className="list-disc list-inside text-muted-foreground space-y-1">
                        <li>双向协议兼容（OpenAI ↔ Anthropic）</li>
                        <li>流式响应支持 (Server-Sent Events)</li>
                        <li>工具/函数调用 (Tool Calling)</li>
                        <li>多模态支持（图片/视觉）</li>
                        <li>灵活的模型路由</li>
                        <li>负载均衡策略</li>
                    </ul>
                </div>
            </CardContent>
        </Card>
    )
}
