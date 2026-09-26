import { useCallback, useEffect, useState } from 'react'
import type { AdminClass, ClassRoster, Role, User } from '../types'
import { api } from '../services/api'
import './AdminManagement.css'

const roleLabels: Record<Role, string> = { student: 'Siswa', teacher: 'Guru', admin: 'Admin' }

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error && err.message ? err.message : fallback
}

function useAdminUsers() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setUsers(await api.getAdminUsers())
    } catch {
      setError('Data akun belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    api.getAdminUsers()
      .then(data => { if (!cancelled) setUsers(data) })
      .catch(() => { if (!cancelled) setError('Data akun belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  return { users, loading, error, load }
}

function CreateUserForm({ onCreated }: { onCreated: (user: User) => void }) {
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [role, setRole] = useState<Role>('student')
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError('')
    setNotice('')
    try {
      const created = await api.createAdminUser({ email, name, role, grade_level: 4, password })
      onCreated(created)
      setEmail('')
      setName('')
      setPassword('')
      setNotice(`Akun ${created.name} dibuat dengan role ${roleLabels[created.role]}.`)
    } catch (err) {
      setError(errorMessage(err, 'Akun belum bisa dibuat. Silakan coba lagi.'))
    } finally {
      setBusy(false)
    }
  }

  return <form className="admin-form" onSubmit={handleSubmit} aria-labelledby="create-user-title">
    <h3 id="create-user-title">Buat akun baru</h3>
    <div className="admin-form-grid">
      <div className="form-group">
        <label htmlFor="admin-user-name">Nama lengkap</label>
        <input id="admin-user-name" className="form-control" value={name} onChange={e => setName(e.target.value)} required maxLength={100} autoComplete="off" />
      </div>
      <div className="form-group">
        <label htmlFor="admin-user-email">Email</label>
        <input id="admin-user-email" className="form-control" type="email" value={email} onChange={e => setEmail(e.target.value)} required autoComplete="off" />
      </div>
      <div className="form-group">
        <label htmlFor="admin-user-role">Role</label>
        <select id="admin-user-role" className="form-control" value={role} onChange={e => setRole(e.target.value as Role)}>
          <option value="student">Siswa</option>
          <option value="teacher">Guru</option>
          <option value="admin">Admin</option>
        </select>
      </div>
      <div className="form-group">
        <label htmlFor="admin-user-password">Kata sandi awal</label>
        <input id="admin-user-password" className="form-control" type="password" value={password} onChange={e => setPassword(e.target.value)} required minLength={8} maxLength={72} autoComplete="new-password" />
        <p className="admin-form-note">8–72 karakter. Akun dapat langsung login dengan role yang dipilih.</p>
      </div>
    </div>
    <button type="submit" className="btn btn-primary" disabled={busy}>{busy ? 'Menyimpan...' : 'Buat akun'}</button>
    {role === 'student' && <p className="admin-form-note">Siswa baru/default kelas 4. Hanya kelas 4 yang memiliki konten pada versi ini.</p>}
    {error && <p className="alert alert-error" role="alert">{error}</p>}
    {notice && <p className="alert" role="status">{notice}</p>}
  </form>
}

function UserRow({ user, isSelf, onChanged }: { user: User; isSelf: boolean; onChanged: () => void }) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const isActive = user.is_active !== false

  async function changeRole(nextRole: Role) {
    if (nextRole === user.role) return
    setBusy(true)
    setError('')
    setNotice('')
    try {
      await api.updateAdminUser(user.id, { role: nextRole })
      setNotice(`Role ${user.name} diubah menjadi ${roleLabels[nextRole]}.`)
      onChanged()
    } catch (err) {
      setError(errorMessage(err, 'Role belum bisa diubah. Silakan coba lagi.'))
    } finally {
      setBusy(false)
    }
  }

  async function toggleActive() {
    setBusy(true)
    setError('')
    setNotice('')
    try {
      await api.updateAdminUser(user.id, { is_active: !isActive })
      setNotice(isActive ? `Akun ${user.name} dinonaktifkan.` : `Akun ${user.name} diaktifkan kembali.`)
      onChanged()
    } catch (err) {
      setError(errorMessage(err, 'Status akun belum bisa diubah. Silakan coba lagi.'))
    } finally {
      setBusy(false)
    }
  }

  return <>
    <tr>
      <th scope="row">{user.name}{isSelf && <span>Akun Anda</span>}</th>
      <td>{user.email}</td>
      <td><span className={`role-badge role-${user.role}`}>{roleLabels[user.role]}</span></td>
      <td>{user.role === 'student' ? `Kelas ${user.grade_level ?? 4}` : '—'}</td>
      <td>
        <label className="visually-hidden" htmlFor={`role-${user.id}`}>Role untuk {user.name}</label>
        <select
          id={`role-${user.id}`}
          className="form-control admin-inline-select"
          value={user.role}
          disabled={busy || isSelf}
          onChange={e => changeRole(e.target.value as Role)}
        >
          <option value="student">Siswa</option>
          <option value="teacher">Guru</option>
          <option value="admin">Admin</option>
        </select>
      </td>
      <td>
        <span className={`role-badge status-${isActive ? 'completed' : 'unassessed'}`}>{isActive ? 'Aktif' : 'Nonaktif'}</span>
      </td>
      <td>
        <button type="button" className="btn btn-secondary btn-sm" onClick={toggleActive} disabled={busy || isSelf}>
          {isActive ? 'Nonaktifkan' : 'Aktifkan'}
        </button>
      </td>
    </tr>
    {(error || notice) && <tr className="admin-feedback-row">
      <td colSpan={7}>
        {error ? <span className="alert alert-error" role="alert">{error}</span> : <span className="alert" role="status">{notice}</span>}
      </td>
    </tr>}
  </>
}

function AccountsView({ currentUserId }: { currentUserId: string }) {
  const { users, loading, error, load } = useAdminUsers()

  return <div className="admin-view">
    <CreateUserForm onCreated={() => void load()} />
    <section className="admin-list" aria-labelledby="admin-users-title" aria-busy={loading}>
      <h2 id="admin-users-title">Akun terdaftar</h2>
      {loading ? <p role="status">Memuat akun...</p>
        : error ? <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={() => load()}>Coba lagi</button></div>
        : !users.length ? <p role="status">Belum ada akun terdaftar.</p>
        : <div className="management-table-scroll" tabIndex={0} role="region" aria-label="Tabel akun, geser untuk melihat kolom">
          <table className="management-table">
            <thead><tr>
              <th scope="col">Nama</th><th scope="col">Email</th><th scope="col">Role</th><th scope="col">Kelas</th>
              <th scope="col">Ubah role</th><th scope="col">Status</th><th scope="col"><span className="visually-hidden">Aksi</span></th>
            </tr></thead>
            <tbody>{users.map(user => <UserRow key={user.id} user={user} isSelf={user.id === currentUserId} onChanged={load} />)}</tbody>
          </table>
        </div>}
      {!loading && !error && users.length > 0 && <p className="admin-form-note">Role dan status aktif berlaku pada request berikutnya tanpa perlu login ulang. Akun Anda sendiri terkunci dari perubahan ini.</p>}
    </section>
  </div>
}

function CreateClassForm({ onCreated }: { onCreated: (klass: AdminClass) => void }) {
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      const created = await api.createAdminClass({ name, grade_level: 4 })
      onCreated(created)
      setName('')
    } catch (err) {
      setError(errorMessage(err, 'Kelas belum bisa dibuat. Silakan coba lagi.'))
    } finally {
      setBusy(false)
    }
  }

  return <form className="admin-form" onSubmit={handleSubmit} aria-labelledby="create-class-title">
    <h3 id="create-class-title">Buat kelas</h3>
    <div className="admin-form-grid">
      <div className="form-group">
        <label htmlFor="admin-class-name">Nama kelas</label>
        <input id="admin-class-name" className="form-control" value={name} onChange={e => setName(e.target.value)} required maxLength={100} placeholder="Kelas 4C" autoComplete="off" />
      </div>
      <div className="form-group">
        <label htmlFor="admin-class-grade">Tingkat</label>
        <input id="admin-class-grade" className="form-control" value="Kelas 4" readOnly aria-describedby="admin-class-grade-note" />
        <p className="admin-form-note" id="admin-class-grade-note">Hanya kelas 4 yang memiliki konten.</p>
      </div>
    </div>
    <button type="submit" className="btn btn-primary" disabled={busy}>{busy ? 'Menyimpan...' : 'Buat kelas'}</button>
    {error && <p className="alert alert-error" role="alert">{error}</p>}
  </form>
}

