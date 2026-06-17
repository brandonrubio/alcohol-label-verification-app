import { createAuthClient } from '@neondatabase/neon-js/auth'

import { setAuthToken } from '@/lib/api'

const authURL = import.meta.env.VITE_NEON_AUTH_URL
const demoMode = import.meta.env.VITE_DEMO_MODE === 'true'

export const authClient = authURL ? createAuthClient(authURL) : null

export async function initializeAuth() {
  if (demoMode) {
    setAuthToken('')
    return { id: 'dev-user', email: 'dev@example.com', name: 'Demo Agent' }
  }

  if (!authClient) {
    throw new Error('VITE_NEON_AUTH_URL is not configured')
  }

  const session = await authClient.getSession()
  const token = session.data?.session?.token
  if (!token) {
    setAuthToken(null)
    return null
  }

  setAuthToken(token)
  const user = session.data?.user
  return user
    ? { id: user.id, email: user.email ?? undefined, name: user.name ?? undefined }
    : null
}

export async function signIn(email: string, password: string) {
  if (!authClient) {
    throw new Error('Auth is not configured')
  }

  const result = await authClient.signIn.email({ email, password })
  if (result.error) {
    throw new Error(result.error.message ?? 'Sign in failed')
  }

  return initializeAuth()
}

export async function signOut() {
  if (authClient) {
    await authClient.signOut()
  }
  setAuthToken(null)
}
