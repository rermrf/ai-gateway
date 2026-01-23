import React from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Layout } from '@/components/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Providers } from '@/pages/Providers'
import { RoutingRules } from '@/pages/RoutingRules'
import { LoadBalance } from '@/pages/LoadBalance'
import { ModelRates } from '@/pages/ModelRates'
import { ApiKeys } from '@/pages/ApiKeys'
import { Settings } from '@/pages/Settings'
import { Login } from '@/pages/Login'
import { Register } from '@/pages/Register'
import { AuthProvider, useAuth } from '@/contexts/AuthContext'
import { Users } from '@/pages/Users'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 分钟
      retry: 1,
    },
  },
})

function RequireAuth({ children, roles }: { children: React.ReactNode, roles?: string[] }) {
  const { isAuthenticated, isLoading, user } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (roles && user && !roles.includes(user.role)) {
    return <div className="flex items-center justify-center min-h-screen">Access Denied</div>
  }

  return children
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />

            <Route path="/" element={
              <RequireAuth>
                <Layout>
                  <Dashboard />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/providers" element={
              <RequireAuth roles={['admin']}>
                <Layout>
                  <Providers />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/routing-rules" element={
              <RequireAuth roles={['admin']}>
                <Layout>
                  <RoutingRules />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/load-balance" element={
              <RequireAuth roles={['admin']}>
                <Layout>
                  <LoadBalance />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/model-rates" element={
              <RequireAuth roles={['admin']}>
                <Layout>
                  <ModelRates />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/api-keys" element={
              <RequireAuth>
                <Layout>
                  <ApiKeys />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/admin/api-keys" element={
              <RequireAuth roles={['admin']}>
                <Layout>
                  <ApiKeys mode="admin" />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/admin/users" element={
              <RequireAuth roles={['admin']}>
                <Layout>
                  <Users />
                </Layout>
              </RequireAuth>
            } />

            <Route path="/settings" element={
              <RequireAuth>
                <Layout>
                  <Settings />
                </Layout>
              </RequireAuth>
            } />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  )
}

export default App
