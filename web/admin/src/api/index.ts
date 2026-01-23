import apiClient from './client'
import type {
    Provider,
    CreateProviderRequest,
    RoutingRule,
    CreateRoutingRuleRequest,
    LoadBalanceGroup,
    CreateLoadBalanceGroupRequest,
    APIKey,
    CreateAPIKeyRequest,
    DashboardStats,
    ApiResponse,
    LoginRequest,
    LoginResponse,
    RegisterRequest,
    UpdateProfileRequest,
    ChangePasswordRequest,
    UsageStats,
    DailyUsage,
    ModelRate,
    CreateModelRateRequest,
    Wallet,
    WalletTransaction,
    ModelWithPricing,

} from '@/types'

// ========== Auth API ==========

export const authApi = {
    login: async (data: LoginRequest): Promise<ApiResponse<LoginResponse>> => {
        const res = await apiClient.post<ApiResponse<{ data: LoginResponse }>>('/auth/login', data)
        // Backend returns: { data: { ... } }, but wait, let's check backend response structure.
        // auth.go Login: c.JSON(http.StatusOK, gin.H{"data": LoginResponse{...}})
        // So response.data is { data: { ... } }
        // Our ApiResponse type is { data: T, ... }
        // So existing type works if we just cast it right.
        return res.data as any
    },

    register: async (data: RegisterRequest): Promise<ApiResponse<any>> => {
        const res = await apiClient.post<ApiResponse<any>>('/auth/register', data)
        return res.data
    },
}

// ========== Model API (User) ==========

export const modelApi = {
    listAvailable: async (): Promise<string[]> => {
        const res = await apiClient.get<ApiResponse<string[]>>('/user/models')
        return res.data.data
    },

    listWithPricing: async (): Promise<ModelWithPricing[]> => {
        const res = await apiClient.get<ApiResponse<ModelWithPricing[]>>('/user/models-with-pricing')
        return res.data.data
    },
}

// ========== User API ==========

export const userApi = {
    getProfile: async (): Promise<UserResponse> => {
        const res = await apiClient.get<ApiResponse<UserResponse>>('/user/profile')
        return res.data.data
    },

    updateProfile: async (data: UpdateProfileRequest): Promise<UserResponse> => {
        const res = await apiClient.put<ApiResponse<UserResponse>>('/user/profile', data)
        return res.data.data
    },

    changePassword: async (data: ChangePasswordRequest): Promise<void> => {
        await apiClient.put('/user/password', data)
    },

    listKeys: async (): Promise<APIKey[]> => {
        const res = await apiClient.get<ApiResponse<APIKey[]>>('/user/api-keys')
        return res.data.data
    },

    createKey: async (data: CreateAPIKeyRequest): Promise<{ data: APIKey; key: string }> => {
        const res = await apiClient.post<{ data: APIKey; key: string }>('/user/api-keys', data)
        return res.data
    },

    deleteKey: async (id: number): Promise<void> => {
        await apiClient.delete(`/user/api-keys/${id}`)
    },

    getWallet: async (): Promise<Wallet> => {
        const res = await apiClient.get<ApiResponse<Wallet>>('/user/wallet')
        return res.data.data
    },

    getTransactions: async (page = 1, size = 20): Promise<{ data: WalletTransaction[], total: number }> => {
        const res = await apiClient.get<ApiResponse<{ data: WalletTransaction[], total: number }>>(`/user/wallet/transactions?page=${page}&size=${size}`)
        return res.data.data
    },

    getUsage: async (): Promise<UsageStats> => {
        const res = await apiClient.get<ApiResponse<UsageStats>>('/user/usage')
        return res.data.data
    },

    getDailyUsage: async (days: number = 30): Promise<DailyUsage[]> => {
        const res = await apiClient.get<ApiResponse<DailyUsage[]>>(`/user/usage/daily?days=${days}`)
        return res.data.data
    },
}

export const apiKeyApi = {
    list: userApi.listKeys,
    create: userApi.createKey,
    delete: userApi.deleteKey,
}

interface UserResponse {
    id: number
    username: string
    email: string
    role: 'user' | 'admin'
    status: 'active' | 'disabled'
    createdAt: number
}

// ========== Admin API (Providers) ==========

