import type { User, Assessment, AssessmentHistory, Skill } from '../types'

const API_BASE = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : '')

let authToken = localStorage.getItem('tilik_token') || ''

export function setAuthToken(token: string) {
  authToken = token
  if (token) {
    localStorage.setItem('tilik_token', token)
  } else {
    localStorage.removeItem('tilik_token')
  }
}

export function getAuthToken(): string {
  return authToken
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  })

  const json = await res.json().catch(() => ({}))

  if (!res.ok) {
    throw new Error(json.error || `HTTP error ${res.status}`)
  }

  return json.data as T
}

export const api = {
  // Health
  checkHealth: async () => {
    const res = await fetch(`${API_BASE}/health`)
    return res.json()
  },

  // Auth
  login: async (email: string, password: string): Promise<{ user: User; token: string }> => {
    const data = await request<{ user: User; token: string }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    setAuthToken(data.token)
    return data
  },

  logout: async (): Promise<void> => {
    try {
      await request('/api/auth/logout', { method: 'POST' })
    } finally {
      setAuthToken('')
    }
  },

  demoLogin: async (role: User['role']): Promise<{ user: User; token: string }> => {
    const data = await request<{ user: User; token: string }>('/api/auth/demo-login', {
      method: 'POST',
      body: JSON.stringify({ role }),
    })
    setAuthToken(data.token)
    return data
  },

  getMe: async (): Promise<User> => {
    return request<User>('/api/auth/me')
  },

  // Curriculum
  getSkills: async (): Promise<Skill[]> => {
    return request<Skill[]>('/api/skills')
  },

  // Diagnostic
  startDiagnostic: async (gradeLevel = 4): Promise<Assessment> => {
    return request<Assessment>('/api/diagnostic/start', {
      method: 'POST',
      body: JSON.stringify({ grade_level: gradeLevel }),
    })
  },

  submitDiagnostic: async (
    assessmentId: string,
    answers: { question_id: string; student_answer: string }[]
  ): Promise<Assessment> => {
    return request<Assessment>('/api/diagnostic/submit', {
      method: 'POST',
      body: JSON.stringify({
        assessment_id: assessmentId,
        answers,
      }),
    })
  },

  getLatestDiagnostic: async (): Promise<Assessment | null> => {
    const assessment = await request<Assessment | null>('/api/diagnostic/latest')
    return assessment ?? null
  },

  getDiagnosticHistory: async (): Promise<AssessmentHistory> => {
    return request<AssessmentHistory>('/api/diagnostic/history')
  },

  getDiagnostic: async (id: string): Promise<Assessment> => {
    return request<Assessment>(`/api/diagnostic/${encodeURIComponent(id)}`)
  },
}
