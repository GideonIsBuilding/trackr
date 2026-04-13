import type {
  Application, ApplicationChecklist, ApplicationSource,
  ApplicationStatus, AnalyticsSummary, StatusHistory, User
} from '@/types'

const BASE = '/api'

function getToken() { return localStorage.getItem('token') }

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string>),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, { ...init, headers })

  if (res.status === 401) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (res.status === 204) return undefined as T

  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? 'Request failed')
  return data as T
}

// --- Auth ---
export async function register(email: string, password: string, timezone: string) {
  return request<{ user: User; token: string }>('/auth/register', {
    method: 'POST', body: JSON.stringify({ email, password, timezone }),
  })
}
export async function login(email: string, password: string) {
  return request<{ user: User; token: string }>('/auth/login', {
    method: 'POST', body: JSON.stringify({ email, password }),
  })
}
export async function forgotPassword(email: string) {
  return request<{ message: string }>('/auth/forgot-password', {
    method: 'POST', body: JSON.stringify({ email }),
  })
}
export async function resetPassword(token: string, password: string) {
  return request<{ message: string }>('/auth/reset-password', {
    method: 'POST', body: JSON.stringify({ token, password }),
  })
}

// --- Applications ---
export async function getApplications(status?: ApplicationStatus) {
  return request<Application[]>(`/applications${status ? `?status=${status}` : ''}`)
}
export async function getApplication(id: string) {
  return request<Application>(`/applications/${id}`)
}
export async function createApplication(data: {
  company: string; role: string; job_url?: string; location?: string
  source: ApplicationSource; notes?: string; applied_at?: string
  cover_letter?: boolean; cv_tailored?: boolean; referral?: boolean
  portfolio_link?: string; video_intro?: boolean; linkedin_connect?: boolean
}) {
  return request<Application>('/applications', { method: 'POST', body: JSON.stringify(data) })
}
export async function updateStatus(id: string, status: ApplicationStatus, note?: string) {
  return request<StatusHistory>(`/applications/${id}/status`, {
    method: 'PATCH', body: JSON.stringify({ status, note }),
  })
}
export async function getStatusHistory(id: string) {
  return request<StatusHistory[]>(`/applications/${id}/history`)
}
export async function deleteApplication(id: string) {
  return request<void>(`/applications/${id}`, { method: 'DELETE' })
}
export async function updateChecklist(id: string, checklist: ApplicationChecklist) {
  return request<Application>(`/applications/${id}/checklist`, {
    method: 'PATCH', body: JSON.stringify(checklist),
  })
}

// --- Analytics ---
export async function getAnalytics() {
  return request<AnalyticsSummary>('/analytics')
}

// --- Reminders ---
export async function configureReminder(id: string, trigger_after_days: number) {
  return request(`/applications/${id}/reminder`, {
    method: 'PUT', body: JSON.stringify({ trigger_after_days }),
  })
}
export async function snoozeReminder(id: string, days: number) {
  return request(`/applications/${id}/reminder/snooze`, {
    method: 'POST', body: JSON.stringify({ days }),
  })
}
