import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import type { User } from '@/types'
import { logout as apiLogout } from '@/api/client'

interface AuthContextValue {
  user: User | null
  signIn: (user: User) => void
  signOut: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  // Only the user display object is kept in localStorage — never the JWT.
  // The JWT lives exclusively in the httpOnly session cookie set by the backend.
  const [user, setUser] = useState<User | null>(() => {
    try { return JSON.parse(localStorage.getItem('user') ?? 'null') } catch { return null }
  })

  const signIn = useCallback((u: User) => {
    localStorage.setItem('user', JSON.stringify(u))
    setUser(u)
  }, [])

  const signOut = useCallback(async () => {
    try { await apiLogout() } catch { /* best-effort — clear client state regardless */ }
    localStorage.removeItem('user')
    setUser(null)
  }, [])

  return <AuthContext.Provider value={{ user, signIn, signOut }}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
