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
    const [successMessage] = useState(location.state?.message || '')
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

                        <div className="relative my-4">
                            <div className="absolute inset-0 flex items-center">
                                <span className="w-full border-t" />
                            </div>
                            <div className="relative flex justify-center text-xs uppercase">
                                <span className="bg-background px-2 text-muted-foreground bg-white dark:bg-gray-800">
                                    或者
                                </span>
                            </div>
                        </div>

                        <Button
                            variant="outline"
                            type="button"
                            className="w-full"
                            onClick={() => window.location.href = '/api/oauth/linuxdo/login'}
                            disabled={loading}
                        >
                            <svg className="mr-2 h-4 w-4" aria-hidden="true" focusable="false" data-prefix="fab" data-icon="linux" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512">
                                <path fill="currentColor" d="M220.6 65.3c13.9 11.2 38.6 30.2 56.4 57.6 15 23.3 32.2 46.5 28.5 76zM448 320c-26.3 3.6-76.3-24.1-98-33.6-11.2-4.9-19.9-15.1-23.7-27.4-4-12.8 1.4-19 36.1-41.9 17.6-11.6 40-38.6 20.9-63-35.3-45.3-95.2-126-177.3-118-24.4 2.4-78.2 26.5-62.7 132.2 4.1 27.6 8.7 54.3-18.4 69.4l-11.1 5.4c-9.5 4-22 6.8-23.2 22.1-1.2 16.3 15.6 14.1 23.1 33.5s29.6 112 56.1 112c32.7 0 28.3-25.7 34.6-47.5 5.6-19.1 19.1-19.4 19.1-19.4 28.9-8.4 56.5 23 83.3 23 23.4 0 54.8-31 83.3-23 0 0 13.5 .2 19.1 19.4 6.3 21.8 1.9 47.5 34.6 47.5 26.5 0 48.6-92.5 56.1-112s24.3-17.2 23.1-33.5c-1.3-15.3-13.8-18.1-23.2-22.1zM152.4 207c0-21.4 17.5-38.8 39.1-38.8 21.6 0 39.1 17.4 39.1 38.8 0 21.4-17.5 38.8-39.1 38.8-21.6 0-39.1-17.4-39.1-38.8zm220.4 1.3c-21.6 0-39.1-17.4-39.1-38.8 0-21.4 17.5-38.8 39.1-38.8 21.6 0 39.1 17.4 39.1 38.8 0 21.4-17.5 38.8-39.1 38.8z" />
                            </svg>
                            使用 Linux Do 登录
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
