import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Loader2, Search, ChevronLeft, ChevronRight } from 'lucide-react'
import { format } from 'date-fns'
import apiClient from '@/api/client'
import type { UsageLog, PaginatedResponse } from '@/types'

export function AuditLogs() {
    const [page, setPage] = useState(1)
    const [pageSize, setPageSize] = useState(20)
    const [filters, setFilters] = useState({
        userId: '',
        apiKeyId: '',
        clientIp: ''
    })

    const { data, isLoading, isError } = useQuery<PaginatedResponse<UsageLog>>({
        queryKey: ['audit-logs', page, pageSize, filters],
        queryFn: async () => {
            const params = new URLSearchParams({
                page: page.toString(),
                pageSize: pageSize.toString(),
                ...filters
            })
            // remove empty filters
            if (!filters.userId) params.delete('userId')
            if (!filters.apiKeyId) params.delete('apiKeyId')
            if (!filters.clientIp) params.delete('clientIp')

            const res = await apiClient.get(`/admin/usage-logs?${params.toString()}`)
            // 后端 /api/* 统一返回 { code, msg, data }
            return res.data.data
        },
    })

    const handleSearch = () => {
        setPage(1) // Reset to first page on search
        // Trigger refetch by updating state (handled by useQuery dependency)
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">审计日志</h2>
            </div>

            {/* Filters */}
            <div className="flex items-center gap-4 bg-card p-4 rounded-lg border">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 flex-1">
                    <Input
                        placeholder="用户 ID"
                        value={filters.userId}
                        onChange={(e) => setFilters(prev => ({ ...prev, userId: e.target.value }))}
                    />
                    <Input
                        placeholder="API Key ID"
                        value={filters.apiKeyId}
                        onChange={(e) => setFilters(prev => ({ ...prev, apiKeyId: e.target.value }))}
                    />
                    <Input
                        placeholder="Client IP"
                        value={filters.clientIp}
                        onChange={(e) => setFilters(prev => ({ ...prev, clientIp: e.target.value }))}
                    />
                </div>
                {/* Search is auto-triggered by state change, but keeping button for explicit feel or complex logic if needed */}
                <Button variant="secondary" onClick={handleSearch}>
                    <Search className="mr-2 h-4 w-4" />
                    筛选
                </Button>
            </div>

            <div className="rounded-md border bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>ID</TableHead>
                            <TableHead>时间</TableHead>
                            <TableHead>用户ID</TableHead>
                            <TableHead>模型</TableHead>
                            <TableHead>状态</TableHead>
                            <TableHead>耗时</TableHead>
                            <TableHead>Tokens (In/Out)</TableHead>
                            <TableHead>Client IP</TableHead>
                            <TableHead>操作</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={9} className="h-24 text-center">
                                    <div className="flex items-center justify-center gap-2">
                                        <Loader2 className="h-4 w-4 animate-spin" />
                                        加载中...
                                    </div>
                                </TableCell>
                            </TableRow>
                        ) : isError ? (
                            <TableRow>
                                <TableCell colSpan={9} className="h-24 text-center text-destructive">
                                    加载失败
                                </TableCell>
                            </TableRow>
                        ) : data?.data.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={9} className="h-24 text-center text-muted-foreground">
                                    暂无日志
                                </TableCell>
                            </TableRow>
                        ) : (
                            data?.data.map((log) => (
                                <TableRow key={log.id}>
                                    <TableCell className="font-mono text-xs">{log.id}</TableCell>
                                    <TableCell className="text-sm text-muted-foreground">
                                        {format(new Date(log.createdAt), 'MM-dd HH:mm:ss')}
                                    </TableCell>
                                    <TableCell>{log.userId}</TableCell>
                                    <TableCell>
                                        <div className="flex flex-col">
                                            <span className="font-medium">{log.model}</span>
                                            <span className="text-xs text-muted-foreground">{log.provider}</span>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={log.statusCode >= 200 && log.statusCode < 300 ? 'default' : 'destructive'}>
                                            {log.statusCode}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>
                                        <span className="text-xs text-muted-foreground">{log.latencyMs}ms</span>
                                    </TableCell>
                                    <TableCell className="font-mono text-xs">
                                        {log.inputTokens} / {log.outputTokens}
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex flex-col text-xs">
                                            <span>{log.clientIp || '-'}</span>
                                            <span className="text-[10px] text-muted-foreground truncate max-w-[150px]" title={log.userAgent}>
                                                {log.userAgent || '-'}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        {/* Placeholder for detail view if needed */}
                                        <span className="font-mono text-[10px] text-muted-foreground" title={log.requestId}>{log.requestId ? log.requestId.substring(0, 8) + '...' : '-'}</span>
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between px-2">
                <div className="flex items-center space-x-2">
                    <p className="text-sm font-medium">每页行数</p>
                    <select
                        className="h-8 w-[70px] rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                        value={pageSize.toString()}
                        onChange={(e) => {
                            setPageSize(Number(e.target.value))
                            setPage(1)
                        }}
                    >
                        {[10, 20, 50, 100].map((pageSize) => (
                            <option key={pageSize} value={`${pageSize}`}>
                                {pageSize}
                            </option>
                        ))}
                    </select>
                </div>
                <div className="flex items-center space-x-2">
                    <div className="text-sm text-muted-foreground">
                        Total {data?.total || 0}
                    </div>
                    <Button
                        variant="outline"
                        className="h-8 w-8 p-0"
                        onClick={() => setPage((old) => Math.max(old - 1, 1))}
                        disabled={page === 1 || isLoading}
                    >
                        <ChevronLeft className="h-4 w-4" />
                    </Button>
                    <div className="text-sm font-medium">
                        Page {page}
                    </div>
                    <Button
                        variant="outline"
                        className="h-8 w-8 p-0"
                        onClick={() => setPage((old) => (!data || old * pageSize >= data.total ? old : old + 1))}
                        disabled={!data || page * pageSize >= data.total || isLoading}
                    >
                        <ChevronRight className="h-4 w-4" />
                    </Button>
                </div>
            </div>
        </div>
    )
}