export const providerApi = {
    list: async (): Promise<Provider[]> => {
        const res = await apiClient.get<ApiResponse<Provider[]>>('/admin/providers')
        return res.data.data
    },

    get: async (id: number): Promise<Provider> => {
        const res = await apiClient.get<ApiResponse<Provider>>(`/admin/providers/${id}`)
        return res.data.data
    },

    create: async (data: CreateProviderRequest): Promise<Provider> => {
        const res = await apiClient.post<ApiResponse<Provider>>('/admin/providers', data)
        return res.data.data
    },

    update: async (id: number, data: CreateProviderRequest): Promise<Provider> => {
        const res = await apiClient.put<ApiResponse<Provider>>(`/admin/providers/${id}`, data)
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/admin/providers/${id}`)
    },
}

// ========== Admin API (Routing Rules) ==========

export const routingRuleApi = {
    list: async (): Promise<RoutingRule[]> => {
        const res = await apiClient.get<ApiResponse<RoutingRule[]>>('/admin/routing-rules')
        return res.data.data
    },

    create: async (data: CreateRoutingRuleRequest): Promise<RoutingRule> => {
        const res = await apiClient.post<ApiResponse<RoutingRule>>('/admin/routing-rules', data)
        return res.data.data
    },

    update: async (id: number, data: CreateRoutingRuleRequest): Promise<RoutingRule> => {
        const res = await apiClient.put<ApiResponse<RoutingRule>>(`/admin/routing-rules/${id}`, data)
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/admin/routing-rules/${id}`)
    },
}

// ========== Admin API (Load Balance) ==========

export const loadBalanceApi = {
    listGroups: async (): Promise<LoadBalanceGroup[]> => {
        const res = await apiClient.get<ApiResponse<LoadBalanceGroup[]>>('/admin/load-balance-groups')
        return res.data.data
    },

    createGroup: async (data: CreateLoadBalanceGroupRequest): Promise<LoadBalanceGroup> => {
        const res = await apiClient.post<ApiResponse<LoadBalanceGroup>>('/admin/load-balance-groups', data)
        return res.data.data
    },

    updateGroup: async (id: number, data: CreateLoadBalanceGroupRequest): Promise<LoadBalanceGroup> => {
        const res = await apiClient.put<ApiResponse<LoadBalanceGroup>>(`/admin/load-balance-groups/${id}`, data)
        return res.data.data
    },

    deleteGroup: async (id: number): Promise<void> => {
        await apiClient.delete(`/admin/load-balance-groups/${id}`)
    },
}

// ========== Admin API (Model Rates) ==========

export const modelRateApi = {
    list: async (): Promise<ModelRate[]> => {
        const res = await apiClient.get<ApiResponse<ModelRate[]>>('/admin/model-rates')
        return res.data.data
    },

    create: async (data: CreateModelRateRequest): Promise<ModelRate> => {
        const res = await apiClient.post<ApiResponse<ModelRate>>('/admin/model-rates', data)
        return res.data.data
    },

    update: async (id: number, data: CreateModelRateRequest): Promise<ModelRate> => {
        const res = await apiClient.put<ApiResponse<ModelRate>>(`/admin/model-rates/${id}`, data)
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/admin/model-rates/${id}`)
    },
}

// ========== Admin API (API Keys) ==========

export const adminApiKeyApi = {
    list: async (): Promise<APIKey[]> => {
        const res = await apiClient.get<ApiResponse<APIKey[]>>('/admin/api-keys')
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/admin/api-keys/${id}`)
    },
}

// ========== Admin API (Users) ==========

export const adminUserApi = {
    list: async (): Promise<UserResponse[]> => {
        const res = await apiClient.get<ApiResponse<UserResponse[]>>('/admin/users')
        return res.data.data
    },

    update: async (id: number, data: { email: string, role: string, status: string }): Promise<UserResponse> => {
        const res = await apiClient.put<ApiResponse<UserResponse>>(`/admin/users/${id}`, data)
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/admin/users/${id}`)
    },

    topUp: async (id: number, amount: number): Promise<void> => {
        await apiClient.post(`/admin/users/${id}/top-up`, { amount })
    }
}


// ========== Dashboard API ==========

export const dashboardApi = {
    getStats: async (): Promise<DashboardStats> => {
        const res = await apiClient.get<ApiResponse<DashboardStats>>('/admin/dashboard/stats')
        return res.data.data
    },
}
