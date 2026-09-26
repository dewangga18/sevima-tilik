import { useCallback, useEffect, useRef, useState } from 'react'
import type { Skill, User, TeacherClass, StudentOverview, StudentInsight } from '../types'
import { api } from '../services/api'
import './ManagementDashboard.css'

const teacherPages = [
  { id: 'overview', label: 'Ringkasan' },
  { id: 'classes', label: 'Kelas & siswa' },
  { id: 'curriculum', label: 'Kurikulum' },
  { id: 'questions', label: 'Bank soal' },
  { id: 'assistant', label: 'Asisten analisis' },
]
const adminPages = [
  { id: 'overview', label: 'Ringkasan' },
  { id: 'accounts', label: 'Akun & role' },
  { id: 'classes', label: 'Kelas & penempatan' },
  { id: 'curriculum', label: 'Kurikulum' },
  { id: 'questions', label: 'Bank soal' },
]

const pendingContent: Record<string, { title: string; description: string }> = {
  accounts: { title: 'Pengelolaan akun belum tersedia', description: 'Pembuatan akun dan perubahan role belum tersedia pada versi ini.' },
  classes: { title: 'Data kelas belum tersedia', description: 'Kelas, penempatan siswa, dan penugasan guru belum terhubung. Daftar siswa akan ditampilkan setelah data kelas tersedia.' },
  questions: { title: 'Pengelolaan soal belum tersedia', description: 'Pembuatan dan review draft soal belum tersedia. Soal diagnostic saat ini memakai bank awal yang sudah tersimpan.' },
  assistant: { title: 'Asisten analisis belum tersedia', description: 'Analisis AI belum aktif. Fitur ini memerlukan data siswa dari kelas yang ditugaskan kepadamu.' },
}

const studentStatusLabels: Record<StudentOverview['latest_status'], string> = {
  completed: 'Terdata',
  in_progress: 'Sedang mengerjakan',
  unassessed: 'Belum dinilai',
}
const evidenceStatusLabels: Record<string, string> = {
  strong_evidence: 'Bukti kuat',
  inconclusive: 'Belum pasti',
  needs_practice: 'Perlu penguatan',
  mastered: 'Dikuasai',
  unassessed: 'Belum dinilai',
}

