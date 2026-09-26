# Demo Script — Tilik Kelas 4

Panduan menjalankan demo end-to-end untuk siswa dan guru. Stack acuan: Docker Compose dari root repository.

## Persiapan (clean machine)

1. Prasyarat: Docker CLI, Docker Compose v2 (macOS + Colima: `colima start`).
2. Salin env dan nyalakan stack:

```bash
cp .env.example .env
docker compose up --build -d
```

3. Tunggu sehat: `curl http://localhost:8080/health` → `{"status":"ok"}`. Migrasi dan seed berjalan otomatis saat API start.
4. Buka `http://localhost:5173`.

Restart API setelah ubah source Go (tidak ada hot reload):

```bash
docker compose restart api
```

## Akun demo (development saja)

Login cepat memakai tombol akun demo (endpoint `POST /api/auth/demo-login`, aktif hanya saat `APP_ENV=development`):

- Siswa baru: Budi Santoso (`budi.baru@tilik.id`)
- Siswa dengan riwayat: Budi Santoso (`budi@tilik.id`)
- Guru: Ibu Siti Rahayu (`siti@tilik.id`) — mengajar 4A dan 4B
- Guru kedua: Pak Rahmat (`rahmat@tilik.id`) — hanya 4B (uji batas akses)
- Admin: Admin Tilik

## Alur demo siswa (belajar dari gap)

1. Login siswa baru → beranda menyapa, statistik nol, riwayat kosong.
2. Mulai diagnostic → jawab soal (jawaban lemah memicu pemeriksaan prasyarat, maksimal 18 soal).
3. Hasil: kandidat akar gap (mis. Perkalian Dasar) + rekomendasi review.
4. Buka micro lesson → 3 latihan adaptif → 3 soal cek ulang (reassessment).
5. Kembali ke beranda → progress/path diperbarui; reload tetap tersimpan.

## Alur demo guru (insight kelas)

1. Login guru (Siti) → dashboard → menu **Kelas & siswa**.
2. Kartu kelas menampilkan "1 terdata dari 4 siswa" (4A) dan 4B kosong.
3. Roster membedakan *Belum dinilai / Sedang mengerjakan / Terdata* + jumlah gap.
4. Klik **Lihat insight** pada siswa lemah → evidence per skill, kandidat akar gap, rekomendasi lesson/practice, perubahan progress.
5. Siswa reassessment → insight guru berubah pada pembukaan berikutnya (agregasi membaca assessment completed terbaru).

## Verifikasi batas akses (opsional, cepat)

```bash
# Token siswa ke route guru → 403
# Guru Rahmat ke kelas 4A → 403 Akses ditolak
```

## Reset data demo terarah

Hapus data seorang siswa tanpa menyentuh pengguna lain (contoh siswa demo baru):

```bash
docker compose exec db psql -U tilik -d tilik_db -c "
DELETE FROM assessments WHERE student_id = 'u-student-new';
DELETE FROM skill_progress WHERE student_id = 'u-student-new';"
```

Jangan memakai `docker compose down -v` untuk reset demo — itu menghapus seluruh database.

## Catatan

- Konten numerasi dan aturan skor adalah versi demo (`progressive-demo-v1`, `learning-demo-v1`), bukan rubrik tervalidasi.
- Produksi menonaktifkan akun/tombol demo; lihat `guides/DEPLOYMENT.md` bagian Production Deployment.
