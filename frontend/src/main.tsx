import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { initializeAuth } from '@/lib/auth'
import { router } from '@/router'
import './index.css'

const queryClient = new QueryClient()

async function bootstrap() {
  const user = await initializeAuth().catch(() => null)
  const isLoginRoute = window.location.pathname === '/login'
  const demoMode = import.meta.env.VITE_DEMO_MODE === 'true'

  if (!user && !isLoginRoute && !demoMode) {
    window.location.href = '/login'
    return
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} context={{ auth: user }} />
      </QueryClientProvider>
    </StrictMode>,
  )
}

void bootstrap()
