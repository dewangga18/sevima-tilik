import { learningStageLabels, type LearningProgress } from '../types'


const labels = { strong_evidence: 'Bukti kuat', inconclusive: 'Belum pasti', needs_practice: 'Perlu penguatan', unassessed: 'Belum dinilai', mastered: 'Dikuasai' }

export function LearningOverview({ progress, loading, error, busy, onRetry, onLearn, onOpen }: {
  progress: LearningProgress | null
  loading: boolean
  error: string
  busy: boolean
  onRetry: () => void
  onLearn: (sourceId: string, skillId: string) => void
  onOpen: (id: string) => void
}) {
  return <section className="learning-overview" aria-labelledby="learning-overview-title" aria-busy={loading}>
    <div className="section-heading"><h2 id="learning-overview-title">Langkah belajar & progress konsep</h2><p>Bukti terbaru dari diagnostic dan reassessment. Persentase memakai aturan demo awal.</p></div>
    {loading ? <p role="status">Memuat progress konsep...</p> : error ? <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={onRetry} disabled={busy}>Muat ulang progress</button></div> : !progress?.source_assessment_id ? <div className="history-empty"><h3>Kenali pijakanmu dulu.</h3><p>Selesaikan cek pemahaman untuk mendapatkan rekomendasi materi belajar.</p></div> : <>
      {progress.active_sessions.length > 0 && <div className="learning-active-list"><h3>Lanjutkan sesi belajar</h3><ul className="history-list">{progress.active_sessions.map(session => <li className="history-row" key={session.id}><div className="history-activity"><h4>{session.skill_name}</h4><p>{learningStageLabels[session.stage]}</p></div><button type="button" className="btn btn-secondary" onClick={() => onOpen(session.id)} disabled={busy}>Lanjutkan belajar</button></li>)}</ul></div>}
      <div className="learning-current-path"><h3>Urutan review saat ini</h3>{progress.learning_path.length ? <ol>{progress.learning_path.map(item => <li key={item.skill_id}><div><h4>{item.skill_name}</h4><p>{item.reason}</p></div><button type="button" className="btn btn-secondary" disabled={busy} onClick={() => onLearn(progress.source_assessment_id!, item.skill_id)}>Belajar konsep</button></li>)}</ol> : <p>Tidak ada review yang teridentifikasi dari bukti terakhir. Konsep yang belum diuji tetap belum dinilai.</p>}</div>
      <details className="learning-progress-map"><summary>Lihat peta progress konsep</summary><dl>{progress.skills.map(skill => <div key={skill.skill_id}><dt>{skill.skill_name}</dt><dd><strong>{skill.score === undefined ? 'Belum diukur' : `${skill.score}%`}</strong><span>{labels[skill.status]}</span><small>{skill.evidence_count ? `${skill.correct_count}/${skill.evidence_count} benar · ${skill.source === 'reassessment' ? 'Reassessment' : 'Diagnostic'}` : 'Belum ada bukti jawaban'}</small></dd></div>)}</dl></details>
    </>}
  </section>
}