function RosterPanel({ klass, users, onChanged }: { klass: AdminClass; users: User[]; onChanged: () => void }) {
  const [roster, setRoster] = useState<ClassRoster | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [busyId, setBusyId] = useState('')
  const [actionError, setActionError] = useState('')
  const [notice, setNotice] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setRoster(await api.getClassRoster(klass.id))
    } catch {
      setError('Data peserta kelas belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setLoading(false)
    }
  }, [klass.id])

  useEffect(() => {
    let cancelled = false
    api.getClassRoster(klass.id)
      .then(data => { if (!cancelled) setRoster(data) })
      .catch(() => { if (!cancelled) setError('Data peserta kelas belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [klass.id])

  async function place(kind: 'enrollment' | 'assignment', userId: string, attach: boolean) {
    setBusyId(`${kind}:${userId}`)
    setActionError('')
    setNotice('')
    try {
      if (kind === 'enrollment') {
        await (attach ? api.addEnrollment(klass.id, userId) : api.removeEnrollment(klass.id, userId))
      } else {
        await (attach ? api.addAssignment(klass.id, userId) : api.removeAssignment(klass.id, userId))
      }
      const subject = users.find(u => u.id === userId)?.name || userId
      setNotice(attach ? `${subject} ditambahkan ke ${klass.name}.` : `${subject} dikeluarkan dari ${klass.name}.`)
      await load()
      onChanged()
    } catch (err) {
      setActionError(errorMessage(err, 'Penempatan belum bisa disimpan. Silakan coba lagi.'))
    } finally {
      setBusyId('')
    }
  }

  const activeStudents = users.filter(u => u.role === 'student' && u.is_active !== false)
  const activeTeachers = users.filter(u => u.role === 'teacher' && u.is_active !== false)
  const enrolled = new Set(roster?.student_ids || [])
  const assigned = new Set(roster?.teacher_ids || [])

  return <section className="admin-roster" aria-labelledby="roster-title" aria-busy={loading}>
    <h3 id="roster-title">Penempatan {klass.name}</h3>
    {loading ? <p role="status">Memuat peserta kelas...</p>
      : error ? <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={() => load()}>Coba lagi</button></div>
      : <>
        {actionError && <p className="alert alert-error" role="alert">{actionError}</p>}
        {notice && <p className="alert" role="status">{notice}</p>}
        <div className="admin-roster-columns">
          <div>
            <h4>Siswa ({enrolled.size})</h4>
            {!enrolled.size ? <p role="status">Belum ada siswa di kelas ini.</p> : <ul className="admin-roster-list">
              {activeStudents.filter(s => enrolled.has(s.id)).map(student => <li key={student.id}>
                <span>{student.name}</span>
                <button type="button" className="btn btn-secondary btn-sm" disabled={busyId === `enrollment:${student.id}`} onClick={() => place('enrollment', student.id, false)}>Keluarkan</button>
              </li>)}
            </ul>}
            <PlacementPicker label="Tambah siswa" users={activeStudents.filter(s => !enrolled.has(s.id))} busyId={busyId} onPick={id => place('enrollment', id, true)} />
          </div>
          <div>
            <h4>Guru ({assigned.size})</h4>
            {!assigned.size ? <p role="status">Belum ada guru yang ditugaskan.</p> : <ul className="admin-roster-list">
              {activeTeachers.filter(t => assigned.has(t.id)).map(teacher => <li key={teacher.id}>
                <span>{teacher.name}</span>
                <button type="button" className="btn btn-secondary btn-sm" disabled={busyId === `assignment:${teacher.id}`} onClick={() => place('assignment', teacher.id, false)}>Lepas</button>
              </li>)}
            </ul>}
            <PlacementPicker label="Tugaskan guru" users={activeTeachers.filter(t => !assigned.has(t.id))} busyId={busyId} onPick={id => place('assignment', id, true)} />
          </div>
        </div>
      </>}
  </section>
}

function PlacementPicker({ label, users, busyId, onPick }: { label: string; users: User[]; busyId: string; onPick: (id: string) => void }) {
  const [selected, setSelected] = useState('')
  const options = users.filter(u => busyId !== `enrollment:${u.id}` && busyId !== `assignment:${u.id}`)

  return <div className="admin-placement">
    <label className="visually-hidden" htmlFor={`placement-${label}`}>{label}</label>
    <select id={`placement-${label}`} className="form-control admin-inline-select" value={selected} onChange={e => setSelected(e.target.value)} disabled={!options.length}>
      <option value="">{options.length ? label : `Tidak ada pilihan untuk ${label.toLowerCase()}`}</option>
      {options.map(user => <option key={user.id} value={user.id}>{user.name}</option>)}
    </select>
    <button type="button" className="btn btn-secondary btn-sm" disabled={!selected || Boolean(busyId)} onClick={() => { onPick(selected); setSelected('') }}>
      {busyId ? 'Menyimpan...' : 'Simpan'}
    </button>
  </div>
}

function ClassesView({ users }: { users: User[] }) {
  const [classes, setClasses] = useState<AdminClass[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeId, setActiveId] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const data = await api.getAdminClasses()
      setClasses(data)
      setActiveId(current => (current && data.some(c => c.id === current) ? current : data[0]?.id || ''))
    } catch {
      setError('Data kelas belum bisa dimuat. Periksa koneksi lalu coba lagi.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    api.getAdminClasses()
      .then(data => {
        if (cancelled) return
        setClasses(data)
        setActiveId(current => (current && data.some(c => c.id === current) ? current : data[0]?.id || ''))
      })
      .catch(() => { if (!cancelled) setError('Data kelas belum bisa dimuat. Periksa koneksi lalu coba lagi.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  const active = classes.find(c => c.id === activeId)

  return <div className="admin-view">
    <CreateClassForm onCreated={created => { void load(); setActiveId(created.id) }} />
    <section className="admin-list" aria-labelledby="admin-classes-title" aria-busy={loading}>
      <h2 id="admin-classes-title">Kelas terdaftar</h2>
      {loading ? <p role="status">Memuat kelas...</p>
        : error ? <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={() => load()}>Coba lagi</button></div>
        : !classes.length ? <p role="status">Belum ada kelas. Buat kelas pertama di atas.</p>
        : <>
          <div className="teacher-class-list">
            {classes.map(klass => <button type="button" key={klass.id} className={klass.id === activeId ? 'btn btn-primary' : 'btn btn-secondary'} aria-pressed={klass.id === activeId} onClick={() => setActiveId(klass.id)}>
              {klass.name}<span>{klass.student_count} siswa · {klass.teacher_count} guru</span>
            </button>)}
          </div>
          {active && <RosterPanel key={active.id} klass={active} users={users} onChanged={load} />}
        </>}
    </section>
  </div>
}

export function AdminManagement({ view, currentUserId }: { view: 'accounts' | 'classes'; currentUserId: string }) {
  const { users, loading, error, load } = useAdminUsers()

  if (loading) return <p role="status">Memuat data administrasi...</p>
  if (error) return <div className="home-data-error" role="alert"><p>{error}</p><button type="button" className="btn btn-secondary" onClick={() => load()}>Coba lagi</button></div>

  if (view === 'accounts') return <AccountsView currentUserId={currentUserId} />
  return <ClassesView users={users} />
}
