import React from 'react'
import type { User } from '../types'

interface NavbarProps {
  user: User | null
  onLogout: () => void
}

export const Navbar: React.FC<NavbarProps> = ({ user, onLogout }) => {
  return (
    <header className="navbar">
      <div className="navbar-container">
        <div className="navbar-brand">
          <div className="brand-logo">T</div>
          <div className="brand-text">
            <span className="brand-title">Tilik</span>
            <span className="brand-tag">Diagnostic Numeracy</span>
          </div>
        </div>

        {user && (
          <div className="navbar-user">
            <div className="user-info">
              <span className="user-name">{user.name}</span>
              <span className={`role-badge role-${user.role}`}>
                {user.role === 'student' ? 'Siswa Kelas 4' : user.role === 'teacher' ? 'Guru' : 'Admin'}
              </span>
            </div>
            <button className="btn btn-secondary btn-sm" onClick={onLogout}>
              Keluar
            </button>
          </div>
        )}
      </div>
    </header>
  )
}
