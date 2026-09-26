import React, { useState } from 'react'
import type { User } from '../types'

interface LoginViewProps {
  onLogin: (email: string, pass: string) => Promise<void>
  onDemoLogin: (role: User['role']) => Promise<void>
  loading: boolean
  error: string
}

export const LoginView: React.FC<LoginViewProps> = ({ onLogin, onDemoLogin, loading, error }) => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (email && password) {
      onLogin(email, password)
    }
  }

  return (
    <div className="login-card" aria-busy={loading}>
      <div className="login-header">
        <h2>Masuk ke Tilik</h2>
        <p className="login-desc">
          Platform asesmen diagnostik & pemetaan prasyarat konsep numerasi
        </p>
      </div>

      {error && <div className="alert alert-error" role="alert">{error}</div>}

      {loading && <p className="login-progress" role="status">Sedang masuk ke Tilik...</p>}

      <form onSubmit={handleSubmit} className="login-form">
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            className="form-control"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="nama@tilik.id"
            maxLength={254}
            required
            disabled={loading}
          />
        </div>

        <div className="form-group">
          <label htmlFor="password">Kata Sandi</label>
          <input
            id="password"
            type="password"
            className="form-control"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            maxLength={72}
            required
            disabled={loading}
          />
        </div>

        <button type="submit" className="btn btn-primary btn-block" disabled={loading}>
          {loading ? 'Memproses...' : 'Masuk'}
        </button>
      </form>

      {import.meta.env.DEV && <div className="demo-accounts-box">
        <p className="demo-title">Pilih Akun Demo Cepat (Klik Langsung):</p>
        <div className="demo-buttons">
          <button
            type="button"
            className="btn btn-demo"
            onClick={() => onDemoLogin('student')}
            disabled={loading}
          >
            <strong>👦 Budi Santoso</strong>
            <span className="demo-sub">Siswa baru · Kelas 4 SD</span>
          </button>
          <button
            type="button"
            className="btn btn-demo"
            onClick={() => onDemoLogin('teacher')}
            disabled={loading}
          >
            <strong>👩‍🏫 Ibu Siti Rahayu</strong>
            <span className="demo-sub">Guru Matematika</span>
          </button>
          <button
            type="button"
            className="btn btn-demo"
            onClick={() => onDemoLogin('admin')}
            disabled={loading}
          >
            <strong>⚙️ Admin Tilik</strong>
            <span className="demo-sub">Administrator</span>
          </button>
        </div>
      </div>}
    </div>
  )
}
