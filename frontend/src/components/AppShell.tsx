import { Link, Outlet, useRouterState } from '@tanstack/react-router'

import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { signOut } from '@/lib/auth'

const navItems = [
  { to: '/verify', label: 'Verify' },
  { to: '/history', label: 'History' },
] as const

export function AppShell() {
  const pathname = useRouterState({ select: (state) => state.location.pathname })

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-4">
          <div>
            <p className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
              TTB Label Review
            </p>
            <h1 className="text-lg font-semibold text-foreground">Label Verification</h1>
          </div>
          <nav className="flex items-center gap-1">
            {navItems.map((item) => (
              <Button
                key={item.to}
                asChild
                size="sm"
                variant={pathname.startsWith(item.to) ? 'secondary' : 'ghost'}
              >
                <Link to={item.to}>{item.label}</Link>
              </Button>
            ))}
            <Separator orientation="vertical" className="mx-2 h-6" />
            <Button
              size="sm"
              variant="outline"
              onClick={() => {
                void signOut().then(() => {
                  window.location.href = '/login'
                })
              }}
            >
              Sign out
            </Button>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
