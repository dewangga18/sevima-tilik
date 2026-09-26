import React from 'react'
import type { User } from '../types'

interface NavbarProps {
  user: User | null
  onLogout: () => void
  loggingOut?: boolean
  onStudentNavigate?: (target: string) => void
}

export const Navbar: React.FC<NavbarProps> = ({ user, onLogout, loggingOut = false, onStudentNavigate }) => {
  return (
    <header className="navbar">
      <div className="navbar-container">
        <div className="navbar-brand">
          <img className="brand-logo" src="/tilik-wordmark.png" alt="Tilik" />
          <div className="brand-text">
            <span className="brand-title">Tilik</span>
            <span className="brand-tag">Diagnostic Numeracy</span>
          </div>
        </div>

        {user?.role === 'student' && <nav className="student-navbar-navigation" aria-label="Navigasi siswa">
          <a href="#welcome-title" onClick={() => onStudentNavigate?.('welcome-title')}>Beranda</a>
          <a href="#student-profile" onClick={() => onStudentNavigate?.('student-profile')}>Profil belajar</a>
          <a href="#recent-history" onClick={() => onStudentNavigate?.('recent-history')}>Riwayat</a>
        </nav>}

        {user && (
          <div className="navbar-user">
            <div className="user-info">
              <span className="user-name">{user.name}</span>
              <span className={`role-badge role-${user.role}`}>
                {user.role === 'student' ? `Siswa kelas ${user.grade_level || 4}` : user.role === 'teacher' ? 'Guru' : 'Admin'}
              </span>
            </div>
            <button className="btn btn-secondary btn-sm" onClick={onLogout} disabled={loggingOut}>
              {loggingOut ? 'Keluar...' : 'Keluar'}
            </button>
          </div>
        )}
      </div>
    </header>
  )
}
