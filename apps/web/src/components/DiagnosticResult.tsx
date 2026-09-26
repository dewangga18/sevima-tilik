import React from 'react'
import type { Assessment } from '../types'

interface DiagnosticResultProps {
  assessment: Assessment
  onRetake: () => void
  retaking?: boolean
}

export const DiagnosticResult: React.FC<DiagnosticResultProps> = ({
  assessment,
  onRetake,
  retaking = false,
}) => {
  const items = assessment.items || []
  const results = assessment.results || []

  const totalCorrect = items.filter((it) => it.is_correct === true).length
  const totalItems = items.length

  const masteredCount = results.filter((r) => r.status === 'mastered').length
  const needsPracticeCount = results.filter((r) => r.status === 'needs_practice').length
  const unassessedCount = results.filter((r) => r.status === 'unassessed').length

  return (
    <div className="results-container">
      <p className="assessment-success" role="status">Jawaban tersimpan. Hasil cek pemahaman siap dilihat.</p>
      {/* Top Banner */}
      <div className="results-summary-card">
        <div className="results-header">
          <div>
            <h2>Hasil Asesmen Diagnostik Numerasi</h2>
            <p className="results-sub">
              Selesai pada {new Date(assessment.completed_at || assessment.started_at).toLocaleString('id-ID')}
            </p>
          </div>
          <button type="button" className="btn btn-secondary btn-sm" onClick={onRetake} disabled={retaking}>
            {retaking ? 'Menyiapkan aktivitas...' : 'Cek pemahaman lagi'}
          </button>
        </div>

        <div className="summary-stats-grid">
          <div className="stat-card">
            <span className="stat-num">{totalCorrect} / {totalItems}</span>
            <span className="stat-label">Jawaban Benar</span>
          </div>
          <div className="stat-card stat-success">
            <span className="stat-num">{masteredCount}</span>
            <span className="stat-label">Konsep Dikuasai</span>
          </div>
          <div className="stat-card stat-warning">
            <span className="stat-num">{needsPracticeCount}</span>
            <span className="stat-label">Perlu Penguatan</span>
          </div>
          <div className="stat-card stat-muted">
            <span className="stat-num">{unassessedCount}</span>
            <span className="stat-label">Belum Terukur</span>
          </div>
        </div>
      </div>

      {/* Per-Skill Evidence Section */}
      <div className="skills-section">
        <h3>Peta Penguasaan Konsep & Bukti (Per-Skill)</h3>
        <p className="section-desc">
          Evaluasi objektif per konsep materi berdasarkan bukti jawaban siswa:
        </p>

        {results.length === 0 && <p>Belum ada bukti konsep yang tersedia pada hasil ini.</p>}
        <div className="skills-grid">
          {results.map((res) => {
            const isMastered = res.status === 'mastered'
            const isNeedsPractice = res.status === 'needs_practice'

            return (
              <div
                key={res.skill_id}
                className={`skill-card ${
                  isMastered ? 'skill-card-mastered' : isNeedsPractice ? 'skill-card-practice' : 'skill-card-unassessed'
                }`}
              >
                <div className="skill-card-top">
                  <h4 className="skill-title">{res.skill_name}</h4>
                  <span className={`status-badge status-${res.status}`}>
                    {isMastered
                      ? '✓ Dikuasai'
                      : isNeedsPractice
                      ? '⚠️ Perlu Latihan'
                      : '— Belum Dinilai'}
                  </span>
                </div>

                <div className="skill-metrics">
                  <div className="metric-row">
                    <span className="metric-key">Akurasi Soal:</span>
                    <span className="metric-val">
                      {res.total_correct} dari {res.total_answered} benar
                    </span>
                  </div>
                  <div className="metric-row">
                    <span className="metric-key">Jumlah Bukti:</span>
                    <span className="metric-val">{res.evidence_count} titik data</span>
                  </div>
                  <div className="metric-row">
                    <span className="metric-key">Tingkat Keyakinan:</span>
                    <span className={`confidence-tag conf-${res.confidence}`}>
                      {res.confidence === 'high' ? 'Tinggi' : res.confidence === 'medium' ? 'Sedang' : 'Awal'}
                    </span>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Question by question review */}
      <div className="review-section">
        <h3>Pembahasan & Bukti Soal</h3>
        <div className="review-list">
          {items.map((item, idx) => (
            <div
              key={item.id}
              className={`review-item ${item.is_correct ? 'review-correct' : 'review-wrong'}`}
            >
              <div className="review-item-header">
                <span className="review-idx">Soal #{idx + 1}</span>
                <span className={`review-badge ${item.is_correct ? 'badge-correct' : 'badge-wrong'}`}>
                  {item.is_correct ? '✓ Jawaban Benar' : '✗ Perlu Diperbaiki'}
                </span>
              </div>
              <p className="review-prompt">{item.question?.prompt}</p>
              <div className="review-answers">
                <p>
                  <strong>Jawabanmu:</strong>{' '}
                  <span className={item.is_correct ? 'text-success' : 'text-danger'}>
                    {item.student_answer || '(Tidak dijawab)'}
                  </span>
                </p>
                {item.question?.explanation && (
                  <div className="review-explanation">
                    <strong>Penjelasan:</strong> {item.question.explanation}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
