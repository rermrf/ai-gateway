import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import apiClient from '@/api/client'
import type { User, LoginResponse } from '@/types'

interface AuthContextType {
    user: User | null
    token: string | null
    isAuthenticated: boolean
    isLoading: boolean
    login: (data: LoginResponse) => void
    logout: () => void
    updateUser: (user: User) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
    const [isLoading, setIsLoading] = useState(true)

    useEffect(() => {
        const storedToken = localStorage.getItem('token')
        const storedUser = localStorage.getItem('user')

        if (storedToken && storedUser) {
            setToken(storedToken)
            try {
                setUser(JSON.parse(storedUser))
            } catch (e) {
                console.error("Failed to parse user from local storage", e)
                localStorage.removeItem('user')
            }
        }
        setIsLoading(false)
    }, [])

    const login = (data: LoginResponse) => {
        const userObj: User = {
            id: data.userId,
            username: data.username,
            role: data.role as 'user' | 'admin',
            email: '', // Login response might not have email, we might need to fetch profile or store whatever we have
            status: 'active', // Assumed active if logged in
            createdAt: 0 // We don't have this in login response usually unless we change backend
        }

        // Better: Fetch full profile after login or adjust backend login response.
        // For now, let's just store what we have and maybe fetch profile asynchronously if needed.
        // Actually, let's just use what we have and rely on profile fetch later.

        localStorage.setItem('token', data.token)
        localStorage.setItem('user', JSON.stringify(userObj))
        setToken(data.token)
        setUser(userObj)

        // Optional: Fetch full profile to fill in gaps
        apiClient.get('/user/profile').then(res => {
            const fullUser = res.data.data
            setUser(fullUser)
            localStorage.setItem('user', JSON.stringify(fullUser))
        }).catch(console.error)
    }

    const logout = () => {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        setToken(null)
        setUser(null)
        window.location.href = '/login'
    }

    const updateUser = (updatedUser: User) => {
        setUser(updatedUser)
        localStorage.setItem('user', JSON.stringify(updatedUser))
    }

    return (
        <AuthContext.Provider value={{ user, token, isAuthenticated: !!token, isLoading, login, logout, updateUser }}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth() {
    const context = useContext(AuthContext)
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider')
    }
    return context
}
