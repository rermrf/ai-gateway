import axios from 'axios'

// 开发环境默认凭据
const DEFAULT_CREDENTIALS = btoa('admin:admin')

const apiClient = axios.create({
    baseURL: '/api/admin',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
})

// 添加请求拦截器（用于 Basic Auth）
apiClient.interceptors.request.use((config) => {
    // 从 localStorage 获取凭据，如果没有则使用默认凭据
    const credentials = localStorage.getItem('adminCredentials') || DEFAULT_CREDENTIALS
    config.headers.Authorization = `Basic ${credentials}`
    return config
})

// 响应拦截器
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            // 清除凭据并重定向到登录
            localStorage.removeItem('adminCredentials')
            // 开发环境暂不跳转
            console.error('Unauthorized - please check credentials')
        }
        return Promise.reject(error)
    }
)

export default apiClient

