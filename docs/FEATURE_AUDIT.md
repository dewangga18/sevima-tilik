# Audit fitur dan batas akses Tilik

Tanggal: 26 September 2026. Scope berdasarkan arahan pengguna terbaru; pemeriksaan source dan smoke HTTP lokal. Audit diperbarui setelah slice learning Phase 3.

| Role / fitur | Status saat ini | Target |
|---|---|---|
| Siswa: data sendiri, latest/history/statistik | Tersedia: query menggunakan ID session; detail/submit/answer memeriksa owner | Isolasi dan role student enforced; teacher/admin 403 |
| Guru: kelas assigned, lebih dari satu | Tersedia: query classroom hanya untuk assignment guru tersebut; guru dilepas kelas kehilangan akses pada request berikutnya | Otorisasi setiap resource dan query berdasarkan assignment |
| Guru: data siswa/dashboard | Data aktual API: kelas, daftar siswa, evidence, kandidat akar, rekomendasi, progress | Data aktual API, completion dan insight dari evidence |
| Guru/admin: generate soal | Belum tersedia | Draft -> validasi -> review -> publish |
| Guru: asisten analisis seorang siswa | Belum tersedia; kontrak usulan ada di `docs/API_CONTRACT.md` tanpa kode | Ambil evidence authorized -> analisis/rekomendasi; provider pending |
| Ranking completion | Belum tersedia; opsional | Tentukan eligibility/window; bukan ranking nilai/mastery |
| Admin: akun/role/kelas/penempatan | Tersedia untuk siswa kelas 4: buat akun, ubah role, aktif/nonaktif, buat kelas, enroll, assignment; menu Akun & role dan Kelas & penempatan memakai data API | Alur admin terpisah dan permission backend |

## Diagnostic Phase 2

Attempt baru menggunakan progressive-demo-v1: satu soal pending, evidence per jawaban, pemeriksaan prerequisite, kandidat gap dan path review. Legacy results tetap bisa dibuka/diselesaikan. Is_correct dan correct_count tidak tersedia saat pengerjaan aktif. Path kini membuka lesson nyata untuk enam skill kelas 4. Practice dan reassessment terpisah memperbarui progress/path terbaru, sementara snapshot diagnostic tetap utuh.

## Temuan yang perlu ditangani

1. StudentMiddleware kini memvalidasi session lalu role student untuk semua diagnostic routes. Smoke teacher/admin menghasilkan 403; ownership antar siswa tetap wajib.
2. Daftar siswa hardcoded sudah dihapus dari halaman aktif. Guru memakai ManagementDashboard dengan data kelas dari API; admin memakai halaman Akun & role serta Kelas & penempatan yang memanggil endpoint admin. Otorisasi kelas kini terbukti: guru nonassigned mendapat 403 dan guru yang dilepas kelas kehilangan akses pada request berikutnya.
3. Migration memiliki users/sessions, curriculum, assessments/items/evidence, lessons, learning sessions/items/request retries, skill progress, classrooms, enrollments, dan teacher_assignments. Tabel users kini juga punya is_active untuk menonaktifkan akun.
4. Form administrasi admin sudah tersedia dan terhubung ke API: buat akun, ubah role, aktif/nonaktif, buat kelas, enroll/keluarkan siswa, tugaskan/lepas guru. Teacher/admin memakai shell dashboard dengan menu sesuai role.
5. Error kurikulum pada shell baru sudah tampil dengan retry. TeacherView lama tidak dipakai oleh App.tsx.

## Bukti dan batas pemeriksaan

Source: apps/api/cmd/server/main.go, internal/handler/{auth,diagnostic,teacher,admin}_handler.go, internal/service/{diagnostic,learning,teacher,admin}_service.go, internal/repository/{assessment,classroom,admin}_repo.go dan db.go; apps/web/src/components/{ManagementDashboard,AdminManagement}.tsx dan App.tsx.

