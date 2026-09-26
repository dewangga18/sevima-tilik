import { useEffect, useRef, useState } from 'react'
import type { Assessment, AssessmentHistory, User } from '../types'
import { api } from '../services/api'
import { StudentHome } from './StudentHome'
import { DiagnosticQuiz } from './DiagnosticQuiz'
import { DiagnosticResult } from './DiagnosticResult'

export function StudentDashboard({ user, navigation }: { user: User; navigation: { target: string; request: number } | null }) {
  const [latest, setLatest] = useState<Assessment | null>(null)
  const [history, setHistory] = useState<AssessmentHistory | null>(null)
  const [current, setCurrent] = useState<Assessment | null>(null)
  const [screen, setScreen] = useState<'home' | 'quiz' | 'result'>('home')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [dataError, setDataError] = useState('')
  const [actionError, setActionError] = useState('')
  const screenTitle = useRef<HTMLHeadingElement>(null)
  const [handledNavigation, setHandledNavigation] = useState(navigation)
  if (navigation !== handledNavigation) {
    setHandledNavigation(navigation)
    setScreen('home')
    setActionError('')
  }

  useEffect(() => {
    let cancelled = false
    Promise.all([api.getLatestDiagnostic(), api.getDiagnosticHistory()])
      .then(([assessment, activities]) => {
        if (cancelled) return
        setLatest(assessment)
        setHistory(activities)
      })
      .catch(() => {
        if (!cancelled) setDataError('Data belajarmu belum bisa dimuat. Periksa koneksi lalu coba lagi.')
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [user.id])

  useEffect(() => {
    if (screen !== 'home') screenTitle.current?.focus()
  }, [screen])

  useEffect(() => {
    if (!navigation) return
    const frame = requestAnimationFrame(() => {
      const target = document.getElementById(navigation.target)
      target?.focus()
      target?.scrollIntoView({ block: 'start' })
    })
    return () => cancelAnimationFrame(frame)
  }, [navigation])

  async function refreshDashboard() {
    setLoading(true)
    setDataError('')
    try {
      const [assessment, activities] = await Promise.all([api.getLatestDiagnostic(), api.getDiagnosticHistory()])
      setLatest(assessment)
      setHistory(activities)
    } catch {
      setDataError('Data belajarmu belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setLoading(false)
    }
  }

  async function startDiagnostic() {
    setBusy(true)
    setActionError('')
    try {
      const assessment = await api.startDiagnostic(user.grade_level || 4)
      setLatest(assessment)
      setCurrent(assessment)
      setScreen('quiz')
    } catch {
      setActionError('Aktivitas belum bisa dimulai. Periksa koneksi lalu coba lagi.')
    } finally {
      setBusy(false)
    }
  }

  async function openAssessment(id: string) {
    setBusy(true)
    setActionError('')
    try {
      const assessment = await api.getDiagnostic(id)
      setCurrent(assessment)
      setScreen(assessment.status === 'completed' ? 'result' : 'quiz')
    } catch {
      setActionError('Aktivitas belum bisa dibuka. Periksa koneksi lalu coba lagi.')
    } finally {
      setBusy(false)
    }
  }

  async function submitDiagnostic(answers: { question_id: string; student_answer: string }[]) {
    if (!current) return
    setSubmitting(true)
    setActionError('')
    try {
      const completed = await api.submitDiagnostic(current.id, answers)
      setCurrent(completed)
      setScreen('result')
      await refreshDashboard()
    } catch {
      setActionError('Jawaban belum bisa dikirim. Periksa koneksi lalu coba lagi. Jika sudah terkirim, buka hasil lewat beranda.')
    } finally {
      setSubmitting(false)
    }
  }

  function goHome() {
    setScreen('home')
    setActionError('')
    void refreshDashboard()
    window.scrollTo({ top: 0 })
  }

  return <div className="student-dashboard">
    {(busy || submitting) && <p className="activity-status" role="status">{submitting ? 'Menyimpan jawaban...' : 'Menyiapkan aktivitas...'}</p>}
    {actionError && <div className="alert alert-error" role="alert">{actionError}</div>}
    {screen === 'home' && <StudentHome user={user} latest={latest} history={history} loading={loading} busy={busy} error={dataError} onStart={startDiagnostic} onOpen={openAssessment} onRetry={refreshDashboard} />}
    {screen !== 'home' && <div className="activity-toolbar">
      <button type="button" className="btn btn-secondary" onClick={goHome} disabled={submitting || busy}>Kembali ke beranda</button>
      <h1 ref={screenTitle} tabIndex={-1}>{screen === 'quiz' ? 'Cek pemahamanmu' : 'Hasil cek pemahaman'}</h1>
      {busy && <p role="status">Menyiapkan aktivitas...</p>}
    </div>}
    {current?.status === 'in_progress' && <div hidden={screen !== 'quiz'}>
      <DiagnosticQuiz key={current.id} assessment={current} onSubmit={submitDiagnostic} submitting={submitting} />
    </div>}
    {screen === 'result' && current?.status === 'completed' && <DiagnosticResult assessment={current} onRetake={startDiagnostic} retaking={busy} />}
  </div>
}
