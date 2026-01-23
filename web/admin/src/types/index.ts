// Provider 类型定义
export interface Provider {
    id: number
    name: string
    type: string // openai, anthropic
    apiKey: string
    baseURL: string
    models?: string[] // Optional list of models
    timeoutMs: number
    isDefault: boolean
    enabled: boolean
    createdAt: number
    updatedAt: number
}

export interface CreateProviderRequest {
    name: string
    type: string
    apiKey: string
    baseURL: string
    models?: string[]
    timeoutMs: number
    isDefault: boolean
    enabled: boolean
}

// 路由规则类型定义
export interface RoutingRule {
    id: number
    ruleType: string
    pattern: string
    providerName: string
    actualModel: string
    priority: number
    enabled: boolean
    createdAt: string
    updatedAt: string
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
    id: number
    name: string
    modelPattern: string
    strategy: string
    enabled: boolean
    createdAt: string
    updatedAt: string
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
    userCount: number
    apiKeyCount: number
    usage: UsageStats
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
    totalInputTokens: number
    totalOutputTokens: number
    avgLatencyMs: number
    successCount: number
    errorCount: number
}

export interface DailyUsage {
    date: string
    requests: number
    inputTokens: number
    outputTokens: number
}
// ========== Model Rate ==========

export interface ModelRate {
    id: number
    modelPattern: string
    promptPrice: number
    completionPrice: number
    enabled: boolean
    createdAt: string
    updatedAt: string
}

export interface CreateModelRateRequest {
    modelPattern: string
    promptPrice: number
    completionPrice: number
    enabled: boolean
}

// ========== Wallet ==========

export interface Wallet {
    id: number
    userId: number
    balance: number
    updatedAt: string
}

export interface WalletTransaction {
    id: number
    walletId: number
    type: string
    amount: number
    balanceBefore: number
    balanceAfter: number
    referenceId: string
    description: string
    createdAt: string
}

export interface TopUpRequest {
    amount: number
}
