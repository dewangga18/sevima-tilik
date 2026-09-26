import React from 'react'
import type { Assessment } from '../types'

interface DiagnosticResultProps {
  assessment: Assessment
  onRetake: () => void
  retaking?: boolean
  onLearn: (sourceId: string, skillId: string) => void
  learning: boolean
}

export const DiagnosticResult: React.FC<DiagnosticResultProps> = ({
  assessment,
  onRetake,
  retaking = false,
  onLearn,
  learning,
}) => {
  const progressive = assessment.rule_version === 'progressive-demo-v1'
  const items = assessment.items || []
  const results = assessment.results || []

  const totalCorrect = items.filter((it) => it.is_correct === true).length
  const totalItems = items.length

  const masteredCount = results.filter((r) => r.status === 'mastered' || r.status === 'strong_evidence').length
  const needsPracticeCount = results.filter((r) => r.status === 'needs_practice').length
  const unassessedCount = results.filter((r) => r.status === 'unassessed').length

  return (
    <div className={`results-container ${progressive ? 'progressive-results' : ''}`}>
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
            <span className="stat-label">{progressive ? 'Bukti kuat' : 'Konsep dikuasai'}</span>
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

      {progressive && <section className="diagnostic-insight" aria-labelledby="diagnostic-insight-title">
        <h3 id="diagnostic-insight-title">Pijakan belajar berikutnya</h3>
        <p>Materi yang diperiksa: <strong>{results.find(result => result.skill_id === assessment.target_skill_id)?.skill_name}</strong>.</p>
        <p>{assessment.stop_reason === 'question_limit' ? 'Batas soal tercapai. Sebagian konsep mungkin masih memerlukan bukti tambahan.' : assessment.stop_reason === 'insufficient_evidence' ? 'Bukti belum cukup untuk menyimpulkan semua konsep yang diperiksa.' : 'Pemeriksaan pada jalur ini selesai berdasarkan jawabanmu.'}</p>
        <div className="root-gap-insight"><h4>Kandidat gap prasyarat</h4>{results.some(result => result.is_root_gap)
          ? <ul>{results.filter(result => result.is_root_gap).map(result => <li key={result.skill_id}><strong>{result.skill_name}</strong> · {result.total_correct}/{result.total_answered} jawaban benar. Prasyarat diperiksa langsung; ini kandidat untuk ditinjau.</li>)}</ul>
          : <p>Belum ada kandidat gap prasyarat yang didukung bukti cukup. Kesulitan pada materi utama tidak otomatis berarti fondasinya lemah.</p>}</div>
        <h4>Urutan review yang disarankan</h4>
        {assessment.learning_path?.length ? <ol className="diagnostic-learning-path">{assessment.learning_path.map(item => <li key={item.skill_id}><strong>{item.skill_name}</strong><p>{item.reason}</p></li>)}</ol> : <p>Tidak ada review tambahan yang teridentifikasi dari jalur ini. Konsep yang belum diuji tetap belum dinilai.</p>}
        {assessment.learning_path?.[0] && <button type="button" className="btn btn-primary start-learning-button" disabled={learning || retaking} onClick={() => onLearn(assessment.id, assessment.learning_path![0].skill_id)}>{learning ? 'Menyiapkan materi...' : `Mulai belajar ${assessment.learning_path[0].skill_name}`}</button>}
        <p className="invitation-note">Ini hasil diagnostic tersimpan. Progress terbaru setelah belajar ada di profil belajarmu.</p>
      </section>}

      {/* Per-Skill Evidence Section */}
      <div className="skills-section">
        <h3>Peta Penguasaan Konsep & Bukti (Per-Skill)</h3>
        <p className="section-desc">
          {progressive ? 'Klasifikasi dari aturan demo awal. Ini belum ukuran mastery akademis yang tervalidasi.' : 'Hasil dari aturan diagnostic awal berdasarkan jawaban siswa.'}
        </p>

        {results.length === 0 && <p>Belum ada bukti konsep yang tersedia pada hasil ini.</p>}
        <div className="skills-grid">
          {results.map((res) => {
            const isMastered = res.status === 'mastered' || res.status === 'strong_evidence'
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
                      ? progressive ? 'Bukti kuat' : 'Dikuasai'
                      : isNeedsPractice
                      ? 'Perlu penguatan'
                      : res.status === 'inconclusive' ? 'Belum pasti' : 'Belum dinilai'}
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