export function ManagementDashboard({ user, onLogout, loggingOut }: { user: User; onLogout: () => void; loggingOut: boolean }) {
  const isAdmin = user.role === 'admin'
  const pages = isAdmin ? adminPages : teacherPages
  const [page, setPage] = useState('overview')
  const [menuOpen, setMenuOpen] = useState(false)
  const [skills, setSkills] = useState<Skill[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const heading = useRef<HTMLHeadingElement>(null)
  const pageTitle = pages.find(item => item.id === page)?.label || 'Ringkasan'

  async function loadCurriculum() {
    setLoading(true)
    setError('')
    try { setSkills(await api.getSkills()) }
    catch { setError('Kurikulum belum bisa dimuat. Periksa koneksi lalu coba lagi.') }
    finally { setLoading(false) }
  }

  useEffect(() => {
    let cancelled = false
    api.getSkills()
      .then(data => { if (!cancelled) setSkills(data) })
      .catch(() => { if (!cancelled) setError('Kurikulum belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])
  useEffect(() => { heading.current?.focus() }, [page])

  const [classes, setClasses] = useState<TeacherClass[]>([])
  const [classesLoading, setClassesLoading] = useState(false)
  const [classesError, setClassesError] = useState('')
  const [classesLoaded, setClassesLoaded] = useState(false)
  const [activeClassId, setActiveClassId] = useState('')
  const [students, setStudents] = useState<StudentOverview[]>([])
  const [studentsLoading, setStudentsLoading] = useState(false)
  const [studentsError, setStudentsError] = useState('')
  const [insight, setInsight] = useState<StudentInsight | null>(null)
  const [selectedStudentId, setSelectedStudentId] = useState('')
  const [insightLoading, setInsightLoading] = useState(false)
  const [insightError, setInsightError] = useState('')

  const loadStudents = useCallback(async (classId: string) => {
    if (!classId) return
    setStudentsLoading(true)
    setStudentsError('')
    setInsight(null)
    try {
      setStudents(await api.getClassStudents(classId))
    } catch {
      setStudentsError('Data siswa belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setStudentsLoading(false)
    }
  }, [])

  const loadClasses = useCallback(async (force = false) => {
    if (!force && classesLoaded) return
    setClassesLoaded(true)
    setClassesLoading(true)
    setClassesError('')
    try {
      const data = await api.getTeacherClasses()
      setClasses(data)
      const next = activeClassId && data.some(c => c.id === activeClassId) ? activeClassId : data[0]?.id || ''
      setActiveClassId(next)
      if (next) await loadStudents(next)
    } catch {
      setClassesError('Data kelas belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setClassesLoading(false)
    }
  }, [classesLoaded, activeClassId, loadStudents])

  async function loadInsight(studentId: string) {
    setSelectedStudentId(studentId)
    setInsightLoading(true)
    setInsightError('')
    setInsight(null)
    try {
      setInsight(await api.getStudentInsight(studentId))
    } catch {
      setInsightError('Insight siswa belum bisa dimuat. Coba lagi.')
    } finally {
      setInsightLoading(false)
    }
  }

  function navigate(id: string) {
    setPage(id)
    setMenuOpen(false)
    window.scrollTo({ top: 0 })
    if (id === 'classes' && !isAdmin) void loadClasses()
  }

  const pending = pendingContent[page]
  return <div className="management-layout">
    <aside className="management-sidebar">
      <div className="management-brand"><img className="brand-logo" src="/tilik-wordmark.png" alt="" aria-hidden="true" /><div><strong>Tilik</strong><span>Dashboard {isAdmin ? 'admin' : 'guru'}</span></div></div>
      <button type="button" className="btn btn-secondary management-menu-toggle" aria-expanded={menuOpen} aria-controls="management-navigation" onClick={() => setMenuOpen(!menuOpen)}>{menuOpen ? 'Tutup menu' : 'Buka menu'}</button>
      <nav id="management-navigation" className={menuOpen ? 'management-navigation is-open' : 'management-navigation'} aria-label={`Navigasi ${isAdmin ? 'admin' : 'guru'}`}>
        {pages.map(item => <button type="button" key={item.id} className={page === item.id ? 'management-nav-item is-current' : 'management-nav-item'} aria-current={page === item.id ? 'page' : undefined} onClick={() => navigate(item.id)}>{item.label}</button>)}
      </nav>
      <p className="management-sidebar-note">Numerasi kelas 4<br />Versi awal Tilik</p>
    </aside>
    <div className="management-workspace">
      <header className="management-topbar"><span className="management-workspace-label">Ruang kerja {isAdmin ? 'admin' : 'guru'}</span><div className="navbar-user"><div className="user-info"><span className="user-name">{user.name}</span><span className="role-badge">{isAdmin ? 'Admin' : 'Guru'}</span></div><button type="button" className="btn btn-secondary" onClick={onLogout} disabled={loggingOut}>{loggingOut ? 'Keluar...' : 'Keluar'}</button></div></header>
      <main id="main-content" className="management-content">
        <div className="management-page-heading"><h1 ref={heading} tabIndex={-1}>{pageTitle}</h1><p>{isAdmin ? 'Kelola akses dan kegiatan belajar di Tilik.' : 'Pantau pembelajaran dan siapkan materi untuk siswamu.'}</p></div>
        {page === 'overview' && <>
          <section className="management-overview" aria-labelledby="workspace-status"><h2 id="workspace-status">Selamat datang, {user.name}.</h2><p>{isAdmin ? 'Area administrasi sudah dipisahkan dari dashboard guru. Pengelolaan akun dan kelas masih dalam pengembangan.' : 'Kurikulum awal dan data kelas yang ditugaskan sudah tersedia. Buka menu Kelas & siswa untuk melihat evidence dan rekomendasi.'}</p><button type="button" className="btn btn-primary" onClick={() => navigate(isAdmin ? 'curriculum' : 'classes')}>{isAdmin ? 'Lihat kurikulum' : 'Lihat kelas & siswa'}</button></section>
          <section className="management-status-section" aria-busy={loading}><h2>Ketersediaan fitur</h2><dl className="management-status-list"><div><dt>Kurikulum kelas 4</dt><dd role="status">{loading ? 'Memuat...' : error ? 'Belum bisa dimuat' : `${skills.length} keterampilan tersedia`}</dd></div><div><dt>{isAdmin ? 'Akun & kelas' : 'Data siswa & completion'}</dt><dd>{isAdmin ? 'Belum tersedia' : 'Tersedia untuk kelas yang ditugaskan'}</dd></div><div><dt>{isAdmin ? 'Pembuatan soal' : 'Asisten analisis'}</dt><dd>Belum tersedia</dd></div></dl></section>
        </>}
        {(page === 'overview' || page === 'curriculum') && error && <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={loadCurriculum}>Coba lagi</button></div>}
        {page === 'curriculum' && <section className="management-curriculum" aria-label="Kurikulum kelas 4" aria-busy={loading}>
          {loading ? <p role="status">Memuat kurikulum...</p> : !error && (skills.length ? <><p className="curriculum-loaded" role="status">{skills.length} keterampilan berhasil dimuat.</p><div className="management-table-scroll" tabIndex={0} role="region" aria-label="Tabel keterampilan, geser untuk melihat kolom"><table className="management-table"><thead><tr><th scope="col">Keterampilan</th><th scope="col">Domain</th><th scope="col">Prasyarat</th></tr></thead><tbody>{skills.map(skill => <tr key={skill.id}><th scope="row">{skill.name}<span>{skill.description}</span></th><td>{skill.domain}</td><td>{skill.prereqs?.length ? skill.prereqs.map(id => skills.find(item => item.id === id)?.name || id).join(', ') : 'Tidak ada'}</td></tr>)}</tbody></table></div></> : <p role="status">Belum ada keterampilan yang tersedia.</p>)}
        </section>}
        {page === 'classes' && !isAdmin && <section className="management-curriculum" aria-label="Kelas dan siswa" aria-busy={classesLoading || studentsLoading}>
          {classesLoading ? <p role="status">Memuat kelas...</p> : classesError ? <div className="home-data-error" role="alert"><p>{classesError}</p><button type="button" className="btn btn-secondary" onClick={() => loadClasses(true)}>Coba lagi</button></div> : !classes.length ? <p role="status">Belum ada kelas yang ditugaskan kepadamu.</p> : <>
            <div className="teacher-class-list">
              {classes.map(c => <button type="button" key={c.id} className={c.id === activeClassId ? 'btn btn-primary' : 'btn btn-secondary'} aria-pressed={c.id === activeClassId} onClick={() => { setActiveClassId(c.id); void loadStudents(c.id) }}>{c.name}<span>{c.assessed_count} terdata dari {c.student_count} siswa</span></button>)}
            </div>
            {studentsLoading ? <p role="status">Memuat siswa...</p> : studentsError ? <div className="home-data-error" role="alert"><p>{studentsError}</p><button type="button" className="btn btn-secondary" onClick={() => loadStudents(activeClassId)}>Coba lagi</button></div> : !students.length ? <p role="status">Belum ada siswa terdaftar di kelas ini.</p> : <div className="management-table-scroll" tabIndex={0} role="region" aria-label="Tabel siswa, geser untuk melihat kolom"><table className="management-table"><thead><tr><th scope="col">Siswa</th><th scope="col">Status</th><th scope="col">Skill terdata</th><th scope="col">Gap terlihat</th><th scope="col">Kandidat akar</th><th scope="col"><span className="visually-hidden">Aksi</span></th></tr></thead><tbody>{students.map(s => <tr key={s.id}><th scope="row">{s.name}</th><td><span className={`role-badge status-${s.latest_status}`}>{studentStatusLabels[s.latest_status]}</span></td><td>{s.assessed_skills}</td><td>{s.visible_gap_count}</td><td>{s.root_gap_count}</td><td><button type="button" className="btn btn-secondary btn-sm" onClick={() => loadInsight(s.id)} aria-expanded={insight?.student_id === s.id}>Lihat insight</button></td></tr>)}</tbody></table></div>}
            {insightLoading && <p role="status">Memuat insight siswa...</p>}
            {insightError && <div className="home-data-error" role="alert"><p>{insightError}</p><button type="button" className="btn btn-secondary" onClick={() => selectedStudentId && loadInsight(selectedStudentId)}>Coba lagi</button></div>}
            {insight && <div className="teacher-insight-panel">
              <h3>Insight: {insight.student_name} ({insight.classroom_name})</h3>
              {!insight.assessment_id ? <p role="status">Siswa ini belum menyelesaikan diagnostic; belum ada evidence yang bisa ditampilkan.</p> : <>
                <h4>Evidence per skill</h4>
                <div className="management-table-scroll" tabIndex={0} role="region" aria-label="Tabel evidence skill"><table className="management-table"><thead><tr><th scope="col">Skill</th><th scope="col">Status</th><th scope="col">Bukti benar</th><th scope="col">Keyakinan</th><th scope="col">Kandidat akar gap</th></tr></thead><tbody>{insight.results.map(r => <tr key={r.skill_id}><th scope="row">{r.skill_name}</th><td><span className={`role-badge status-${r.status}`}>{evidenceStatusLabels[r.status] || r.status}</span></td><td>{r.total_correct}/{r.total_answered}</td><td>{r.confidence}</td><td>{r.is_root_gap ? 'Ya' : '—'}</td></tr>)}</tbody></table></div>
                {insight.recommendations.length > 0 && <><h4>Rekomendasi pembelajaran berikutnya</h4><ul className="teacher-recommendation-list">{insight.recommendations.map(rec => <li key={rec.skill_id}><strong>{rec.skill_name}</strong><span>{rec.reason}</span></li>)}</ul></>}
                {insight.progress.length > 0 && <><h4>Perubahan progress belajar</h4><ul className="teacher-progress-list">{insight.progress.map(p => <li key={p.skill_id}><strong>{p.skill_name}</strong><span>Skor {p.score}% · {evidenceStatusLabels[p.status] || p.status} · {p.evidence_count} bukti</span></li>)}</ul></>}
              </>}
            </div>}
          </>}
        </section>}
        {pending && !(page === 'classes' && !isAdmin) && <section className="management-empty"><h2>{pending.title}</h2><p>{pending.description}</p><button type="button" className="btn btn-secondary" onClick={() => navigate('overview')}>Kembali ke ringkasan</button></section>}
      </main>
    </div>
  </div>
}
