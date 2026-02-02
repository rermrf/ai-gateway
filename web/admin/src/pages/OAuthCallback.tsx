import { useEffect, useRef } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/contexts/AuthContext'
import apiClient from '@/api/client'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'

export function OAuthCallback() {
    const [searchParams] = useSearchParams()
    const navigate = useNavigate()
    const { login } = useAuth()
    const processed = useRef(false)

    useEffect(() => {
        const code = searchParams.get('code')
        const state = searchParams.get('state')

        if (!code || !state) {
            toast.error('登录失败：缺少必要的参数')
            navigate('/login')
            return
        }

        if (processed.current) return
        processed.current = true

        const handleCallback = async () => {
            try {
                // 直接调用后端回调接口
                // 注意：这里我们使用 apiClient 手动调用，以便处理响应
                const res = await apiClient.get('/oauth/linuxdo/callback', {
                    params: { code, state }
                })

                if (res.data.code === 0 && res.data.data) {
                    login(res.data.data)
                    toast.success('登录成功')
                    navigate('/')
                } else {
                    throw new Error(res.data.msg || '登录失败')
                }
            } catch (error: any) {
                console.error('OAuth callback error:', error)
                const msg = error.response?.data?.msg || error.message || '登录失败，请重试'
                toast.error(msg)
                navigate('/login')
            }
        }

        handleCallback()
    }, [searchParams, navigate, login])

    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 dark:bg-gray-900">
            <div className="text-center space-y-4">
                <Loader2 className="h-10 w-10 animate-spin text-primary mx-auto" />
                <h2 className="text-xl font-semibold">正在登录...</h2>
                <p className="text-muted-foreground">请稍候，正在验证您的身份</p>
            </div>
        </div>
    )
}
