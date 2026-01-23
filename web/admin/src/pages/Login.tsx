import { useState } from 'react'
import { useNavigate, Link, useLocation } from 'react-router-dom'
import { useAuth } from '@/contexts/AuthContext'
import { authApi } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { AlertCircle, CheckCircle } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function Login() {
    const navigate = useNavigate()
    const location = useLocation()
    const { login } = useAuth()
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [successMessage, setSuccessMessage] = useState(location.state?.message || '')
    const [loading, setLoading] = useState(false)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError('')
        setLoading(true)

        try {
            const res = await authApi.login({ username, password })
            // Ensure we handle the structure correctly depending on how authApi returns data
            // authApi.login returns res.data which is of type LoginResponse (after our fix)
            // But wait, the return type in authApi.login was: Promise<ApiResponse<LoginResponse>> ? 
            // No, in api/index.ts we did: return res.data as any. 
            // And res.data from axios is the JSON body.
            // Backend returns { data: { token: ... } }
            // So res.data is { data: { token: ... } }
            // So we need to access res.data.data

            // Let's double check api/index.ts implementation I just wrote:
            // const res = await apiClient.post<ApiResponse<{ data: LoginResponse }>>('/auth/login', data)
            // return res.data as any
            // If backend returns { data: { ... } }, then res.data is that object.
            // So the return value is { data: LoginResponse }

            if (res.data) {
                login(res.data)
                navigate('/')
            } else {
                setError('Login failed: No data received')
            }
        } catch (err: any) {
            console.error(err)
            setError(err.response?.data?.error || '登录失败，请检查用户名和密码')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="flex items-center justify-center min-h-screen bg-gray-100 dark:bg-gray-900 px-4">
            <Card className="w-full max-w-md">
                <CardHeader className="space-y-1">
                    <CardTitle className="text-2xl font-bold text-center">登录 AI Gateway</CardTitle>
                    <CardDescription className="text-center">
                        请输入您的账号密码
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {successMessage && (
                            <Alert className="bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-900">
                                <CheckCircle className="h-4 w-4" />
                                <AlertDescription>{successMessage}</AlertDescription>
                            </Alert>
                        )}
                        {error && (
                            <Alert variant="destructive">
                                <AlertCircle className="h-4 w-4" />
                                <AlertDescription>{error}</AlertDescription>
                            </Alert>
                        )}
                        <div className="space-y-2">
                            <Label htmlFor="username">用户名</Label>
                            <Input
                                id="username"
                                type="text"
                                placeholder="输入用户名"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <Label htmlFor="password">密码</Label>
                            </div>
                            <Input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                required
                            />
                        </div>
                        <Button type="submit" className="w-full" disabled={loading}>
                            {loading ? '登录中...' : '登录'}
                        </Button>
                    </form>
                </CardContent>
                <CardFooter className="flex justify-center">
                    <p className="text-sm text-muted-foreground">
                        还没有账号？{' '}
                        <Link to="/register" className="text-primary hover:underline">
                            立即注册
                        </Link>
                    </p>
                </CardFooter>
            </Card>
        </div>
    )
}
