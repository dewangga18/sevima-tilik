import type { Assessment, AssessmentHistory, User } from '../types'
import './StudentHome.css'

interface StudentHomeProps {
  user: User
  latest: Assessment | null
  history: AssessmentHistory | null
  loading: boolean
  busy: boolean
  error: string
  onStart: () => void
  onOpen: (id: string) => void
  onRetry: () => void
}

export function StudentHome({ user, latest, history, loading, busy, error, onStart, onOpen, onRetry }: StudentHomeProps) {
  const firstName = user.name.trim().split(/\s+/)[0]
  const initials = user.name.trim().split(/\s+/).slice(0, 2).map(name => name[0]).join('')
  const isActive = latest?.status === 'in_progress'
  const lastCompleted = history?.items.find(item => item.status === 'completed')
  const dataReady = !loading && !error && history !== null
  const status = !dataReady ? 'Belum tersedia' : isActive ? 'Sedang dikerjakan' : history.completed_count > 0 ? 'Sudah mulai belajar' : 'Belum mulai'
  const startLabel = isActive ? 'Lanjutkan cek pemahaman' : history && history.completed_count > 0 ? 'Cek pemahaman lagi' : 'Mulai cek pemahaman'

  return (
    <div className="student-home">
      <section className="home-welcome" aria-labelledby="welcome-title">
        <div className="welcome-copy">
          <h1 id="welcome-title" tabIndex={-1}>Halo, {firstName}.</h1>
          <p className="welcome-lead">Selangkah lagi untuk<br className="desktop-break" /> lebih paham.</p>
          <p className="welcome-description">Di sini, kamu bisa mengenali apa yang sudah kamu pahami dan bagian yang perlu dipelajari lagi. Mulai dari langkah kecil, sesuai kemampuanmu.</p>
        </div>
        <div className="learning-invitation">
          <h2>{isActive ? 'Lanjutkan langkahmu' : 'Kenali pemahamanmu'}</h2>
          <p>{isActive ? 'Cek pemahamanmu belum selesai. Kamu bisa melanjutkan saat sudah siap.' : 'Coba beberapa soal tentang bilangan dan pecahan. Kerjakan semampumu untuk menemukan pijakan belajar berikutnya.'}</p>
          <button type="button" className="btn btn-primary" onClick={onStart} disabled={loading || busy || Boolean(error)}>
            {busy ? 'Menyiapkan aktivitas...' : loading ? 'Memuat aktivitas...' : startLabel}
          </button>
          <span className="invitation-note">Tidak ada batas waktu. Kerjakan dengan tenang.</span>
        </div>
      </section>

      {error && <div className="home-data-error" role="alert">
        <p>{error}</p>
        <button type="button" className="btn btn-secondary" onClick={onRetry} disabled={busy}>Coba lagi</button>
      </div>}

      <section className="profile-section" id="student-profile" tabIndex={-1} aria-labelledby="profile-title" aria-busy={loading}>
        <div className="section-heading">
          <h2 id="profile-title">Profil belajarmu</h2>
          <p>Setiap langkahmu tercatat di sini.</p>
        </div>
        <div className="profile-panel">
          <div className="profile-identity">
            <span className="profile-initials" aria-hidden="true">{initials}</span>
            <div><h3>{user.name}</h3><p>Kelas {user.grade_level || 4} {(user.grade_level || 4) <= 6 ? 'SD' : 'SMP'}</p><span className="profile-status">{loading ? 'Memuat status...' : status}</span></div>
          </div>
          <dl className="profile-statistics">
            <div><dt>Cek pemahaman selesai</dt><dd>{loading ? 'Memuat...' : dataReady ? history.completed_count : 'Belum tersedia'}</dd><dd className="statistic-note">{dataReady && history.completed_count === 0 ? 'Langkah pertamamu menanti' : 'Dari aktivitas yang tersimpan'}</dd></div>
            <div><dt>Konsep terukur</dt><dd>{loading ? 'Memuat...' : !dataReady ? 'Belum tersedia' : lastCompleted ? lastCompleted.assessed_skill_count : 'Belum diukur'}</dd><dd className="statistic-note">Dari hasil selesai terbaru</dd></div>
          </dl>
        </div>
      </section>

      <section className="history-section" id="recent-history" tabIndex={-1} aria-labelledby="history-title" aria-busy={loading}>
        <div className="section-heading"><h2 id="history-title">Aktivitas terakhir</h2><p>Lihat kembali perjalanan belajarmu.</p></div>
        {loading ? <div className="history-empty" role="status"><p>Memuat riwayat belajarmu...</p></div>
          : error ? <div className="history-empty"><p>Riwayat belum bisa ditampilkan. Coba muat kembali lewat tombol di atas.</p></div>
          : history?.items.length ? <ul className="history-list">
            {history.items.map(item => <li key={item.id} className="history-row">
              <div className="history-activity"><h3>Cek pemahaman bilangan &amp; pecahan</h3><p><time dateTime={item.completed_at || item.started_at}>{new Date(item.completed_at || item.started_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })}</time> · Kelas {item.grade_level}</p></div>
              <span className={`history-status ${item.status === 'completed' ? 'history-completed' : ''}`}>{item.status === 'completed' ? 'Selesai' : 'Belum selesai'}</span>
              <button type="button" className="btn btn-secondary" onClick={() => onOpen(item.id)} disabled={busy}>{item.status === 'completed' ? 'Lihat hasil' : 'Lanjutkan'}</button>
            </li>)}
          </ul> : <div className="history-empty"><h3>Perjalananmu dimulai di sini.</h3><p>Belum ada aktivitas belajar. Setelah kamu mulai cek pemahaman, riwayatnya akan muncul di bagian ini.</p></div>}
      </section>
    </div>
  )
}
