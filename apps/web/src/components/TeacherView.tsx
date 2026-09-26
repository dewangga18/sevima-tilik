import React, { useEffect, useState } from 'react'
import type { Skill, User } from '../types'
import { api } from '../services/api'

interface TeacherViewProps {
  user: User
}

export const TeacherView: React.FC<TeacherViewProps> = ({ user }) => {
  const [skills, setSkills] = useState<Skill[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getSkills()
      .then(setSkills)
      .catch((err) => console.error(err))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="teacher-container">
      <div className="teacher-header-card">
        <h2>Dashboard Guru — Kelas 4A (Numerasi)</h2>
        <p className="teacher-sub">
          Selamat datang, {user.name}. Pantau pemahaman konsep dan gap prasyarat siswa kelas 4.
        </p>
      </div>

      <div className="teacher-grid">
        {/* Class overview */}
        <div className="teacher-card">
          <h3>Siswa Terdaftar di Kelas 4A</h3>
          <div className="student-list">
            <div className="student-row">
              <div className="student-meta">
                <span className="student-avatar">👦</span>
                <div>
                  <strong>Budi Santoso</strong>
                  <span className="student-sub">Siswa Demo Aktif • Kelas 4 SD</span>
                </div>
              </div>
              <span className="badge badge-active">Sedang Mengerjakan / Terdata</span>
            </div>
            <div className="student-row">
              <div className="student-meta">
                <span className="student-avatar">👧</span>
                <div>
                  <strong>Ani Wijaya</strong>
                  <span className="student-sub">Siswa • Kelas 4 SD</span>
                </div>
              </div>
              <span className="badge badge-pending">Belum Tes</span>
            </div>
            <div className="student-row">
              <div className="student-meta">
                <span className="student-avatar">👦</span>
                <div>
                  <strong>Deni Pratama</strong>
                  <span className="student-sub">Siswa • Kelas 4 SD</span>
                </div>
              </div>
              <span className="badge badge-pending">Belum Tes</span>
            </div>
          </div>
        </div>

        {/* Skill DAG map */}
        <div className="teacher-card">
          <h3>Domain & Peta Keterampilan (Grade 4 Slice)</h3>
          {loading ? (
            <p>Memuat kurikulum...</p>
          ) : (
            <div className="curriculum-skills-list">
              {skills.map((s) => (
                <div key={s.id} className="curriculum-item">
                  <div className="curriculum-item-top">
                    <strong>{s.name}</strong>
                    <span className="domain-tag">{s.domain} (Kelas {s.grade_level})</span>
                  </div>
                  <p className="curriculum-desc">{s.description}</p>
                  {s.prereqs && s.prereqs.length > 0 && (
                    <div className="prereq-tags">
                      <span className="prereq-label">Prasyarat:</span>
                      {s.prereqs.map((pr) => (
                        <span key={pr} className="prereq-badge">{pr}</span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
