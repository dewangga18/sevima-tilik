# Tilik: Product Scope

## Status dan sumber keputusan

Arah produk bersumber dari Product Handover pengguna dan instruksi untuk memprioritaskan kelas 4. Detail slice kurikulum dan aturan numerik di implementation plan masih usulan untuk ditinjau, bukan keputusan pedagogis yang sudah tervalidasi.

Panduan proses tetap berada di `guides/PRD.md`. Rencana eksekusi berada di `docs/IMPLEMENTATION_PLAN.md`.

## Chosen Direction

Diagnostic learning dan adaptive numeracy untuk kelas 4 SD sampai kelas 9 SMP. Web dibangun lebih dahulu; konten dan demo kelas 4 diselesaikan sebelum perluasan kelas lain.

## Problem Statement

Kesalahan pada materi tertentu dapat berasal dari prasyarat yang belum dipahami. Nilai total saja tidak cukup untuk menentukan konsep yang perlu diperbaiki.

## Target User

- Utama: siswa kelas 4 untuk MVP pertama; target produk tetap kelas 4–9.
- Pendukung: guru yang perlu melihat bukti learning gap dan intervensi berikutnya.
- Admin: membuat akun/role dan kelas, mengatur siswa serta guru pada kelas, dan membantu menyiapkan soal; school management lengkap bukan prioritas.

## Batas akses dan fitur per role

Arahan pengguna: siswa hanya mengakses data dirinya; guru hanya mengakses siswa dalam kelas yang ditugaskan kepadanya, termasuk beberapa kelas; admin mengelola akun/role, kelas, serta penempatan siswa/guru.

- Siswa: profil, statistik, recent test, jawaban, hasil, dan progress milik sendiri. Identitas pemilik ditentukan session backend, bukan ID dari client.
- Guru: daftar kelas yang di-assign, daftar/detail siswa pada kelas tersebut, dashboard completion dan insight. Semua detail, agregasi, export, dan tool AI harus memeriksa assignment di backend. Tidak ada akses semua siswa hanya karena role teacher.
- Admin: review bank soal (kandidat draft, approve/reject dengan alasan). Prosesnya deterministik dan tidak bergantung AI. Akses guru ke bank soal sengaja tidak diberikan pada versi ini; ruang lingkup awal "guru dan admin" untuk draft soal ditunda sampai generator soal benar-benar ada.
- Asisten guru: chat/brainstorming analisis seorang siswa yang boleh diakses, mengambil evidence lalu menyiapkan rekomendasi terstruktur. Jawaban membedakan evidence dan dugaan; tidak mengubah nilai atau menerbitkan soal otomatis. Provider/biaya belum dipilih.
- Admin: membuat akun dan menentukan role, membuat kelas, mengatur enrollment siswa dan assignment guru, membantu guru menyiapkan soal. Hak akses data akademik individual lintas kelas untuk admin belum ditentukan; jangan memberi akses global secara implisit.
- Ranking completion masih opsi, bukan kewajiban. Dashboard completion didahulukan; completion tidak disamakan dengan academic mastery dan ranking tidak ditampilkan publik kepada siswa.

Status implementasi dan gap dicatat di `docs/FEATURE_AUDIT.md`. Daftar ini adalah scope tujuan, bukan klaim seluruh fitur tersedia.

## Core Demo Flow

Login siswa kelas 4 -> diagnostic -> gagal pada skill target -> pertanyaan prasyarat -> dugaan root gap dengan bukti -> learning path -> micro lesson -> adaptive practice -> reassessment -> progress tersimpan -> guru melihat perubahan siswa yang sama dan rekomendasi intervensi.

Demo kelas 8/aljabar dari handover menjadi skenario perluasan setelah demo kelas 4 lolos.

## Must-Have Features

- Authentication dan grade onboarding.
- Progressive diagnostic, skill graph, pemeriksaan prasyarat, hasil berbasis bukti.
- Learning path, micro lesson, adaptive practice, reassessment, mastery map/progress.
- Teacher dashboard: kelas, daftar/detail siswa, topic/skill mastery, visible gap, dugaan prerequisite gap, rekomendasi intervensi.
- XP, level, streak, achievement, daily goal sebagai engagement setelah learning loop berfungsi.
- Manajemen siswa/guru dan role minimum tanpa school management lengkap.

## AI Agent Actions

Arahan 26 September 2026: fokus generator soal memakai Gemini, local sampai Railway; asisten analisis dan penambahan skill kurikulum ditunda. Generator tersedia untuk guru dan admin; review/publish tetap admin. Asisten tersedia untuk guru pada siswa assigned dan admin pada siswa yang terdaftar di kelas. Arahan terbaru secara eksplisit memberi admin akses evidence individual melalui fitur analisis ini. Provider key disimpan terenkripsi di PostgreSQL, dapat diganti/dihapus dari halaman admin, dan tidak dikembalikan oleh API. Kunci enkripsi storage tetap env backend. Scanner/batch masih di luar slice aktif.

Core scoring, grading, graph traversal, mastery, dan keputusan path bersifat deterministic. AI dapat membantu hint, penjelasan, dan contoh alternatif; kegagalan AI tidak boleh memutus pembelajaran.

Asisten analisis guru dan generate draft soal masuk scope arahan terbaru. Dua aksi asisten: mengambil bukti skill siswa yang diizinkan dan menyiapkan rekomendasi/intervensi terstruktur. Agent tidak diberi wewenang mengubah nilai, role, atau mengeksekusi assignment tanpa persetujuan. Kebutuhan dan provider harus dikonfirmasi sebelum implementasi integrasi.

## Nice to Have

AI explanation untuk siswa, assignment aktivitas individu/kelompok/kelas, filter tambahan, ranking completion, Expo mobile, serta cakupan kurikulum lebih luas.

## Out of Scope

Shop, coins economy, avatar marketplace, guild, chat sosial antarpengguna, multiplayer, parent social network, LMS lengkap, homework solver, dan penilaian akademis otomatis oleh LLM.

## Success Criteria

- Jawaban siswa memengaruhi pertanyaan berikutnya dan path; hasil bukan tampilan statis.
- Bukti prasyarat membedakan visible gap dari dugaan akar gap; bukti kurang tidak diberi label pasti.
- Lesson -> practice -> reassessment memperbarui progress, termasuk ketika siswa belum membaik.
- Guru melihat data siswa yang sama, dengan akses terbatas ke kelasnya.
- Docker menjalankan aplikasi; data bertahan saat restart; learning loop tetap bekerja tanpa AI.

Ini kriteria demonstrasi perilaku sistem, bukan klaim diagnostic accuracy atau efektivitas pendidikan yang sudah terbukti.

## Constraints

- Stack disepakati: React + TypeScript web, Go HTTP API, Docker Compose. Colima adalah runtime lokal macOS.
- React Native + Expo hanya setelah web demo selesai. Next.js/Supabase backend/Drizzle dari suggested stack handover tidak menggantikan kesepakatan ini.
- PostgreSQL sudah dikonfirmasi pengguna sebagai database. Dependency baru dan provider AI tetap perlu persetujuan sebelum pemasangan.
- Deadline dan bank soal tervalidasi belum diberikan. Jangan menjanjikan seluruh domain atau kelas 4–9 selesai dalam hackathon.
- UI friendly, modern, tidak childish; siswa mobile-first, guru desktop/tablet. Arah palette/font mengikuti guides/DESIGN_DIRECTION.md dan guides/DESIGN_SYSTEM.md.
- Data demo memakai identitas fiktif yang diberi label; jangan mengirim data identitas anak ke provider AI.
