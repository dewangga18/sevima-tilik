import { useEffect, useState } from 'react'
import type { User } from './types'
import { api, getAuthToken, setAuthToken } from './services/api'
import { Navbar } from './components/Navbar'
import { LoginView } from './components/LoginView'
import { StudentDashboard } from './components/StudentDashboard'
import { ManagementDashboard } from './components/ManagementDashboard'
import './App.css'

export function App() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(() => Boolean(getAuthToken()))
  const [loginLoading, setLoginLoading] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')
  const [studentNavigation, setStudentNavigation] = useState<{ target: string; request: number } | null>(null)

  useEffect(() => {
    if (!getAuthToken()) return
    let cancelled = false
    api.getMe()
      .then(data => { if (!cancelled) setUser(data) })
      .catch((err: unknown) => {
        if (cancelled) return
        if ((err as { status?: number }).status === 401) {
          setAuthToken('')
          setError('Sesi telah berakhir. Silakan masuk kembali.')
        } else {
          setError('Sesi belum bisa dimuat. Periksa koneksi atau masuk kembali.')
        }
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  async function handleLogin(email: string, password: string) {
    setLoginLoading(true)
    setNotice('')
    setError('')
    try {
      const data = await api.login(email, password)
      setUser(data.user)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Login gagal. Periksa email dan password.')
    } finally {
      setLoginLoading(false)
    }
  }

  async function handleDemoLogin(role: User['role']) {
    setLoginLoading(true)
    setNotice('')
    setError('')
    try {
      const data = await api.demoLogin(role)
      setUser(data.user)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal masuk ke akun demo.')
    } finally {
      setLoginLoading(false)
    }
  }

  async function handleLogout() {
    if (loggingOut) return
    setLoggingOut(true)
    setError('')
    setNotice('')
    try {
      await api.logout()
      setNotice('Kamu sudah keluar dari Tilik.')
    } catch {
      setError('Keluar dari perangkat ini berhasil, tetapi sesi server belum bisa dicabut. Periksa koneksi lalu coba lagi.')
    } finally {
      setUser(null)
      setLoggingOut(false)
    }
  }

  if (loading) return <div className="center-screen" role="status"><div className="spinner" aria-hidden="true" /><p className="loading-text">Memuat Tilik...</p></div>

  return <div className="app-layout">
    <a className="skip-link" href="#main-content">Lewati ke konten utama</a>
    {user && user.role !== 'student' ? <ManagementDashboard key={user.id} user={user} onLogout={handleLogout} loggingOut={loggingOut} /> : <>
    <Navbar user={user} onLogout={handleLogout} loggingOut={loggingOut} onStudentNavigate={target => setStudentNavigation(previous => ({ target, request: (previous?.request || 0) + 1 }))} />
    <main id="main-content" className={`main-content ${user?.role === 'student' ? 'main-content-student' : ''}`}>
      {error && user && <div className="alert alert-error main-alert" role="alert"><span>{error}</span><button type="button" className="alert-close" aria-label="Tutup pesan" onClick={() => setError('')}>Tutup</button></div>}
      {!user ? <div className="auth-wrapper">{notice && <p className="alert auth-notice" role="status">{notice}</p>}<LoginView onLogin={handleLogin} onDemoLogin={handleDemoLogin} loading={loginLoading} error={error} /></div>
        : user.role === 'student' ? <StudentDashboard key={user.id} user={user} navigation={studentNavigation} /> : null}
    </main>
    </>}
  </div>
}

export default App
