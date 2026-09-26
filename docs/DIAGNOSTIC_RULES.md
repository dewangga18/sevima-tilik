# Phase 2: kesiapan diagnostic progresif

Status: instruksi melanjutkan Phase 2 menggunakan usulan aturan demo yang sudah disampaikan. Versi `progressive-demo-v1`: threshold 3 jawaban, cap 18. Aturan ini digunakan untuk attempt baru; legacy-v1 tetap memakai aturan sebelumnya. Ini bukan rubric pendidikan tervalidasi.

## Inventory aktual

PostgreSQL berisi 18 soal untuk 6 skill. Semua soal berbeda ID; setiap skill punya 3 soal total.

| Skill | Level 1 | Level 2 |
|---|---:|---:|
| mul_basic | 2 | 1 |
| div_basic | 2 | 1 |
| frac_rep | 2 | 1 |
| frac_equiv | 1 | 2 |
| frac_cmp_same_den | 2 | 1 |
| frac_cmp_diff_den | 0 | 3 |

Ini belum memenuhi target plan 3 soal berbeda per skill per level yang digunakan. Belum ada bank terpisah untuk reassessment. Jangan menurunkan difficulty ke level yang tidak memiliki soal atau mengulang soal agar evidence terlihat cukup.

## Aturan demo awal

Minimum 3 jawaban per skill; 3 benar memberi bukti kuat, 0–1 benar memberi indikasi perlu penguatan, 2 benar masih belum pasti. Maksimal 18 soal; kandidat root gap memerlukan evidence langsung pada prasyarat. Angka ini hanya untuk demo awal; perubahan berikutnya harus memakai versi aturan baru. Academic mastery jangka panjang tidak boleh disimpulkan dari label demo ini.

Entry/tie-break/level fallback dan stop conditions untuk versi ini didefinisikan pada bagian Selection. Confidence maksimal medium untuk tiga jawaban; tidak mengklaim confidence pendidikan tinggi. Exhaustion menghasilkan insufficient evidence, bukan gagal otomatis.

## Acceptance fixtures draft

| Kasus | Perilaku wajib |
|---|---|
| Siswa kuat | Variasi soal tersedia; difficulty naik hanya jika level berikutnya punya konten; tidak mengulang soal |
| Lemah pada perbandingan berbeda penyebut | Uji prasyarat pecahan senilai/perbandingan penyebut sama sebelum menyimpulkan kandidat root gap |
| Bukti campuran | Hasil inconclusive ketika ambang belum terpenuhi; langkah review tanpa diagnosis pasti |
| Prasyarat belum pernah diuji | Status unassessed; tidak dianggap gagal atau menjadi alasan locked path |
| Bank habis | Stop dengan insufficient evidence; jangan membuat score/mastery palsu |
| DAG bercabang | Kandidat gap dan provenance terkait branch yang benar; tie-break deterministik |
| Graph cycle | Validasi seed menolak cycle, runtime bounded; tidak loop |
| Retry/reload | Jawaban committed tidak berganda; resume mempertahankan next question dan evidence |
| Dua submit bersamaan | Hanya satu transisi committed; tidak mengganti question/evidence dua kali |
| Akses siswa lain | 403 sebelum jawaban, next question, atau path bocor |

## Urutan slice implementasi

1. Finalisasi rubric dan versi aturan; review bank/graph untuk skenario kelas 4.
2. Definisikan kontrak start/answer/next/completion dan idempotency di API_CONTRACT sebelum endpoint berubah.
3. Implementasikan selection/evidence/backtracking dan persistensi satu jawaban -> satu next question.
4. Hubungkan quiz React dengan loading/error/retry/stop; jangan menampilkan future questions/kunci sebelum waktunya.
5. Hasil menyertakan provenance/kandidat gap dan path topologis; aksi menuju lesson baru aktif ketika lesson tersedia.
6. Uji fixtures di service dengan PostgreSQL serta replay full demo di browser; baru tandai checkpoint Phase 2 selesai.

## Slice struktural yang sudah berjalan

GetSkills memvalidasi graph; diagnostic baru menolak graph invalid dan bank soal kosong sebelum membuat attempt. Resume attempt yang sudah ada dipertahankan. Validasi menghasilkan urutan topologis stabil untuk traversal berikutnya, dan kini dipakai oleh adaptive selection serta learning path progressive-demo-v1.

Regression tests mencakup DAG bercabang, urutan input berbeda, missing/duplicate/self-cycle/cycle, dan PostgreSQL invalid curriculum/empty bank tanpa assessment baru. Penanganan rows/scan errors di repository meneruskan error agar handler mengirim pesan aman.

## Selection progressive-demo-v1

Entry kelas 4: perbandingan pecahan berbeda penyebut. Kumpulkan tiga jawaban berbeda per skill. Setelah evidence lemah, telusuri prasyarat dengan urutan skill ID stabil, depth-first; skill kuat berhenti pada branch tersebut. Inconclusive tidak menjadi root gap dan tidak memicu diagnosis pasti. Difficulty mendekati level berikutnya setelah benar dan sebelumnya setelah salah, menggunakan soal yang tersedia dan belum digunakan; tie-break distance/level/question ID. Tidak ada duplikasi soal dalam attempt.

Skill weak menjadi kandidat root gap hanya jika berada pada rantai weak dari target dan semua prasyaratnya sudah memberi evidence kuat; leaf weak tanpa prasyarat juga kandidat. Mixed/unknown prerequisites tidak diberi root label. Path memuat skill weak pada urutan topologis; inconclusive mendapat rekomendasi review dengan label bukti belum cukup, bukan locked skill. Bank exhaustion/cap memberi stop reason eksplisit; tidak menaikkan mastery.
