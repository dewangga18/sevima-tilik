import { useEffect, useRef, useState } from 'react'
import type { Assessment, AssessmentHistory, User, LearningProgress, LearningSession } from '../types'
import { api } from '../services/api'
import { StudentHome } from './StudentHome'
import { ProgressiveQuiz } from './ProgressiveQuiz'
import { DiagnosticQuiz } from './DiagnosticQuiz'
import { LearningActivity } from './LearningActivity'
import { DiagnosticResult } from './DiagnosticResult'

export function StudentDashboard({ user, navigation }: { user: User; navigation: { target: string; request: number } | null }) {
  const [latest, setLatest] = useState<Assessment | null>(null)
  const [history, setHistory] = useState<AssessmentHistory | null>(null)
  const [current, setCurrent] = useState<Assessment | null>(null)
  const [screen, setScreen] = useState<'home' | 'quiz' | 'result' | 'learning'>('home')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [dataError, setDataError] = useState('')
  const [actionError, setActionError] = useState('')
  const [progress, setProgress] = useState<LearningProgress | null>(null)
  const [learningSession, setLearningSession] = useState<LearningSession | null>(null)
  const [progressLoading, setProgressLoading] = useState(true)
  const [progressError, setProgressError] = useState('')
  const startRequest = useRef<{ source: string; skill: string; id: string } | null>(null)
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
    api.getLearningProgress()
      .then(data => { if (!cancelled) setProgress(data) })
      .catch(() => { if (!cancelled) setProgressError('Progress konsep belum bisa dimuat. Coba lagi.') })
      .finally(() => { if (!cancelled) setProgressLoading(false) })
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

  async function refreshLearningProgress() {
    setProgressLoading(true)
    setProgressError('')
    try { setProgress(await api.getLearningProgress()) }
    catch { setProgressError('Progress konsep belum bisa dimuat. Coba lagi.') }
    finally { setProgressLoading(false) }
  }

  async function refreshDashboard() {
    void refreshLearningProgress()
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
    if (!current || submitting) return
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

  async function answerDiagnostic(questionId: string, answer: string) {
    if (!current || submitting) return
    setSubmitting(true)
    setActionError('')
    try {
      const updated = await api.answerDiagnostic(current.id, questionId, answer)
      setCurrent(updated)
      setLatest(updated)
      if (updated.status === 'completed') {
        setScreen(previous => previous === 'home' ? 'home' : 'result')
        await refreshDashboard()
      }
    } catch {
      setActionError('Jawaban belum bisa disimpan. Coba kirim lagi dengan pilihan yang sama. Jika sudah tersimpan, buka kembali aktivitas dari beranda.')
    } finally { setSubmitting(false) }
  }

  async function startLearning(source: string, skill: string) {
    if (busy || submitting) return
    setBusy(true)
    setActionError('')
    try {
      if (!startRequest.current || startRequest.current.source !== source || startRequest.current.skill !== skill) {
        startRequest.current = { source, skill, id: Array.from(crypto.getRandomValues(new Uint8Array(16)), value => value.toString(16).padStart(2, '0')).join('') }
      }
      setLearningSession(await api.startLearning(source, skill, startRequest.current.id))
      startRequest.current = null
      setScreen('learning')
      void refreshLearningProgress()
    } catch (error) {
      setActionError(error instanceof Error ? error.message : 'Materi belum bisa dibuka. Coba lagi.')
    } finally { setBusy(false) }
  }

  async function openLearningSession(id: string) {
    if (busy || submitting) return
    setBusy(true)
    setActionError('')
    try { setLearningSession(await api.getLearningSession(id)); setScreen('learning') }
    catch { setActionError('Aktivitas belajar belum bisa dibuka. Coba lagi.') }
    finally { setBusy(false) }
  }

  async function saveLearningAction(questionId?: string, answer?: string) {
    if (!learningSession || submitting) return
    setSubmitting(true)
    setActionError('')
    try {
      const updated = questionId && answer
        ? await api.answerLearning(learningSession.id, questionId, answer)
        : await api.completeLesson(learningSession.id)
      setLearningSession(updated)
      await refreshLearningProgress()
    } catch (error) {
      setActionError(`${error instanceof Error ? error.message : 'Aktivitas belum bisa disimpan.'} Coba ulangi tindakan yang sama, atau buka kembali dari beranda.`)
    } finally { setSubmitting(false) }
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
    {screen === 'home' && <StudentHome user={user} latest={latest} history={history} loading={loading} busy={busy} error={dataError} onStart={startDiagnostic} onOpen={openAssessment} onRetry={refreshDashboard} progress={progress} progressLoading={progressLoading} progressError={progressError} onProgressRetry={refreshLearningProgress} onLearn={startLearning} onOpenLearning={openLearningSession} />}
    {screen !== 'home' && <div className="activity-toolbar">
      <button type="button" className="btn btn-secondary" onClick={goHome} disabled={submitting || busy}>Kembali ke beranda</button>
      <h1 ref={screenTitle} tabIndex={-1}>{screen === 'quiz' ? 'Cek pemahamanmu' : screen === 'learning' ? learningSession?.skill_name || 'Belajar konsep' : 'Hasil cek pemahaman'}</h1>
      {busy && <p role="status">Menyiapkan aktivitas...</p>}
    </div>}
    {current?.status === 'in_progress' && <div hidden={screen !== 'quiz'}>
      {current.rule_version === 'progressive-demo-v1'
        ? <ProgressiveQuiz key={current.items?.find(item => !item.answered_at)?.question_id || current.id} assessment={current} onAnswer={answerDiagnostic} submitting={submitting} />
        : <DiagnosticQuiz key={current.id} assessment={current} onSubmit={submitDiagnostic} submitting={submitting} />}
    </div>}
    {screen === 'result' && current?.status === 'completed' && <DiagnosticResult assessment={current} onRetake={startDiagnostic} retaking={busy} onLearn={startLearning} learning={busy} />}
    {learningSession && <div hidden={screen !== 'learning'}><LearningActivity key={learningSession.items.find(item => !item.answered_at)?.question_id || `${learningSession.id}-${learningSession.stage}`} session={learningSession} busy={busy || submitting} onCompleteLesson={() => saveLearningAction()} onAnswer={saveLearningAction} onHome={goHome} onReview={skill => startLearning(learningSession.source_assessment_id, skill)} /></div>}
  </div>
}
