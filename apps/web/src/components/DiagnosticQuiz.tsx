import React, { useState } from 'react'
import type { Assessment, AssessmentItem } from '../types'

interface DiagnosticQuizProps {
  assessment: Assessment
  onSubmit: (answers: { question_id: string; student_answer: string }[]) => Promise<void>
  submitting: boolean
}

export const DiagnosticQuiz: React.FC<DiagnosticQuizProps> = ({
  assessment,
  onSubmit,
  submitting,
}) => {
  const items = assessment.items || []
  const [currentIndex, setCurrentIndex] = useState(0)
  const [answers, setAnswers] = useState<Record<string, string>>(() => {
    const initial: Record<string, string> = {}
    items.forEach((item) => {
      if (item.student_answer) {
        initial[item.question_id] = item.student_answer
      }
    })
    return initial
  })

  if (items.length === 0) {
    return <div className="card">Tidak ada soal yang tersedia.</div>
  }

  const currentItem: AssessmentItem = items[currentIndex]
  const currentQuestion = currentItem.question

  const handleSelectOption = (option: string) => {
    setAnswers((prev) => ({
      ...prev,
      [currentItem.question_id]: option,
    }))
  }

  const handleNext = () => {
    if (currentIndex < items.length - 1) {
      setCurrentIndex((prev) => prev + 1)
    }
  }

  const handlePrev = () => {
    if (currentIndex > 0) {
      setCurrentIndex((prev) => prev - 1)
    }
  }

  const handleFinish = () => {
    const formatted = items.map((item) => ({
      question_id: item.question_id,
      student_answer: answers[item.question_id] || '',
    }))
    onSubmit(formatted)
  }

  const answeredCount = Object.values(answers).filter(Boolean).length
  const progressPercent = Math.round(((currentIndex + 1) / items.length) * 100)

  return (
    <div className="quiz-container">
      {/* Progress header */}
      <div className="quiz-progress-bar">
        <div className="quiz-progress-fill" style={{ width: `${progressPercent}%` }} />
      </div>

      <div className="quiz-header">
        <div className="quiz-step-info">
          <span className="quiz-badge">Soal {currentIndex + 1} dari {items.length}</span>
          <span className="quiz-skill-label">
            Materi: {currentQuestion?.skill_id.replace('_', ' ').toUpperCase()}
          </span>
        </div>
        <span className="quiz-answered-count">
          Terjawab: {answeredCount}/{items.length}
        </span>
      </div>

      {/* Question Card */}
      <div className="question-card">
        <h3 className="question-prompt">{currentQuestion?.prompt}</h3>

        <div className="options-list">
          {currentQuestion?.options.map((option, idx) => {
            const isSelected = answers[currentItem.question_id] === option
            const letters = ['A', 'B', 'C', 'D']
            return (
              <button
                key={idx}
                type="button"
                className={`option-btn ${isSelected ? 'option-selected' : ''}`}
                onClick={() => handleSelectOption(option)}
                disabled={submitting}
              >
                <span className="option-letter">{letters[idx] || idx + 1}</span>
                <span className="option-text">{option}</span>
                {isSelected && <span className="option-check">✓</span>}
              </button>
            )
          })}
        </div>
      </div>

      {/* Navigation Buttons */}
      <div className="quiz-nav">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={handlePrev}
          disabled={currentIndex === 0 || submitting}
        >
          ← Sebelumnya
        </button>

        <div className="quiz-nav-right">
          {currentIndex < items.length - 1 ? (
            <button
              type="button"
              className="btn btn-primary"
              onClick={handleNext}
              disabled={submitting}
            >
              Selanjutnya →
            </button>
          ) : (
            <button
              type="button"
              className="btn btn-success"
              onClick={handleFinish}
              disabled={submitting}
            >
              {submitting ? 'Memeriksa Jawaban...' : '✓ Selesai & Kirim Jawaban'}
            </button>
          )}
        </div>
      </div>

      {/* Question quick selector pills */}
      <div className="quiz-pills">
        {items.map((item, idx) => {
          const isAnswered = Boolean(answers[item.question_id])
          const isCurrent = idx === currentIndex
          return (
            <button
              key={idx}
              type="button"
              className={`pill-btn ${isCurrent ? 'pill-current' : ''} ${isAnswered ? 'pill-answered' : ''}`}
              onClick={() => setCurrentIndex(idx)}
              disabled={submitting}
            >
              {idx + 1}
            </button>
          )
        })}
      </div>
    </div>
  )
}
