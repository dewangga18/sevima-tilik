import { useEffect, useRef, useState } from 'react'
import type { Skill, User } from '../types'
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

  function navigate(id: string) {
    setPage(id)
    setMenuOpen(false)
    window.scrollTo({ top: 0 })
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
          <section className="management-overview" aria-labelledby="workspace-status"><h2 id="workspace-status">Selamat datang, {user.name}.</h2><p>{isAdmin ? 'Area administrasi sudah dipisahkan dari dashboard guru. Pengelolaan akun dan kelas masih dalam pengembangan.' : 'Kurikulum awal tersedia. Data siswa dan completion kelas akan muncul setelah penugasan kelas terhubung.'}</p><button type="button" className="btn btn-primary" onClick={() => navigate('curriculum')}>Lihat kurikulum</button></section>
          <section className="management-status-section" aria-busy={loading}><h2>Ketersediaan fitur</h2><dl className="management-status-list"><div><dt>Kurikulum kelas 4</dt><dd role="status">{loading ? 'Memuat...' : error ? 'Belum bisa dimuat' : `${skills.length} keterampilan tersedia`}</dd></div><div><dt>{isAdmin ? 'Akun & kelas' : 'Data siswa & completion'}</dt><dd>Belum tersedia</dd></div><div><dt>{isAdmin ? 'Pembuatan soal' : 'Asisten analisis'}</dt><dd>Belum tersedia</dd></div></dl></section>
        </>}
        {(page === 'overview' || page === 'curriculum') && error && <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={loadCurriculum}>Coba lagi</button></div>}
        {page === 'curriculum' && <section className="management-curriculum" aria-label="Kurikulum kelas 4" aria-busy={loading}>
          {loading ? <p role="status">Memuat kurikulum...</p> : !error && (skills.length ? <><p className="curriculum-loaded" role="status">{skills.length} keterampilan berhasil dimuat.</p><div className="management-table-scroll" tabIndex={0} role="region" aria-label="Tabel keterampilan, geser untuk melihat kolom"><table className="management-table"><thead><tr><th scope="col">Keterampilan</th><th scope="col">Domain</th><th scope="col">Prasyarat</th></tr></thead><tbody>{skills.map(skill => <tr key={skill.id}><th scope="row">{skill.name}<span>{skill.description}</span></th><td>{skill.domain}</td><td>{skill.prereqs?.length ? skill.prereqs.map(id => skills.find(item => item.id === id)?.name || id).join(', ') : 'Tidak ada'}</td></tr>)}</tbody></table></div></> : <p role="status">Belum ada keterampilan yang tersedia.</p>)}
        </section>}
        {pending && <section className="management-empty"><h2>{pending.title}</h2><p>{pending.description}</p><button type="button" className="btn btn-secondary" onClick={() => navigate('overview')}>Kembali ke ringkasan</button></section>}
      </main>
    </div>
  </div>
}
