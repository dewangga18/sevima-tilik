# Audit fitur dan batas akses Tilik

Tanggal: 26 September 2026. Scope berdasarkan arahan pengguna terbaru; pemeriksaan source dan smoke HTTP lokal. Tidak ada endpoint baru pada perubahan dokumentasi ini.

| Role / fitur | Status saat ini | Target |
|---|---|---|
| Siswa: data sendiri, latest/history/statistik | Tersedia: query menggunakan ID session; detail/submit/answer memeriksa owner | Isolasi dan role student enforced; teacher/admin 403 |
| Guru: kelas assigned, lebih dari satu | Belum tersedia: tidak ada tabel classroom/enrollment/teacher assignment | Otorisasi setiap resource dan query berdasarkan assignment |
| Guru: data siswa/dashboard | Shell dashboard dengan kurikulum aktual; data kelas/completion belum tersedia | Data aktual API, completion dan insight dari evidence |
| Guru/admin: generate soal | Belum tersedia | Draft -> validasi -> review -> publish |
| Guru: asisten analisis seorang siswa | Belum tersedia | Ambil evidence authorized -> analisis/rekomendasi; provider pending |
| Ranking completion | Belum tersedia; opsional | Tentukan eligibility/window; bukan ranking nilai/mastery |
| Admin: akun/role/kelas/penempatan | Shell admin terpisah tersedia; CRUD belum tersedia | Alur admin terpisah dan permission backend |

## Diagnostic Phase 2

Attempt baru menggunakan progressive-demo-v1: satu soal pending, evidence per jawaban, pemeriksaan prerequisite, kandidat gap dan path review. Legacy results tetap bisa dibuka/diselesaikan. Is_correct dan correct_count tidak tersedia saat pengerjaan aktif. Lesson/practice/reassessment belum tersedia; path adalah rekomendasi, bukan aktivitas belajar yang sudah berfungsi.

## Temuan yang perlu ditangani

1. StudentMiddleware kini memvalidasi session lalu role student untuk semua diagnostic routes. Smoke teacher/admin menghasilkan 403; ownership antar siswa tetap wajib.
2. Daftar siswa hardcoded sudah dihapus dari halaman aktif; guru/admin memakai ManagementDashboard dengan kurikulum API dan status jujur untuk data kelas yang belum tersedia. Ini belum bukti authorization kelas.
3. Migration hanya memiliki users/sessions, curriculum, assessments/items dan evidence. Klaim kelas/enrollment sudah selesai di plan dikoreksi.
4. Guru/admin memakai shell dashboard dengan menu sesuai role. Form/action administrasi belum tersedia.
5. Error kurikulum pada shell baru sudah tampil dengan retry. TeacherView lama tidak dipakai oleh App.tsx.

## Bukti dan batas pemeriksaan

Source: apps/api/cmd/server/main.go, internal/handler/auth_handler.go, internal/service/diagnostic_service.go, internal/repository/assessment_repo.go dan db.go; apps/web/src/components/TeacherView.tsx dan App.tsx.

Smoke HTTP terbaru: tanpa session history 401; student latest/history 200; teacher/admin semua diagnostic route 403; assessment ID tidak ada 404 untuk siswa. Ownership antar siswa telah diuji pada regression PostgreSQL sebelumnya. Smoke ini tidak membuktikan assignment guru karena model/endpoint belum ada. Tidak dibuat assessment baru untuk audit.

Urutan: role guard siswa -> core learning P0 Phase 2–3 -> classroom/assignment dan data guru Phase 4 -> admin serta AI P1. Ranking completion tetap opsional. Tidak perlu build untuk perubahan dokumentasi saja.
