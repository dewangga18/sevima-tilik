import { useEffect, useRef, useState } from 'react'
import type { AIGeneratedDraft, AIGenerateInput, Skill } from '../types'
import { api } from '../services/api'
import './AIIntegration.css'

export function AIGenerator({ onGenerated }: { onGenerated?: () => void }) {
  const [skills, setSkills] = useState<Skill[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [skillId, setSkillId] = useState('')
  const [difficulty, setDifficulty] = useState(1)
  const [purpose, setPurpose] = useState<AIGenerateInput['purpose']>('practice')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [draft, setDraft] = useState<AIGeneratedDraft | null>(null)
  const retry = useRef<{ signature: string; id: string } | null>(null)

  async function loadSkills() {
    setLoading(true); setLoadError('')
    try {
      const data = (await api.getSkills()).filter(skill => skill.grade_level <= 4)
      setSkills(data); setSkillId(data[0]?.id || '')
    } catch { setLoadError('Kurikulum belum bisa dimuat. Periksa koneksi lalu coba lagi.') }
    finally { setLoading(false) }
  }
  useEffect(() => {
    let cancelled = false
    api.getSkills().then(data => {
      if (!cancelled) { const available = data.filter(skill => skill.grade_level <= 4); setSkills(available); setSkillId(available[0]?.id || '') }
    }).catch(() => { if (!cancelled) setLoadError('Kurikulum belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  async function generate(event: React.FormEvent) {
    event.preventDefault(); setBusy(true); setError('')
    const input = { skill_id: skillId, grade_level: 4 as const, difficulty, purpose }
    const signature = JSON.stringify(input)
    if (!retry.current || retry.current.signature !== signature) retry.current = { signature, id: crypto.randomUUID() }
    try {
      setDraft(await api.generateQuestion({ ...input, request_id: retry.current.id }))
      retry.current = null
      onGenerated?.()
    } catch (err) { setError(err instanceof Error ? err.message : 'Draft belum bisa dibuat. Coba lagi.') }
    finally { setBusy(false) }
  }

  const question = draft?.candidate.candidate
  return <section className="ai-section" aria-labelledby="ai-generator-title">
    <h2 id="ai-generator-title">Buat draft soal kelas 4</h2>
    <p>Gemini menyiapkan satu soal. Admin perlu memeriksa kunci, penjelasan, dan kesesuaian skill sebelum mengaktifkannya.</p>
    {loading ? <p role="status">Memuat keterampilan...</p> : loadError ? <div role="alert"><p>{loadError}</p><button type="button" className="btn btn-secondary" onClick={loadSkills}>Coba lagi</button></div> : !skills.length ? <p role="status">Belum ada keterampilan kelas 4 untuk dibuatkan soal.</p> : <form onSubmit={generate} aria-busy={busy}>
      <fieldset disabled={busy}>
        <legend className="visually-hidden">Spesifikasi draft soal</legend>
        <div className="ai-form-grid">
          <div className="form-group"><label htmlFor="ai-skill">Keterampilan</label><select id="ai-skill" value={skillId} onChange={e => setSkillId(e.target.value)}>{skills.map(skill => <option key={skill.id} value={skill.id}>{skill.name}</option>)}</select></div>
          <div className="form-group"><label htmlFor="ai-difficulty">Kesulitan</label><select id="ai-difficulty" value={difficulty} onChange={e => setDifficulty(Number(e.target.value))}><option value={1}>Level 1</option><option value={2}>Level 2</option></select></div>
          <div className="form-group"><label htmlFor="ai-purpose">Tujuan soal</label><select id="ai-purpose" value={purpose} onChange={e => setPurpose(e.target.value as AIGenerateInput['purpose'])}><option value="practice">Latihan</option><option value="diagnostic">Diagnostic</option><option value="reassessment">Reassessment</option></select></div>
        </div>
        <button className="btn btn-primary" type="submit">{busy ? 'Menyiapkan draft...' : 'Buat satu draft soal'}</button>
      </fieldset>
    </form>}
    {busy && <p role="status">Menunggu Gemini. Hasil akan disimpan sebagai draft untuk review.</p>}
    {error && <p className="alert alert-error" role="alert">{error}</p>}
    {!busy && !error && !draft && <p className="ai-note">Belum ada draft dari sesi ini. Jika Gemini belum siap, admin bisa mengganti key di Pengaturan AI. Bank soal yang sudah aktif tetap dapat dipakai.</p>}
    {question && <div className="ai-preview" aria-labelledby="ai-draft-title">
      <h3 id="ai-draft-title">Draft tersimpan, menunggu review</h3>
      <p role="status">ID: {draft?.candidate.id}. Belum masuk soal siswa.</p>
      <p><strong>{question.question}</strong></p>
      <ol>{question.options.map(option => <li key={option.id}>{option.id}. {option.text}</li>)}</ol>
      <p><strong>Kunci:</strong> {question.correct_answer}</p>
      <p><strong>Penjelasan:</strong> {question.explanation}</p>
      <p className="ai-note">{draft?.candidate.review_note}</p>
    </div>}
  </section>
}
