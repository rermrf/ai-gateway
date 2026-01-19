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
    ApiResponse
} from '@/types'

// ========== Provider API ==========

export const providerApi = {
    list: async (): Promise<Provider[]> => {
        const res = await apiClient.get<ApiResponse<Provider[]>>('/providers')
        return res.data.data
    },

    get: async (id: number): Promise<Provider> => {
        const res = await apiClient.get<ApiResponse<Provider>>(`/providers/${id}`)
        return res.data.data
    },

    create: async (data: CreateProviderRequest): Promise<Provider> => {
        const res = await apiClient.post<ApiResponse<Provider>>('/providers', data)
        return res.data.data
    },

    update: async (id: number, data: CreateProviderRequest): Promise<Provider> => {
        const res = await apiClient.put<ApiResponse<Provider>>(`/providers/${id}`, data)
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/providers/${id}`)
    },
}

// ========== Routing Rule API ==========

export const routingRuleApi = {
    list: async (): Promise<RoutingRule[]> => {
        const res = await apiClient.get<ApiResponse<RoutingRule[]>>('/routing-rules')
        return res.data.data
    },

    create: async (data: CreateRoutingRuleRequest): Promise<RoutingRule> => {
        const res = await apiClient.post<ApiResponse<RoutingRule>>('/routing-rules', data)
        return res.data.data
    },

    update: async (id: number, data: CreateRoutingRuleRequest): Promise<RoutingRule> => {
        const res = await apiClient.put<ApiResponse<RoutingRule>>(`/routing-rules/${id}`, data)
        return res.data.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/routing-rules/${id}`)
    },
}

// ========== Load Balance API ==========

export const loadBalanceApi = {
    listGroups: async (): Promise<LoadBalanceGroup[]> => {
        const res = await apiClient.get<ApiResponse<LoadBalanceGroup[]>>('/load-balance-groups')
        return res.data.data
    },

    createGroup: async (data: CreateLoadBalanceGroupRequest): Promise<LoadBalanceGroup> => {
        const res = await apiClient.post<ApiResponse<LoadBalanceGroup>>('/load-balance-groups', data)
        return res.data.data
    },

    deleteGroup: async (id: number): Promise<void> => {
        await apiClient.delete(`/load-balance-groups/${id}`)
    },
}

// ========== API Key API ==========

export const apiKeyApi = {
    list: async (): Promise<APIKey[]> => {
        const res = await apiClient.get<ApiResponse<APIKey[]>>('/api-keys')
        return res.data.data
    },

    create: async (data: CreateAPIKeyRequest): Promise<{ data: APIKey; key: string }> => {
        const res = await apiClient.post<{ data: APIKey; key: string }>('/api-keys', data)
        return res.data
    },

    delete: async (id: number): Promise<void> => {
        await apiClient.delete(`/api-keys/${id}`)
    },
}

// ========== Dashboard API ==========

export const dashboardApi = {
    getStats: async (): Promise<DashboardStats> => {
        const res = await apiClient.get<ApiResponse<DashboardStats>>('/dashboard/stats')
        return res.data.data
    },
}
