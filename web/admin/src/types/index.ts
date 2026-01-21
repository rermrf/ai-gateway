// Provider 类型定义
export interface Provider {
    ID: number
    Name: string
    Type: string
    APIKey: string
    BaseURL: string
    TimeoutMs: number
    IsDefault: boolean
    Enabled: boolean
    CreatedAt: string
    UpdatedAt: string
}

export interface CreateProviderRequest {
    name: string
    type: string
    apiKey: string
    baseURL: string
    timeoutMs?: number
    isDefault?: boolean
    enabled?: boolean
}

// 路由规则类型定义
export interface RoutingRule {
    ID: number
    RuleType: string
    Pattern: string
    ProviderName: string
    ActualModel: string
    Priority: number
    Enabled: boolean
    CreatedAt: string
    UpdatedAt: string
}

export interface CreateRoutingRuleRequest {
    ruleType: string
    pattern: string
    providerName: string
    actualModel?: string
    priority?: number
    enabled?: boolean
}

// 负载均衡类型定义
export interface LoadBalanceGroup {
    ID: number
    Name: string
    ModelPattern: string
    Strategy: string
    Enabled: boolean
    CreatedAt: string
    UpdatedAt: string
}

export interface CreateLoadBalanceGroupRequest {
    name: string
    modelPattern: string
    strategy: string
    enabled?: boolean
}

// API Key 类型定义
export interface APIKey {
    id: number
    key: string
    name: string
    enabled: boolean
    createdAt: string
    expiresAt?: string
}

export interface CreateAPIKeyRequest {
    name: string
    expiresAt?: string
}

// 仪表盘统计类型定义
export interface DashboardStats {
    providerCount: number
    routingRuleCount: number
    loadBalanceCount: number
    apiKeyCount: number
}

// API 响应类型
export interface ApiResponse<T> {
    data: T
    message?: string
    error?: string
}

// ========== Auth & User Types ==========

export interface User {
    id: number
    username: string
    email: string
    role: 'user' | 'admin'
    status: 'active' | 'disabled'
    createdAt: number
}

export interface LoginResponse {
    token: string
    userId: number
    username: string
    role: string
}

export interface LoginRequest {
    username: string
    password: string
}

export interface RegisterRequest {
    username: string
    email: string
    password: string
}

export interface UpdateProfileRequest {
    email: string
}

export interface ChangePasswordRequest {
    oldPassword: string
    newPassword: string
}

export interface UsageStats {
    totalRequests: number
    totalTokens: number
    cost: number
}

export interface DailyUsage {
    date: string
    requests: number
    tokens: number
}
