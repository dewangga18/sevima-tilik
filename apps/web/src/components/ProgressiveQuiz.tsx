import { useEffect, useRef, useState } from 'react'
import type { Assessment } from '../types'

export function ProgressiveQuiz({ assessment, onAnswer, submitting }: {
  assessment: Assessment
  onAnswer: (questionId: string, answer: string) => Promise<void>
  submitting: boolean
}) {
  const [answer, setAnswer] = useState('')
  const prompt = useRef<HTMLHeadingElement>(null)
  const item = assessment.items?.find(question => !question.answered_at)
  const answered = assessment.items?.filter(question => question.answered_at).length || 0
  useEffect(() => { prompt.current?.focus() }, [item?.question_id])
  if (!item?.question) return <div className="card" role="status"><p>Soal belum tersedia. Kembali ke beranda lalu buka aktivitas lagi.</p></div>

  return <div className="quiz-container progressive-quiz" aria-busy={submitting}>
    <div className="quiz-header"><span className="quiz-badge">Soal {answered + 1}</span><span className="quiz-answered-count">{answered} jawaban tersimpan · maksimal {assessment.max_questions || 18} soal</span></div>
    <p className="progressive-quiz-note">Soal berikutnya menyesuaikan jawabanmu. Tidak ada batas waktu.</p>
    <div className="question-card">
      <h2 className="question-prompt" ref={prompt} tabIndex={-1}>{item.question.prompt}</h2>
      <div className="options-list" role="group" aria-label="Pilih jawaban">
        {item.question.options.map((option, index) => <button type="button" className={`option-btn ${answer === option ? 'option-selected' : ''}`} key={option} aria-pressed={answer === option} disabled={submitting} onClick={() => setAnswer(option)}><span className="option-letter" aria-hidden="true">{String.fromCharCode(65 + index)}</span><span className="option-text">{option}</span>{answer === option && <span className="option-selection-label">Dipilih</span>}</button>)}
      </div>
    </div>
    <button type="button" className="btn btn-primary progressive-answer-button" disabled={!answer || submitting} onClick={() => onAnswer(item.question_id, answer)}>{submitting ? 'Menyimpan jawaban...' : 'Kirim jawaban'}</button>
    <p className="progressive-quiz-note">Jawaban yang sudah dikirim tersimpan dan tidak dapat diubah. Pembahasan tersedia setelah selesai.</p>
  </div>
}
