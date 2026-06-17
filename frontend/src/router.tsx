import { Outlet, createRootRoute, createRoute, createRouter, redirect } from '@tanstack/react-router'

import { AppShell } from '@/components/AppShell'
import { BatchUpload } from '@/components/BatchUpload'
import { ResultChecklist } from '@/components/ResultChecklist'
import { StatusBadge } from '@/components/StatusBadge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { VerificationForm } from '@/components/VerificationForm'
import { api } from '@/lib/api'
import { initializeAuth, signIn } from '@/lib/auth'

const rootRoute = createRootRoute({
  component: () => <Outlet />,
})

const appLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'app',
  component: AppShell,
})

const indexRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: '/verify' })
  },
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
})

const verifyRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/verify',
  component: VerifyPage,
})

const batchRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/batch',
  component: BatchPage,
})

const historyRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/history',
  loader: async () => api.listVerifications(),
  component: HistoryPage,
})

const resultRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/results/$id',
  loader: async ({ params }) => api.getVerification(params.id),
  component: ResultPage,
})

const routeTree = rootRoute.addChildren([
  loginRoute,
  appLayoutRoute.addChildren([indexRoute, verifyRoute, batchRoute, historyRoute, resultRoute]),
])

export const router = createRouter({
  routeTree,
  context: {
    auth: null as Awaited<ReturnType<typeof initializeAuth>> | null,
  },
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

function LoginPage() {
  const demoMode = import.meta.env.VITE_DEMO_MODE === 'true'

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Sign in</CardTitle>
        </CardHeader>
        <CardContent>
          {demoMode ? (
            <div className="space-y-4">
              <Alert>
                <AlertDescription>
                  Demo mode is enabled. Continue without Neon Auth for local testing.
                </AlertDescription>
              </Alert>
              <Button
                className="w-full"
                onClick={() => {
                  void initializeAuth().then(() => {
                    window.location.href = '/verify'
                  })
                }}
              >
                Continue in demo mode
              </Button>
            </div>
          ) : (
            <form
              className="space-y-4"
              onSubmit={(event) => {
                event.preventDefault()
                const formData = new FormData(event.currentTarget)
                const email = String(formData.get('email') ?? '')
                const password = String(formData.get('password') ?? '')
                void signIn(email, password).then(() => {
                  window.location.href = '/verify'
                })
              }}
            >
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" name="email" type="email" required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input id="password" name="password" type="password" required />
              </div>
              <Button className="w-full" type="submit">
                Sign in
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function VerifyPage() {
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Verify a single label</h2>
        <p className="text-sm text-muted-foreground">
          Enter the COLA application details, upload the label image, and review the checklist.
        </p>
      </div>
      <VerificationForm />
    </div>
  )
}

function BatchPage() {
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Batch verification</h2>
        <p className="text-sm text-muted-foreground">
          Process multiple labels against the same application data. Failures appear first.
        </p>
      </div>
      <BatchUpload />
    </div>
  )
}

function HistoryPage() {
  const data = historyRoute.useLoaderData()
  const sorted = [...data.results].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  )

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Recent verifications</h2>
        <p className="text-sm text-muted-foreground">
          Open a saved result to review the field checklist again.
        </p>
      </div>
      <Card>
        <CardContent className="divide-y divide-border p-0">
          {sorted.length === 0 ? (
            <p className="p-6 text-sm text-muted-foreground">No verifications yet.</p>
          ) : (
            sorted.map((item) => (
              <div
                key={item.id}
                className="flex items-center justify-between gap-4 px-6 py-4"
              >
                <div>
                  <p className="font-medium">{item.image_name}</p>
                  <p className="text-sm text-muted-foreground">
                    {new Date(item.created_at).toLocaleString()}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <StatusBadge status={item.status} />
                  <Button asChild size="sm" variant="outline">
                    <a href={`/results/${item.id}`}>Open</a>
                  </Button>
                </div>
              </div>
            ))
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function ResultPage() {
  const result = resultRoute.useLoaderData()

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Verification detail</h2>
        <p className="text-sm text-muted-foreground">{result.image_name}</p>
      </div>
      <ResultChecklist result={result} />
    </div>
  )
}
