import { useEffect, useRef, useState } from 'react'
import type { LearningSession } from '../types'

const outcomeLabels = { strong_evidence: 'Bukti kuat', inconclusive: 'Belum pasti', needs_practice: 'Perlu penguatan', mastered: 'Dikuasai', unassessed: 'Belum dinilai' }

export function LearningActivity({ session, busy, onCompleteLesson, onAnswer, onHome, onReview }: {
  session: LearningSession
  busy: boolean
  onCompleteLesson: () => void
  onAnswer: (questionId: string, answer: string) => void
  onHome: () => void
  onReview: (skillId: string) => void
}) {
  const [answer, setAnswer] = useState('')
  const heading = useRef<HTMLHeadingElement>(null)
  const pending = session.items.find(item => !item.answered_at)
  useEffect(() => { heading.current?.focus() }, [session.stage, pending?.question_id])

  if (session.stage === 'lesson') return <section className="learning-lesson" aria-busy={busy}>
    <h2 ref={heading} tabIndex={-1}>{session.lesson?.title || 'Materi belum tersedia'}</h2>
    <p className="learning-meta">Sekitar {session.lesson?.estimated_minutes || 2} menit · Satu konsep, satu langkah.</p>
    {session.lesson?.steps.length ? <ol className="lesson-steps">{session.lesson.steps.map(step => <li key={step.heading}><h3>{step.heading}</h3><p>{step.body}</p><p className="lesson-example">{step.example}</p></li>)}</ol> : <p>Materi belum bisa ditampilkan. Kembali ke beranda lalu coba lagi.</p>}
    <button type="button" className="btn btn-primary" disabled={busy || !session.lesson?.steps.length} onClick={onCompleteLesson}>{busy ? 'Menyimpan...' : 'Selesai membaca, mulai latihan'}</button>
    <p className="learning-meta">Selesai membaca tidak otomatis menaikkan score pemahaman.</p>
  </section>

  if (session.stage === 'completed' || session.stage === 'exhausted') return <section className="learning-result">
    <h2 ref={heading} tabIndex={-1}>{session.stage === 'completed' ? 'Lihat langkah yang sudah kamu tempuh' : 'Soal baru belum cukup'}</h2>
    {session.stage === 'completed' ? <>
      <p role="status">Reassessment tersimpan. Progress belajarmu sudah diperbarui.</p>
      <dl className="learning-score-comparison"><div><dt>Bukti sebelumnya</dt><dd>{session.before_score === undefined ? 'Belum diukur' : `${session.before_score}%`}</dd></div><div><dt>Reassessment terakhir</dt><dd>{session.score}%</dd></div></dl>
      <p><strong>{session.outcome ? outcomeLabels[session.outcome] : 'Belum pasti'}</strong> · Dari tiga jawaban reassessment.</p>
      <p>{session.outcome === 'strong_evidence' ? 'Ada bukti yang lebih kuat pada latihan konsep ini. Kamu bisa melihat rekomendasi berikutnya di beranda.' : 'Konsep ini masih perlu ditinjau. Itu bagian dari belajar; pilih review atau prasyarat yang disarankan di beranda.'}</p>
      <p className="learning-meta">Persentase ini memakai aturan demo awal, bukan nilai ujian atau mastery akademis tervalidasi.</p>
      <details className="learning-review"><summary>Lihat pembahasan reassessment</summary><ol>{session.items.filter(item => item.stage === 'reassessment').map(item => <li key={item.id}><h3>{item.question?.prompt}</h3><p>Jawabanmu: {item.student_answer} · {item.is_correct ? 'Benar' : 'Perlu ditinjau'}</p><p>{item.question?.explanation}</p></li>)}</ol></details>
    </> : <p>Progress tidak dinaikkan karena bukti belum cukup. Pilih rekomendasi lain atau lanjut setelah soal baru tersedia.</p>}
    <button type="button" className="btn btn-primary" disabled={busy} onClick={onHome}>Lihat progress di beranda</button>
  </section>

  const answered = session.items.filter(item => item.stage === session.stage && item.answered_at).length
  const lastPractice = [...session.items].reverse().find(item => item.stage === 'practice' && item.answered_at)
  if (!pending?.question) return <p role="status">Soal belum tersedia. Kembali ke beranda lalu buka kembali aktivitas ini.</p>
  return <div className="quiz-container progressive-quiz learning-quiz" aria-busy={busy}>
    <div className="quiz-header"><span className="quiz-badge">{session.stage === 'practice' ? 'Latihan' : 'Reassessment'} {answered + 1} dari 3</span><span className="quiz-answered-count">Level soal {pending.question.difficulty}</span></div>
    {session.stage === 'practice' && lastPractice && <div className="practice-feedback" role="status"><strong>{lastPractice.is_correct ? 'Jawaban sebelumnya benar.' : 'Mari tinjau jawaban sebelumnya.'}</strong><p>{lastPractice.question?.explanation}</p></div>}
    {session.stage === 'reassessment' && <p className="learning-meta">Soal baru untuk melihat pemahamanmu setelah latihan. Pembahasan tersedia setelah selesai.</p>}
    {session.review_skill_id && session.stage === 'practice' && <div className="learning-prerequisite-review"><p>Kamu bisa meninjau prasyarat terlebih dahulu. Sesi ini tetap tersimpan.</p><button type="button" className="btn btn-secondary" disabled={busy} onClick={() => onReview(session.review_skill_id!)}>Tinjau materi prasyarat</button></div>}
    <div className="question-card"><h2 className="question-prompt" ref={heading} tabIndex={-1}>{pending.question.prompt}</h2><div className="options-list" role="group" aria-label="Pilih jawaban">{pending.question.options.map((option, index) => <button type="button" key={option} className={`option-btn ${answer === option ? 'option-selected' : ''}`} disabled={busy} aria-pressed={answer === option} onClick={() => setAnswer(option)}><span className="option-letter" aria-hidden="true">{String.fromCharCode(65 + index)}</span><span className="option-text">{option}</span>{answer === option && <span className="option-selection-label">Dipilih</span>}</button>)}</div></div>
    {session.stage === 'practice' && session.lesson?.hint && <details className="learning-hint"><summary>Lihat petunjuk konsep</summary><p>{session.lesson.hint}</p></details>}
    <button type="button" className="btn btn-primary progressive-answer-button" disabled={busy || !answer} onClick={() => onAnswer(pending.question_id, answer)}>{busy ? 'Menyimpan jawaban...' : 'Kirim jawaban'}</button>
  </div>
}