Smoke HTTP terbaru: tanpa session history 401; student latest/history 200; teacher/admin semua diagnostic route 403; assessment ID tidak ada 404 untuk siswa. Ownership antar siswa telah diuji pada regression PostgreSQL sebelumnya. Tidak dibuat assessment baru untuk audit.

Smoke admin pada stack Compose: anonim 401; guru pada route admin 403; akun siswa kelas 4 baru berhasil dibuat dan dapat login dengan role-nya; email duplikat 409; role/grade/password tidak valid 400; role change berlaku pada request berikutnya tanpa login ulang (token lama 403 di route siswa, 200 di route guru); akun nonaktif membuat token lama 401 dan login baru 403; admin mengubah akun sendiri 409; kelas tidak ada 404; enroll guru sebagai siswa 400; field tak dikenal 400; guru yang dilepas kelas 403 pada kelas tersebut tetapi tetap 200 pada kelas lain. Akun dan kelas uji dibuat untuk verifikasi lalu dihapus.

Urutan: role guard siswa -> core learning P0 Phase 2–3 -> classroom/assignment dan data guru Phase 4 -> administrasi minimum Phase 5 -> engagement (reward/streak/achievement) -> AI P1 setelah provider disetujui. Ranking completion tetap opsional. Build/lint web, Go vet/build, dan browser learning checkpoint telah dijalankan; checkpoint administrasi memakai verifikasi HTTP dan build/lint, bukan browser test interaktif.

## Learning Phase 3

Enam lesson dan 72 soal learning terpisah dari diagnostic tersedia. Session role student/owner diwajibkan pada seluruh learning routes. Score berubah hanya setelah tiga reassessment selesai, bukan setelah membaca lesson atau practice. Progress/history/resume memakai storage aktual dan UI membedakan loading/error/empty/success. Bank awal mendukung paling banyak dua siklus per skill; ketika tidak cukup soal baru, API/UI menjelaskan batasnya tanpa menaikkan progress. Klasifikasi dan konten adalah demo awal; review pendidikan eksternal belum dilakukan.

## Admin Phase 5

Manajemen administrasi minimum tersedia sebagai vertical slice penuh: `AdminService` (validasi email/nama/role/kelas/password, grade dibatasi kelas 4, guard akun sendiri), `AdminHandler` (body 4 KiB, field tak dikenal ditolak, error aman), `AdminMiddleware` (403 `Fitur ini hanya tersedia untuk admin` untuk role lain di semua environment), dan `AdminManagement.tsx` untuk halaman Akun & role serta Kelas & penempatan.

Perubahan role dan status aktif berlaku pada request berikutnya tanpa login ulang karena session divalidasi ulang ke database tiap request. Menonaktifkan akun juga menghapus session, sehingga akses dicabut dua lapis. `grade_level` dikosongkan saat role menjadi guru/admin agar guru tidak diam-diam membawa kelas siswa.

Belum ada penghapusan akun/kelas permanen dan belum ada reset kata sandi. Endpoint AI belum ada kode; hanya kontrak usulan di `docs/API_CONTRACT.md`. Reward, streak, daily goal, dan achievement belum diimplementasikan; aturan XP sudah diputuskan dan tercatat di `docs/IMPLEMENTATION_PLAN.md` Phase 5.


## Update generator AI, 26 September 2026

Generator draft kelas 4 tersedia untuk guru/admin, level 1–2 dan purpose diagnostic/practice/reassessment. Admin mengatur model/key pada Pengaturan AI; key terenkripsi PostgreSQL, tidak dibaca balik API. Model aktif Gemini 3.8 Flash (2.5 ditolak Google untuk akun baru). Local dan Railway lolos generasi nyata serta replay idempoten. Aktivasi tetap review admin; scoring/mastery tidak berubah karena generation. Asisten analisis, scanner/batch, dan penambahan skill kurikulum belum tersedia. Kontrak aktif: `docs/API_CONTRACT.md`.
