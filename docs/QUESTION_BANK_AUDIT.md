# Audit bank soal kandidat

Pemeriksaan 26 September 2026. File belum diimpor ke database atau diaktifkan pada app. Audit ini memeriksa struktur dan beberapa masalah matematika yang terlihat; bukan review lengkap kurikulum.

## Inventory

| File dalam data/processed | Kandidat | Cakupan |
|---|---:|---|
| questions_diagnostic_part1.json | 9 | Kelas 4; perkalian, pembagian, representasi pecahan |
| questions_diagnostic_part2.json | 9 | Kelas 4; pecahan senilai dan perbandingan |
| questions_diagnostic_part3.json | 9 | Kelas 4; tiga skill baru |
| questions_diagnostic_part4.json | 9 | Kelas 5–6; JSON belum valid |
| questions_diagnostic_part5.json | 9 | Kelas 5; operasi pecahan/desimal |
| questions_practice_part1.json | 18 | Kelas 4; tiga skill fondasi yang sudah tersedia |

Total kandidat 63, terdiri dari 45 diagnostic dan 18 practice; belum ada file reassessment baru. Jumlah part4 dihitung setelah mengoreksi typo hanya dalam memori, tanpa mengubah file asli. Ada 15 skill; enam sudah tersedia di app, sembilan memerlukan perluasan kurikulum. Distribusi kelas: 45 kelas 4, 15 kelas 5, 3 kelas 6.

## Temuan yang menghalangi aktivasi

- `questions_diagnostic_part4.json:188`: `"text": "text":` membuat JSON gagal diparse.
- `diag-fm-003`: jumlah kelipatan positif 12 di bawah 50 adalah 120; tidak tersedia pada opsi, sedangkan key menunjuk 84. Explanation berisi dugaan revisi yang saling bertentangan.
- `diag-gcd-002`: dengan pembagian kedua jenis permen sama rata, FPB 24 dan 36 = 12 kantong, sehingga coklat per kantong = 2 (A); key menunjuk 4 (C). Wording pembagian tiap jenis juga perlu diperjelas.
- `diag-gcd-003`: KPK 8, 12, 18 = 72 detik. Ketiga kali setelah 08:00:00 adalah 08:03:36; tidak ada pada opsi.
- `diag-per-003`: hitungan persentase menghasilkan sekitar 57,35%, bukan key 58,5%; jumlah siswa hasil kenaikan juga tidak bulat.
- `diag-frac-add-001`: opsi 1/2 dan 2/4 sama-sama benar, sementara engine multiple choice saat ini hanya mendukung satu jawaban benar.
- `diag-frac_rep-003`: pizza disebut berbeda ukuran; jumlah yang dimakan tidak dapat dibandingkan hanya dari proporsi tanpa memperjelas bahwa yang ditanya adalah proporsi setiap pizza.
- `diag-cmpden-002`: jumlah siswa kelas tidak ditentukan sama; pecahan lebih besar tidak otomatis berarti jumlah siswa lebih banyak.
- `diag-pv-003`: angka ribuan dan satuan tidak ditentukan; explanation mengasumsikan keduanya nol.
- Beberapa level 3 menggabungkan skill tambahan. Tag skill utama dan prasyarat perlu ditinjau agar kesalahan tidak keliru diatribusikan sebagai gap skill tunggal.
- 54 dari 63 key menunjuk opsi A. Posisi opsi perlu diseimbangkan atau diacak konsisten per attempt setelah konten diperbaiki.

Prerequisite pada file tidak boleh langsung mengganti graph aktif: ada ID yang belum tersedia seperti `addition_basic`, `counting`, `division_concept`, dan `penjumlahan_berulang`. Tidak ada lesson untuk sembilan skill baru. Sumber/metadata yang ditandai `metadata_inferred` belum diverifikasi hanya karena menyertakan URL kurikulum.

## Cara integrasi yang disarankan

1. Pertahankan JSON kandidat sebagai sumber konten yang dapat direview di Git; jangan bundel key dalam frontend/public assets.
2. Tambahkan importer backend Go dengan dry-run: validasi JSON, ID, purpose, grade, difficulty, opsi/key, duplikasi, skill dan metadata. Simpan hasil validasi dan alasan penolakan; jangan menganggap semua file processed sudah approved.
3. Map `question` menjadi `prompt`, `skill` menjadi `skill_id`, `options[].text` menjadi opsi runtime. Resolve `correct_answer` dari ID opsi ke teks jawaban, misalnya A menjadi 12. Metadata sumber, original option IDs dan misconception mapping tetap tersimpan server-side.
4. Simpan kandidat yang belum reviewed sebagai draft yang dikecualikan dari selection. Pada tahap awal, aktifkan hanya soal reviewed untuk enam skill kelas 4 yang sudah punya graph dan lesson. Part3–5 tetap draft sampai vertical slice kurikulumnya tersedia.
5. Import approved questions dalam transaksi, idempotent berdasarkan ID/hash konten. Konflik isi pada ID lama harus dilaporkan; jangan menimpa soal yang telah dipakai pada assessment. Deteksi duplikasi prompt dengan seed yang sudah ada.
6. Backend memilih soal approved dari PostgreSQL berdasarkan purpose dan cakupan kelas/skill. Frontend tetap memakai API diagnostic/learning; kunci tetap server-only dan explanation mengikuti status pengerjaan.
7. Pertahankan bank reassessment terpisah. Menambah practice tidak otomatis menambah jumlah siklus lengkap jika reassessment yang belum ditawarkan masih terbatas.

Importer dapat berupa command lokal melalui Docker; endpoint upload/admin baru belum diperlukan untuk memasukkan konten ini. Jika behavior selection/API berubah, perbarui `docs/API_CONTRACT.md` dalam commit yang sama dan uji diagnostic/practice/reassessment setelah import.

Pada audit awal terdapat icon dan plan AI terpisah yang bukan bagian dari import bank soal. Plan AI kini digabungkan ke `docs/IMPLEMENTATION_PLAN.md`; file plan lama sudah dihapus.
