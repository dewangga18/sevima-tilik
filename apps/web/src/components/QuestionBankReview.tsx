import { useCallback, useEffect, useState } from 'react'
import type { CandidateReview, CandidateStatus, QuestionBankSummary } from '../types'
import { api } from '../services/api'
import './QuestionBankReview.css'

const reviewFilters: { id: CandidateStatus; label: string }[] = [
  { id: 'draft', label: 'Menunggu review' },
  { id: 'approved', label: 'Disetujui' },
  { id: 'rejected', label: 'Ditolak' },
]

const PAGE_SIZE = 25

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error && err.message ? err.message : fallback
}

function useQuestionBankSummary() {
  const [summary, setSummary] = useState<QuestionBankSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setSummary(await api.getQuestionBankSummary())
    } catch {
      setError('Ringkasan bank soal belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    api.getQuestionBankSummary()
      .then(data => { if (!cancelled) setSummary(data) })
      .catch(() => { if (!cancelled) setError('Ringkasan bank soal belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  return { summary, loading, error, load }
}

function useCandidates(status: CandidateStatus) {
  const [candidates, setCandidates] = useState<CandidateReview[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setCandidates(await api.listQuestionCandidates(status, PAGE_SIZE))
    } catch {
      setError('Daftar kandidat belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setLoading(false)
    }
  }, [status])

  useEffect(() => {
    let cancelled = false
    api.listQuestionCandidates(status, PAGE_SIZE)
      .then(data => { if (!cancelled) setCandidates(data) })
      .catch(() => { if (!cancelled) setError('Daftar kandidat belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [status])

  return { candidates, loading, error, load }
}

function CandidateCard({ candidate, onDecided }: { candidate: CandidateReview; onDecided: (message: string) => void }) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [note, setNote] = useState('')
  const [showReject, setShowReject] = useState(false)
  const isDraft = candidate.status === 'draft'
  const question = candidate.candidate

  async function approve() {
    setBusy(true)
    setError('')
    try {
      const result = await api.approveQuestionCandidate(candidate.id, note)
      onDecided(result.activated
        ? `${candidate.id} disetujui dan masuk bank soal aktif.`
        : `${candidate.id} tidak diaktifkan: ${result.report.items[0]?.reason || 'alasan tidak diketahui'}`)
    } catch (err) {
      setError(errorMessage(err, 'Keputusan belum bisa disimpan. Silakan coba lagi.'))
    } finally {
      setBusy(false)
    }
  }

  async function reject() {
    if (!note.trim()) {
      setError('Tulis alasan penolakan supaya reviewer berikutnya punya konteks.')
      return
    }
    setBusy(true)
    setError('')
    try {
      await api.rejectQuestionCandidate(candidate.id, note)
      onDecided(`${candidate.id} ditolak.`)
      setShowReject(false)
    } catch (err) {
      setError(errorMessage(err, 'Keputusan belum bisa disimpan. Silakan coba lagi.'))
    } finally {
      setBusy(false)
    }
  }

  return <li className="bank-candidate">
    <div className="bank-candidate-head">
      <strong>{question.prompt}</strong>
      <span className="bank-candidate-meta">
        {question.skill} · level {question.difficulty} · {question.purpose} · kelas {question.grade}
      </span>
    </div>
    <ol className="bank-options">
      {question.options.map(option => {
        const isKey = option.id === question.correct_answer
        return <li key={option.id} className={isKey ? 'bank-option is-key' : 'bank-option'}>
          <span className="bank-option-id">{option.id}</span>
          <span>{option.text}</span>
          {isKey && <span className="bank-key-flag">kunci</span>}
        </li>
      })}
    </ol>
    {question.explanation && <p className="bank-explanation"><strong>Penjelasan:</strong> {question.explanation}</p>}
    {question.misconceptions?.length > 0 && <p className="bank-explanation"><strong>Misconception:</strong> {question.misconceptions.map(m => m.wrong_answer).join(', ')}</p>}
    <p className="bank-candidate-source">Sumber: {candidate.source_file}{question.license_note ? ` · ${question.license_note}` : ''}</p>

    {isDraft && !candidate.eligible && <p className="bank-blocked" role="note">{candidate.blocked_reason}</p>}

    {isDraft && <>
      <div className="form-group">
        <label htmlFor={`note-${candidate.id}`}>Catatan review (opsional saat menyetujui, wajib saat menolak)</label>
        <input id={`note-${candidate.id}`} className="form-control" value={note} onChange={e => setNote(e.target.value)} maxLength={200} placeholder="misal: kunci sudah diverifikasi" />
      </div>
      <div className="bank-actions">
        <button type="button" className="btn btn-primary btn-sm" onClick={approve} disabled={busy || !candidate.eligible}>
          {busy ? 'Menyimpan...' : 'Setujui & aktifkan'}
        </button>
        {showReject
          ? <button type="button" className="btn btn-secondary btn-sm" onClick={reject} disabled={busy}>{busy ? 'Menyimpan...' : 'Konfirmasi tolak'}</button>
          : <button type="button" className="btn btn-secondary btn-sm" onClick={() => setShowReject(true)} disabled={busy}>Tolak</button>}
        {showReject && <button type="button" className="btn btn-secondary btn-sm" onClick={() => setShowReject(false)} disabled={busy}>Batal</button>}
      </div>
    </>}

    {!isDraft && <p className="bank-review-note">Status: {candidate.status === 'approved' ? 'disetujui' : 'ditolak'}{candidate.review_note ? ` — ${candidate.review_note}` : ''}</p>}
    {error && <p className="alert alert-error" role="alert">{error}</p>}
  </li>
}

export function QuestionBankReview() {
  const { summary, loading: summaryLoading, error: summaryError, load: reloadSummary } = useQuestionBankSummary()
  const [status, setStatus] = useState<CandidateStatus>('draft')
  const { candidates, loading, error, load } = useCandidates(status)
  const [notice, setNotice] = useState('')

  function handleDecided(message: string) {
    setNotice(message)
    void load()
    void reloadSummary()
  }

  return <div className="admin-view">
    <section className="admin-list" aria-labelledby="bank-summary-title" aria-busy={summaryLoading}>
      <h2 id="bank-summary-title">Ringkasan bank soal</h2>
      {summaryLoading ? <p role="status">Memuat ringkasan bank soal...</p>
        : summaryError ? <div className="home-data-error" role="alert"><p>{summaryError}</p><button type="button" className="btn btn-secondary" onClick={() => reloadSummary()}>Coba lagi</button></div>
        : summary && <>
          <div className="bank-summary-grid">
            <div><span>Soal aktif</span><strong>{summary.active.reduce((total, row) => total + row.question_count, 0)}</strong></div>
            {summary.candidate.map(row => <div key={row.status}><span>{row.status === 'draft' ? 'Menunggu review' : row.status === 'approved' ? 'Disetujui' : 'Ditolak'}</span><strong>{row.count}</strong></div>)}
          </div>
          <div className="management-table-scroll" tabIndex={0} role="region" aria-label="Tabel soal aktif per keterampilan, geser untuk melihat kolom">
            <table className="management-table">
              <thead><tr><th scope="col">Keterampilan</th><th scope="col">Soal aktif</th></tr></thead>
              <tbody>{summary.active.map(row => <tr key={row.skill_id}><th scope="row">{row.skill_name}</th><td>{row.question_count}</td></tr>)}</tbody>
            </table>
          </div>
        </>}
    </section>

    <section className="admin-list" aria-labelledby="bank-review-title" aria-busy={loading}>
      <div className="bank-review-head">
        <h2 id="bank-review-title">Review kandidat</h2>
        <div className="bank-filters" role="group" aria-label="Filter status kandidat">
          {reviewFilters.map(filter => <button
            key={filter.id}
            type="button"
            className={status === filter.id ? 'btn btn-primary btn-sm' : 'btn btn-secondary btn-sm'}
            aria-pressed={status === filter.id}
            onClick={() => { setNotice(''); setStatus(filter.id) }}
          >{filter.label}</button>)}
        </div>
      </div>
      <p className="admin-form-note">Kunci jawaban ditampilkan agar admin bisa memverifikasi sebelum menyetujui. Kunci ini tidak pernah dikirim ke siswa.</p>
      {notice && <p className="alert" role="status">{notice}</p>}
      {loading ? <p role="status">Memuat kandidat...</p>
        : error ? <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={() => load()}>Coba lagi</button></div>
        : !candidates.length ? <p role="status">Tidak ada kandidat pada status ini.</p>
        : <ul className="bank-candidate-list">
          {candidates.map(candidate => <CandidateCard key={candidate.id} candidate={candidate} onDecided={handleDecided} />)}
        </ul>}
    </section>
  </div>
}
