const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
const DEMO_MODE = import.meta.env.VITE_DEMO_MODE === 'true'

let authToken: string | null = DEMO_MODE ? '' : null

export function setAuthToken(token: string | null) {
  authToken = token
}

export function getAuthToken() {
  return authToken
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (authToken !== null) {
    headers.set('Authorization', `Bearer ${authToken}`)
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers,
    })
  } catch (error) {
    if (error instanceof TypeError) {
      throw new Error(
        `Could not reach the API at ${API_BASE_URL}. Make sure the backend is running (e.g. cd backend && source scripts/load-env.sh && go run ./cmd/api).`,
      )
    }
    throw error
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(body.error ?? 'Request failed')
  }

  return response.json() as Promise<T>
}

export const api = {
  me: () => request<import('./types').UserProfile>('/api/v1/me'),
  listVerifications: () =>
    request<{ results: import('./types').VerificationResult[] }>('/api/v1/verifications'),
  getVerification: (id: string) =>
    request<import('./types').VerificationResult>(`/api/v1/verifications/${id}`),
  createVerification: (application: import('./types').ApplicationData, image: File) => {
    const form = new FormData()
    form.append('application', JSON.stringify(application))
    form.append('image', image)
    return request<import('./types').VerificationResult>('/api/v1/verifications', {
      method: 'POST',
      body: form,
    })
  },
}
