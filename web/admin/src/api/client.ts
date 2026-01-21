import axios from 'axios'

const apiClient = axios.create({
    baseURL: '/api', // Base URL changed to /api to cover /auth, /user, and /admin
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
})

// Request interceptor
apiClient.interceptors.request.use((config) => {
    const token = localStorage.getItem('token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// Response interceptor
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token')
            localStorage.removeItem('user')
            // Redirect to login if not already there
            if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
                window.location.href = '/login'
            }
        }
        return Promise.reject(error)
    }
)

export default apiClient

