# Student Home: UI Brief

Mode: Operate. Scope: perbaikan alur siswa Phase 1/P0, sesuai permintaan pengguna; tidak menambah lesson, XP, streak, atau mastery palsu sebelum engine terkait tersedia.

Design Read: beranda belajar untuk siswa kelas 4 terlebih dahulu, dengan bahasa ramah yang tetap cocok untuk kelas 5–9. Gunakan hijau gelap dan lime dari screenshot referensi, dengan surface terang untuk membaca matematika pada layar ponsel di lingkungan sekolah/rumah. ENERGY 2 / RHYTHM 2 / MOTION 1.

## Hierarki yang disepakati

Navigasi Beranda, Profil belajar, dan Riwayat berada di navbar; link dari aktivitas kembali ke bagian beranda yang diminta tanpa menghapus pilihan soal.

1. Sambutan personal dan langkah belajar berikutnya. Satu CTA untuk mulai atau lanjutkan cek pemahaman; tidak otomatis membuka soal atau hasil saat login/reload.
2. Profil identitas siswa dan statistik assessment aktual. Nol hanya ketika query berhasil dan memang belum ada aktivitas; kegagalan tampil sebagai data belum tersedia dengan retry.
3. Lima aktivitas terbaru, dengan tanggal/status dan tindakan membuka hasil atau melanjutkan. Empty state menjelaskan riwayat akan muncul setelah siswa mulai.

## Keputusan visual

- Warna utama, text, surface, border, serta radii memakai token yang sudah ada. Lime menandai tindakan belajar pada panel hijau gelap; bagian lainnya tetap tenang.
- Sambutan lapang, profil ringkas, history berupa daftar: variasi mengikuti kebutuhan konten, bukan kartu identik untuk semua informasi.
- Satu family sans sistem untuk keterbacaan form, soal, dan angka tanpa download font tambahan. Font size/spacing memakai token.
- Teks rata kiri; hilangkan styling root/template Vite yang memaksakan center alignment, dark scheme, dan border halaman.
- Tidak menambah foto/ilustrasi/logo/icon pack. Inisial profil berasal dari nama akun; status memakai teks, bukan warna saja.
- Motion hanya feedback hover/focus/loading, bukan animasi masuk semua section. Hormati reduced motion.
- Student baru default adalah profil demo terpisah. Data lama dipertahankan dan setiap profil mempertahankan progress saat login ulang.

## Verifikasi

Desktop/mobile dalam satu inspection batch, lalu maksimal satu batch koreksi. Cek fresh/active/completed/history/error states, tombol start/continue/history/back/retry/logout, keyboard focus, overflow, contrast, serta build/lint. History menunjukkan hasil aktual API. Statistical claims tidak dibuat untuk XP/streak/level yang belum diimplementasikan.

## Delivery gate: PASS

- Hierarchy: welcome -> profile -> history; initial login and active-attempt reload stay on home. Explicit start/back/open controls verified against the running API.
- Real data: completed count and measured skills update after actual submission. PostgreSQL tests verify ownership and total count beyond the five-item history limit. No invented XP/streak/mastery.
- Responsive: screenshots inspected at 1280 and 320px; no overflow at 320/360/768/1280px. Quiz and result also fit 320px.
- Accessibility: Tab reaches named navigation controls, activity heading receives focus, options expose pressed state, targets are 44px minimum, reduced-motion reset is present. Essential palette contrast 6.19–14.65:1.
- Recovery: controlled network failures display unavailable data and retry; retry restores the profile. Empty, active, and completed states checked.
- Validation: frontend build/lint, Go race tests with real isolated PostgreSQL, and Go vet pass.

## Progressive diagnostic Phase 2

ProgressiveQuiz mempertahankan satu tindakan Kirim jawaban, menampilkan jumlah jawaban tersimpan dan cap tanpa total soal palsu, serta mempertahankan pilihan saat network error. Setelah commit, soal berikutnya mendapat fokus. Hasil menampilkan status evidence demo, kandidat prasyarat, dan path topologis dengan penjelasan; tidak menyatakan mastery tervalidasi atau membuat tombol lesson palsu. Browser menguji strong 3 soal dan gap 18 soal, retry, history reopen, serta overflow 320/768/1280px.
