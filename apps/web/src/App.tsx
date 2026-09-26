import { useEffect, useState } from 'react'
import type { User, Assessment } from './types'
import { api, getAuthToken } from './services/api'
import { Navbar } from './components/Navbar'
import { LoginView } from './components/LoginView'
import { DiagnosticQuiz } from './components/DiagnosticQuiz'
import { DiagnosticResult } from './components/DiagnosticResult'
import { TeacherView } from './components/TeacherView'
import './App.css'

export function App() {
  const [user, setUser] = useState<User | null>(null)
  const [assessment, setAssessment] = useState<Assessment | null>(null)
  const [loading, setLoading] = useState(() => Boolean(getAuthToken()))
  const [loginLoading, setLoginLoading] = useState(false)
  const [quizSubmitting, setQuizSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [assessmentLoadError, setAssessmentLoadError] = useState('')
  const [assessmentLoading, setAssessmentLoading] = useState(false)

  // Check auth session on load
  useEffect(() => {
    const token = getAuthToken()
    if (!token) {
      return
    }

    api.getMe()
      .then((userData) => {
        setUser(userData)
        if (userData.role === 'student') {
          return loadLatestAssessment()
        }
      })
      .catch(() => {
        setUser(null)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  async function loadLatestAssessment() {
    setAssessmentLoading(true)
    setAssessmentLoadError('')
    try {
      const latest = await api.getLatestDiagnostic()
      setAssessment(latest)
    } catch {
      setAssessmentLoadError('Asesmen terakhir belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setAssessmentLoading(false)
    }
  }

  const handleLogin = async (email: string, pass: string) => {
    setLoginLoading(true)
    setError('')
    try {
      const data = await api.login(email, pass)
      setUser(data.user)
      if (data.user.role === 'student') {
        await loadLatestAssessment()
      }
    } catch (err: any) {
      setError(err.message || 'Login gagal. Periksa email dan password.')
    } finally {
      setLoginLoading(false)
    }
  }

  const handleLogout = async () => {
    try {
      await api.logout()
    } finally {
      setUser(null)
      setAssessment(null)
      setAssessmentLoadError('')
    }
  }

  const handleDemoLogin = async (role: User['role']) => {
    setLoginLoading(true)
    setError('')
    try {
      const data = await api.demoLogin(role)
      setAssessment(null)
      setAssessmentLoadError('')
      setUser(data.user)
      if (data.user.role === 'student') {
        await loadLatestAssessment()
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal masuk ke akun demo.')
    } finally {
      setLoginLoading(false)
    }
  }

  const handleStartDiagnostic = async () => {
    setLoading(true)
    setError('')
    try {
      const newAssessment = await api.startDiagnostic(4)
      setAssessment(newAssessment)
    } catch (err: any) {
      setError(err.message || 'Gagal memulai tes diagnostik.')
    } finally {
      setLoading(false)
    }
  }

  const handleSubmitQuiz = async (
    answers: { question_id: string; student_answer: string }[]
  ) => {
    if (!assessment) return
    setQuizSubmitting(true)
    setError('')
    try {
      const completed = await api.submitDiagnostic(assessment.id, answers)
      setAssessment(completed)
    } catch (err: any) {
      setError(err.message || 'Gagal mengirimkan jawaban.')
    } finally {
      setQuizSubmitting(false)
    }
  }

  const handleRetakeDiagnostic = async () => {
    handleStartDiagnostic()
  }

  if (loading) {
    return (
      <div className="center-screen">
        <div className="spinner" />
        <p className="loading-text">Memuat Tilik...</p>
      </div>
    )
  }

  return (
    <div className="app-layout">
      <Navbar user={user} onLogout={handleLogout} />

      <main className="main-content">
        {error && (
          <div className="alert alert-error main-alert">
            <span>{error}</span>
            <button className="alert-close" onClick={() => setError('')}>✕</button>
          </div>
        )}

        {!user ? (
          <div className="auth-wrapper">
            <LoginView onLogin={handleLogin} onDemoLogin={handleDemoLogin} loading={loginLoading} error={error} />
          </div>
        ) : user.role === 'teacher' || user.role === 'admin' ? (
          <TeacherView user={user} />
        ) : (
          /* Student Flow */
          <div className="student-flow">
            {assessmentLoading ? (
              <p role="status">Memuat asesmen terakhir...</p>
            ) : assessmentLoadError ? (
              <div className="onboarding-card">
                <p role="alert">{assessmentLoadError}</p>
                <button type="button" className="btn btn-primary" onClick={loadLatestAssessment}>
                  Coba lagi
                </button>
              </div>
            ) : !assessment ? (
              <div className="onboarding-card">
                <div className="onboarding-badge">Kelas 4 SD</div>
                <h2>Tes Diagnostik Pemahaman Pecahan & Fondasi</h2>
                <p className="onboarding-desc">
                  Tes ini dirancang untuk mendeteksi konsep numerasi apa saja yang sudah kamu kuasai
                  dan prasyarat mana yang perlu diperkuat agar belajar matematika jadi lebih mudah dan menyenangkan.
                </p>

                <div className="onboarding-features">
                  <div className="feature-item">
                    <span className="feature-icon">🎯</span>
                    <div>
                      <strong>6 Soal Esensial</strong>
                      <span>Mencakup perkalian, pembagian, dan representasi pecahan</span>
                    </div>
                  </div>
                  <div className="feature-item">
                    <span className="feature-icon">🔍</span>
                    <div>
                      <strong>Pemetaan Titik Hambat</strong>
                      <span>Mendeteksi apakah kesulitan berakar dari konsep prasyarat</span>
                    </div>
                  </div>
                  <div className="feature-item">
                    <span className="feature-icon">⏱️</span>
                    <div>
                      <strong>Sekitar 5-10 Menit</strong>
                      <span>Kerjakan dengan teliti dan tenang</span>
                    </div>
                  </div>
                </div>

                <button
                  type="button"
                  className="btn btn-primary btn-lg"
                  onClick={handleStartDiagnostic}
                >
                  🚀 Mulai Tes Diagnostik Sekarang
                </button>
              </div>
            ) : assessment.status === 'completed' ? (
              <DiagnosticResult
                assessment={assessment}
                onRetake={handleRetakeDiagnostic}
              />
            ) : (
              <DiagnosticQuiz
                assessment={assessment}
                onSubmit={handleSubmitQuiz}
                submitting={quizSubmitting}
              />
            )}
          </div>
        )}
      </main>
    </div>
  )
}

export default App
