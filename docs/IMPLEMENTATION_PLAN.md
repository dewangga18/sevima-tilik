# Tilik: Implementation Plan

## Status

Checkpoint siswa Phase 1–3 sudah tersedia untuk slice kelas 4. Phase 3B (import bank soal) selesai: 65 soal approved aktif, 136 draft tersimpan tanpa ikut selection, 24 kandidat invalid dihapus dari data dan database. Prioritas aktif berikutnya Phase 4: assignment kelas dan insight guru. Arah produk mengikuti `docs/PRODUCT.md`; keputusan dan checkpoint yang belum terbukti tetap perlu review sebelum dilanjutkan. Kontrak endpoint yang tersedia berada di `docs/API_CONTRACT.md`.

Plan ini menjadi sumber eksekusi tunggal, termasuk generator soal, asisten guru, dan usulan scanner tulisan tangan dari plan AI lama. Semua checklist baru masih pending; rencana endpoint bukan klaim fitur sudah tersedia. Scanner direncanakan sebagai koreksi berbantuan review manusia setelah core P0, bukan penilaian akademis otomatis oleh LLM.

Target produk kelas 4–9, tetapi seluruh checkpoint wajib pertama memakai kelas 4. Setiap fase menambah perilaku end-to-end yang bisa didemokan. Checklist belum dicentang sampai bukti checkpoint dicatat di dokumen ini.

## Eksekusi gabungan dalam satu scope

Pengguna mengizinkan fase/slice terkait dikerjakan langsung sebagai satu paket. Fase adalah checkpoint, bukan alasan berhenti: lanjutkan pekerjaan terkait selama requirement, authorization, dan dependency sudah jelas. Selesaikan perubahan engine -> API -> UI -> verifikasi sebagai satu vertical slice.

| Paket | Cakupan yang boleh digabung | Batas |
|---|---|---|
| Diagnostic dan recovery | Phase 2 + loading/error/retry/ownership/reliability yang langsung terkait | Threshold/rubric harus disetujui; jangan menunggu Phase 6 untuk quality dasar |
| Perbaikan gap sampai reassessment | Phase 2–3 ketika rubric, lesson, dan bank variasi sudah siap | Catat checkpoint diagnostic dan learning loop masing-masing; jangan klaim salah satunya selesai tanpa bukti |
| Import konten kelas 4 | Phase 3B: audit -> validasi/dry-run -> draft PostgreSQL -> aktivasi reviewed -> verifikasi siswa | Tidak mengubah graph, mengaktifkan kelas baru, atau menimpa soal yang sudah dipakai secara implisit |
| Kelas dan pengelolaan | Model kelas/assignment Phase 4 + administrasi penempatan Phase 5 | Otorisasi kelas dan insight guru P0 didahulukan; engagement tetap setelah core P0 |
| Draft soal dan asisten | Slice AI guru/admin memakai authorization dan evidence yang sama | Provider/biaya/dependency belum disetujui; jangan mengirim data atau memasang dependency baru secara implisit |
| Scanner dan koreksi batch | Slice AI setelah assignment/ownership tersedia; mulai satu lembar lalu batch dengan review | Hasil OCR/AI tidak langsung menjadi evidence/mastery; kapasitas, penyimpanan foto, dan izin masih harus ditentukan |

Gabungan tidak menghapus prioritas P0 -> P1 -> P2 -> P3 atau keputusan yang masih pending. Commit tetap dipisahkan per perubahan logis.

## Keputusan sebelum implementasi

| Keputusan | Usulan / status | Menghalangi |
|---|---|---|
| Web/backend/runtime | React + TypeScript / Go / Compose sudah disepakati; Vite mengikuti default guide | Tidak |
| Deadline | Belum diberikan; ukuran fase bukan estimasi waktu | Komitmen timeline |
| Database dan driver | PostgreSQL dikonfirmasi pengguna; implementasi memakai pgx yang sudah tersedia | Tidak |
| Auth | Tentukan provider atau implementasi session minimum dan hashing teruji; akun demo siswa/guru, bukan guest sebagai pengganti auth | Phase 1 |
| Slice kelas 4 | Usulan jalur pecahan di bawah, menunggu review | Phase 1 content |
| Bank soal dan rubric | Seed 18 soal tersedia; progressive-demo-v1 memakai tiga jawaban per skill dan cap 18. Ini aturan demo, bukan rubric pendidikan tervalidasi | 6 lesson + 72 soal practice/reassessment tersedia; review pendidikan eksternal pending |
| Desain | Konteks sudah jelas; isi direction dan tokens sebelum UI | Phase 1 UI |
| AI | Generator/asisten dan scanner berbantuan review direncanakan. Gemini adalah kandidat dari plan lama; model yang tersedia, biaya/kuota dan izin data belum dikonfirmasi | Panggilan provider saja; core learning tidak bergantung AI |
| Scanner | Mulai kelas 4, satu lembar dan review guru; batch sesudah checkpoint satu lembar | Konfirmasi kunci, normalisasi jawaban, izin/retensi foto, dan cara menyimpan hasil sebelum implementasi |
| Docker lokal | Colima/Compose menjalankan web, API, dan PostgreSQL; checkpoint learning diuji pada stack ini | Tidak |

Jangan memasang dependency atau scaffold sebelum keputusan yang terkait disetujui. Jangan mengisi keputusan yang belum pasti sebagai fakta.

## Slice kurikulum pertama: usulan kelas 4

Enam skill: perkalian dasar, pembagian dasar, representasi pecahan, pecahan senilai, perbandingan pecahan berpenyebut sama, dan perbandingan pecahan berbeda penyebut.

Graph adalah DAG dengan edge prasyarat eksplisit, bukan satu rantai seluruh matematika:

- Perkalian dan pembagian mendukung pecahan senilai.
- Representasi pecahan mendukung pecahan senilai dan perbandingan berpenyebut sama.
- Pecahan senilai dan perbandingan berpenyebut sama mendukung perbandingan berbeda penyebut.

Kesesuaian kelas dan edge ini perlu review sebelum dipakai. Prasyarat fondasi boleh berasal dari kelas sebelumnya. Jangan menyebut seluruh kurikulum kelas 4 sudah tercakup.

Usulan bank awal: minimal tiga soal berbeda per skill per level yang digunakan (level 1–2); satu micro lesson per skill. Prioritaskan variasi bukti dan soal reassessment yang belum dilihat. Jika soal habis, hentikan dengan status bukti belum cukup, jangan mengulang soal untuk menciptakan mastery palsu. Level 3–5 ditambahkan setelah jalur dasar bekerja.

Setiap soal punya skill ID, topic/domain, difficulty, grade range, tipe, opsi, kunci, penjelasan, dan misconception mapping bila didukung. Kunci dan scoring hanya ada di backend. Pecahan dibandingkan secara eksak, bukan float tanpa toleransi yang jelas.

## Aturan learning engine: spesifikasi wajib sebelum Phase 2

- Versikan aturan pemilihan soal, evidence, mastery, difficulty, dan stopping condition.
- Pisahkan status belum dinilai dari skor rendah. Confidence dan jumlah bukti terlihat pada hasil guru.
- Versi progressive-demo-v1 menggunakan minimum tiga jawaban dan cap 18 dengan label evidence, bukan Mastered. Window reassessment Phase 3 memakai tiga jawaban baru sesuai `docs/LEARNING_RULES.md`; jangan menyamakan evidence diagnostic dengan mastery tervalidasi.
- Satu kesalahan tidak cukup untuk menyatakan root gap. Root gap hanya kandidat setelah soal prasyarat memberi bukti; beberapa kandidat dan hasil inconclusive harus didukung.
- Backtracking memakai visited set, batas langkah, urutan tie-break stabil, serta deteksi cycle pada seed.
- Jawaban salah berulang menurunkan level/memeriksa prasyarat; jawaban benar konsisten menaikkan level dalam rentang konten yang tersedia.
- Mastery diperbarui dari bukti; reassessment bisa menaikkan, mempertahankan, atau menurunkannya. Label Mastered tidak ditentukan oleh XP.
- Request jawaban memiliki identitas unik; retry tidak menggandakan attempt, mastery, atau reward. Reset demo tidak menghapus data pengguna lain.

Simpan tabel input -> pertanyaan berikutnya -> status/bukti yang diharapkan untuk lintasan siswa kuat, gap prasyarat, jawaban campuran, dan bank soal habis sebagai spesifikasi aturan engine.

## Phase 1: Login -> diagnostic singkat -> hasil tersimpan [P0]

Goal: satu siswa kelas 4 menyelesaikan flow web/API/storage, bukan scaffold kosong.

Dependencies: keputusan auth/DB/dependency, review seed awal, desain minimum, Docker lokal siap.

### Langkah 1A: konfigurasi Docker FE + BE terlebih dahulu

Ini pekerjaan implementasi pertama setelah plan dikonfirmasi. Keputusan auth, seed kurikulum, dan visual yang masih pending tidak menghalangi smoke check Docker; persetujuan dependency scaffold tetap diperlukan.

- [x] Pastikan Colima berjalan, context Docker terhubung, `docker compose version` dikenali, dan container smoke test berhasil.
- [x] Scaffold minimum `apps/web` React + TypeScript dan `apps/api` Go; tetapkan versi runtime, package manager/lockfile, serta module Go. Tidak membangun fitur/layer lengkap.
- [x] Buat `apps/web/Dockerfile` dengan target development, build, production; dev server bind `0.0.0.0:5173`. Build production memasukkan public API configuration pada build time; dokumentasikan bahwa perubahan `VITE_*` memerlukan rebuild. Siapkan fallback SPA di server production bila routing client dipakai.
- [x] Buat `apps/api/Dockerfile` dengan target development, build, production; API bind `0.0.0.0:8080` dan menyediakan `GET /health`. Jangan mensyaratkan `go.sum` sebelum module memiliki dependency yang menghasilkan file tersebut. Healthcheck harus memakai executable yang tersedia dalam image.
- [x] Tambahkan `.dockerignore` per app dan `.gitignore` untuk env/secrets, build output, serta dependency lokal; simpan `.env.example` tanpa kredensial nyata.
- [x] Compose root memakai development target, build context independen, port configurable, source bind mount, dan volume dependency web; API healthcheck mengatur readiness. Dokumentasikan restart API secara manual saat source berubah jika belum ada reload tool yang disetujui.
- [x] Browser memakai API URL host-visible dari public env, bukan hostname `api`; komunikasi antarkontainer memakai service name. CORS mengikuti origin dan port frontend yang dikonfigurasi.
- [x] Web menampilkan hasil call `/health` dengan loading dan error state; endpoint ini tidak bergantung pada auth, database, atau provider AI.
- [x] Build development dan production target kedua app secara independen. Jalankan smoke check production container serta `docker compose up --build` untuk dev; verifikasi web -> API, perubahan source frontend, dan restart backend.

Checkpoint 1A: **LOLOS**. `docker compose up --build` menyalakan web dan api; curl `http://localhost:8080/health` menghasilkan `{"status":"ok"}`. Build production target kedua app lolos (`tilik-api-prod` dan `tilik-web-prod`).

### Langkah 1B: persistensi, auth, dan diagnostic awal

Dependencies: checkpoint 1A dan keputusan auth/DB/seed/desain terkait selesai.

- [x] API contract/error shape dan model user/role, grade, skill/prerequisite, soal, assessment, answer, evidence tersedia.
- [ ] Model kelas, enrollment siswa, dan assignment guru belum tersedia; selesaikan sebagai prerequisite Phase 4.
- [x] Tambahkan database yang disepakati ke Compose dengan storage persisten, readiness, dan env examples; pertahankan app build independen dari source app lain.
- [x] Migration dan seed akun siswa/guru fiktif; data progress tidak direset pada startup.
- [ ] Kelas 4A saat ini label UI, belum record database/enrollment; jangan mengklaim seed kelas selesai.
- [x] Login/logout -> session backend -> onboarding kelas 4 -> diagnostic singkat menggunakan soal reviewed.
- [x] Submit jawaban -> server mengevaluasi -> simpan bukti -> tampilkan hasil per skill dan belum dinilai. Progressive selection/backtracking ditambahkan Phase 2.
- [x] Validasi input, session/ownership, CORS, safe errors, serta loading/error/empty states pada flow ini sejak awal.

Demo checkpoint: **LOLOS**.
- Login siswa `budi@tilik.id` -> token didapat -> GET `/api/skills` memuat DAG prasyarat pecahan kelas 4 -> POST `/api/diagnostic/start` memuat 6 soal tanpa bocor answer_key -> POST `/api/diagnostic/submit` mengevaluasi server-side dan menghitung evidence per skill.
- Akun guru `siti@tilik.id` mencoba submit/membaca attempt siswa lain menghasilkan `403 Forbidden` (`{"success":false,"error":"Akses ditolak"}`).
- Token tidak valid ditolak (`401 Unauthorized`).
- `docker compose restart`: container direstart, query `GET /api/diagnostic/latest` membuktikan data attempt dan hasil 6 skill tetap ada (persisten).
- `npm run build` di `apps/web` dan `go build ./...` di `apps/api` keduanya lolos mandiri.

Review hardening Phase 1:
- Penjelasan/misconception disembunyikan pada assessment baru, resume, dan latest selama belum selesai; hasil selesai tetap memiliki penjelasan.
- Completion dan evidence disimpan atomically; dua submit bersamaan menghasilkan satu sukses dan satu `409`, tanpa menimpa jawaban/timestamp. Kegagalan storage me-rollback perubahan.
- Diagnostic handler mengirim pesan aman dan mencatat detail repository di server; restore failure memiliki error/retry terpisah dari hasil kosong.
- Frontend production menerima `VITE_API_URL` melalui build argument; build tanpa konfigurasi ditolak. URL kustom diverifikasi pada bundle image.
- Demo quick-login berbasis role hanya development, tanpa password frontend; production tidak mendaftarkan route demo atau menerima akun/session demo. Seed user demo terpisah dari seed kurikulum.
- Verifikasi: `go test -race ./...` dengan PostgreSQL test/schema terisolasi, `go vet ./...`, `go build ./...`, `npm run build`, `npm run lint`, kedua Docker production builds, HTTP smoke test, serta browser test untuk retry via klik/keyboard dan quick-login. Pemeriksaan ini mencakup review fixes; tidak menggantikan validasi pedagogis, konten, atau fitur fase berikutnya.

Fallback: seed akun demo menghindari kebutuhan registration/password recovery, tetapi tidak menghapus auth. Jika Docker belum siap, build native hanya checkpoint sementara; fase belum selesai sampai checkpoint Docker lolos.

### Penyesuaian Phase 1/P0: beranda siswa sebelum assessment

Permintaan pengguna: login siswa membuka welcoming page, profil/statistik, lalu recent history; pengerjaan dimulai secara eksplisit.

- [x] Beranda menjadi tampilan awal untuk siswa baru maupun siswa yang memiliki attempt aktif/selesai.
- [x] Akun demo siswa baru terpisah dari akun demo lama; jangan menghapus evidence yang sudah ada atau mereset tiap login.
- [x] History dan statistik berasal dari PostgreSQL, dengan ownership, empty/loading/error/retry states.
- [x] Mulai/lanjutkan assessment dan buka hasil history hanya melalui tindakan siswa; kembali ke beranda tetap tersedia.
- [x] Periksa desktop/mobile, keyboard, contrast, build/lint, dan regresi API.

Checkpoint: login siswa baru -> sambutan/profil/statistik nol/history kosong -> mulai -> kerjakan -> hasil -> kembali beranda -> statistik/history diperbarui -> reload tetap beranda dengan data tersimpan. Akun dengan attempt aktif harus memilih lanjutkan; hasil lama tetap dapat dibuka tanpa mengubah evidence.

Bukti checkpoint: build dan lint frontend lolos; Go race tests menggunakan PostgreSQL terisolasi dan `go vet` lolos. Browser memverifikasi alur mulai -> submit -> hasil -> beranda -> buka hasil tersimpan, serta reload attempt aktif tetap beranda. Lebar 320/360/768/1280px tidak overflow; soal dan hasil diuji pada 320px. Error/retry dan statistik selesai diuji dengan respons browser terkontrol; riwayat/ownership/count diuji dengan PostgreSQL nyata. Contrast teks utama/muted/panel/action berada pada 6.19–14.65:1. Data assessment yang dibuat khusus pemeriksaan browser dibersihkan agar akun demo kembali belum mulai.

### Penyesuaian shell guru/admin dan navigasi siswa [Phase 1/P0]

Arahan pengguna: halaman guru/admin memakai dashboard pengelolaan, dengan sidebar dan area kerja; Beranda, Profil belajar, dan Riwayat siswa ditempatkan pada navbar. Scope ini mengubah presentasi/navigasi, bukan menyatakan API kelas/admin/AI sudah tersedia.

Dependencies: shell login dan beranda siswa tersedia; gunakan token desain serta komponen/dependency yang ada. Backend dan kontrak API tidak berubah pada slice ini.

- [x] Pindahkan tiga link siswa ke navbar; hapus navigasi duplikat di konten beranda.
- [x] Link profil/riwayat dapat membuka beranda dari pengerjaan/hasil, menuju bagian yang diminta, tanpa menghapus pilihan jawaban yang belum dikirim.
- [x] Guru memakai sidebar Ringkasan, Kelas & siswa, Kurikulum, Bank soal, dan Asisten analisis; tiap menu membuka konten yang sesuai.
- [x] Admin memakai sidebar Ringkasan, Akun & role, Kelas & penempatan, Kurikulum, dan Bank soal; tidak memakai halaman siswa atau dashboard guru yang sama persis.
- [x] Kurikulum memakai data aktual API dengan loading/error/retry; hilangkan daftar siswa statis dari halaman aktif. Tampilkan status belum tersedia untuk fitur kelas/admin/AI yang belum memiliki backend.
- [x] Desktop memakai sidebar dan topbar; mobile memakai menu yang dapat dibuka/ditutup dengan keyboard dan layar sentuh. Tabel memiliki scroll lokal tanpa menyebabkan overflow halaman.
- [x] Verifikasi navigasi, logout, focus keyboard, fresh/active student, desktop/mobile, serta frontend build/lint.
- [x] FE menangani loading/error/empty/success pada kurikulum/dashboard, login/logout, dan assessment: status proses, disabled control, retry, empty evidence/history, dan konfirmasi hasil tersimpan.

Checkpoint: **LOLOS**.

 login siswa -> tiga link di navbar -> mulai aktivitas -> link Riwayat kembali ke beranda -> lanjutkan dengan pilihan tersimpan. Login guru/admin -> shell dashboard sesuai role -> menu Kurikulum memuat data aktual -> menu lain menampilkan status fitur yang jujur. Menu mobile dapat dibuka lalu memilih halaman dan tertutup kembali.

Fitur data siswa kelas assigned, admin CRUD, dan AI tetap mengikuti Phase 4/5/slice AI; role guard siswa tetap pekerjaan P0 berikutnya. Perubahan shell tidak menggantikan authorization backend.

Bukti: frontend build/lint lolos tanpa warning; browser student/guru/admin pada 320/768/1280px tidak memiliki overflow halaman. Tabel kurikulum scroll lokal, menu mobile menutup setelah navigasi, fokus pindah ke heading. Tiga link siswa berada di header; navigasi dari quiz ke riwayat lalu lanjut mempertahankan pilihan. Respons kurikulum terkontrol memverifikasi loading, error/retry, empty, success; logout sukses menampilkan konfirmasi. API dan permission belum berubah.

### Audit batas role sebelum melanjutkan core flow [P0]

- [x] Seluruh route diagnostic siswa memakai student middleware setelah session validation; teacher/admin ditolak 403 sebelum body/resource diproses. Kontrak API diperbarui.
- [x] Ownership latest/history/detail/submit tetap berlaku; regression guard meliputi lima route dan semua role, smoke HTTP lintas owner menghasilkan 403.

Checkpoint role guard: **LOLOS**. Go race tests, vet/build lolos; HTTP teacher/admin start/submit/latest/history/detail menghasilkan 403, tanpa session 401, siswa latest/history 200, resource tidak ada 404, detail siswa lain 403.

Audit fitur terbaru berada di `docs/FEATURE_AUDIT.md`; perubahan scope AI/admin di bawah tidak melewati checkpoint core learning Phase 2–4.

## Phase 2: Progressive diagnostic -> prerequisite gap -> path [P0]

Goal: differentiator utama sudah tampak pada hasil siswa.

Dependencies: Phase 1, rubric engine disetujui, graph dan bank soal lengkap untuk slice.

Phase 2 memakai `progressive-demo-v1` dari usulan aturan demo yang disampaikan sebelum instruksi melanjutkan: tiga jawaban per skill, strong 3/3, weak 0–1/3, mixed 2/3, maksimal 18 soal. Inventory dan aturan berada di `docs/DIAGNOSTIC_RULES.md`. Bank 3 soal total per skill cukup untuk diagnostic demo dengan fallback level tersedia; target 3 soal per level dan bank reassessment terpisah tetap pekerjaan konten Phase 3, tidak diklaim selesai.

- [x] Validasi graph prasyarat sebelum diagnostic baru: missing references, duplicate skill/edge, self-cycle/cycle ditolak; topological tie-break stabil. Query errors diteruskan, bukan disembunyikan; bank soal kosong tidak membuat attempt kosong.
- [x] Uji structural fixtures DAG bercabang, urutan input berbeda, cycle/missing/duplicate, serta PostgreSQL failed-start tanpa assessment tersimpan.
- [x] Implementasikan service deterministic: pilih soal berdasarkan bukti, grade entry point, difficulty, backtracking, batas assessment, dan confidence.
- [x] Simpan target skill dan provenance jawaban prasyarat; bedakan visible gap, kandidat root gap, misconception candidate, dan belum cukup bukti.
- [x] Bentuk learning path dari prerequisite yang belum dikuasai dengan urutan topologis; jangan mengunci path karena skill belum pernah diukur seolah gagal.
- [x] UI hasil menampilkan alasan rekomendasi, kandidat root gap/provenance, dan urutan review; API tidak mengirim kunci/penjelasan/correctness aktif.
- [x] Tombol belajar membuka micro lesson nyata untuk enam skill melalui slice Phase 3.
- [x] Uji fixture kuat/gap/campuran, graph bercabang/cycle, stop condition, dan exhaustion di dekat service Go.

Demo checkpoint diagnostic -> kandidat gap -> rekomendasi path: **LOLOS**. Lesson/practice belum termasuk checkpoint ini.

Skenario: siswa gagal perbandingan pecahan -> sistem memeriksa pecahan senilai -> evidence menunjukkan kandidat gap -> path mendahulukan lesson pecahan senilai. Siswa kuat mendapat soal lebih menantang yang tersedia; jawaban campuran tidak menghasilkan diagnosis pasti. Replay fixture menghasilkan hasil sama.

Fallback: semua penjelasan memakai template reviewed; AI tidak diperlukan. Jika evidence kurang, hasil inconclusive dengan langkah review berikutnya, bukan root gap buatan.

Bukti: Go race tests dengan PostgreSQL terisolasi memverifikasi strong/mixed/prerequisite gap/exhaustion, cap 18, provenance/path topologis, retry completion, concurrent start/answer (termasuk jawaban berbeda), rollback storage, ownership dan legacy compatibility. Browser memverifikasi target kuat berhenti setelah 3 soal, jalur target/pecahan senilai lemah -> kandidat Perkalian Dasar pada 18 soal, hasil/path tersimpan dapat dibuka dari history. Network failure mempertahankan pilihan dan retry; hasil tidak overflow pada 320/768/1280px. Build/lint frontend dan Go vet/build lolos.

## Phase 3: Lesson -> adaptive practice -> reassessment [P0]

Goal: gap yang ditemukan punya tindakan pembelajaran dan hasil yang bisa diamati.

Dependencies: Phase 2; seed lesson dan bank practice/reassessment terpisah untuk enam skill. Aturan demo/provenance di `docs/LEARNING_RULES.md`; review pendidikan eksternal belum diklaim.

Slice Phase 3 selesai untuk demo kelas 4: schema/seed -> lesson -> tiga practice -> tiga reassessment -> progress/path/current history, dengan satu answer transaction dan UI state lengkap.

- [x] Tambahkan lesson/progress dan sesi practice dengan ownership dan status tersimpan.
- [x] Path membuka micro lesson satu konsep dan practice pada level bukti siswa. Lesson/hint diperiksa secara matematis; review pendidikan eksternal pending.
- [x] Perbarui difficulty dan prerequisite recommendation saat salah berulang; tampilkan feedback tanpa mempermalukan siswa.
- [x] Reassessment menggunakan soal berbeda; update mastery dan path dari evidence, bukan dari klik selesai lesson.
- [x] Home/progress/mastery map mengambil hasil aktual API, termasuk resume saat reload.
- [x] Uji batas status mastery, retry jawaban, dan siswa yang belum membaik.

Demo checkpoint: **LOLOS** — root gap -> lesson -> practice -> reassessment -> progress/path berubah sesuai jawaban dan bertahan setelah refresh. Lintasan tidak membaik tetap menunjukkan kebutuhan review.

Bukti checkpoint: Go race tests dengan PostgreSQL memverifikasi ownership, concurrent start/answer, retry immutable, rollback completion, strong/mixed/declining/unassessed baselines, exhaustion, dan diagnostic yang lebih baru. Browser: root Perkalian Dasar -> lesson -> 3 practice -> 3 reassessment -> 0% menjadi 100% -> path maju ke Pecahan Senilai; reload/resume/history bertahan. Failure answer mempertahankan pilihan; progress loading/error/empty/success dan retry diuji. Lesson/practice/home tanpa overflow 320/768/1280px; detail progress dapat dibuka via keyboard. Web build/lint dan Go vet/build lolos. Konten demo belum review guru eksternal; tidak mengklaim academic Mastered.

## Phase 3B: Bank soal kandidat -> PostgreSQL -> aktivitas siswa [P0]

Goal: konten tambahan dapat dipakai app tanpa mengubah evidence lama atau memasukkan soal yang belum layak ke diagnostic.

Dependencies: Phase 3; audit di `docs/QUESTION_BANK_AUDIT.md`. Enam file `data/processed` memuat 63 kandidat: 45 diagnostic dan 18 practice, tanpa tambahan reassessment. Satu file belum valid JSON; beberapa soal memiliki key/opsi/wording bermasalah. Ini inventory kandidat, bukan 63 soal approved.

- [x] Perbaiki JSON dan review key, opsi ekuivalen, explanation, wording, atribusi skill/difficulty, duplikasi seed serta metadata/provenance; catat keputusan per soal. Jangan mengubah soal bermasalah dengan menebak maksud penulis.
- [x] Buat command importer Go tanpa dependency baru, dapat dijalankan melalui Docker, dengan dry-run dan laporan valid/ditolak/konflik. Dry-run tidak menulis database.
- [x] Validasi ID unik, purpose, grade, difficulty, opsi/key, skill dan metadata. Resolve key A/B/C/D ke teks opsi yang dinilai engine; simpan original option IDs, sumber dan misconception mapping server-side.
- [x] Simpan kandidat/draft di PostgreSQL tanpa ikut selection aktif. Skill baru atau kelas 5–6 tetap draft sampai graph, lesson dan vertical slice terkait tersedia; prerequisite dari file tidak otomatis mengganti graph aktif.
- [x] Import/aktivasi soal reviewed dalam transaksi, idempotent berdasarkan ID/hash. Laporkan konflik isi pada ID lama dan duplikasi prompt; jangan menimpa soal yang telah dipakai pada assessment.
- [x] Aktifkan dahulu konten yang cocok dengan enam skill kelas 4; selection diagnostic/practice/reassessment hanya mengambil soal approved untuk scope yang didukung. Pertahankan bank reassessment terpisah dan exclusion soal yang pernah ditawarkan.
- [x] Seimbangkan posisi opsi tanpa mengubah jawaban tersimpan; bila memakai pengacakan, urutan stabil saat resume/retry attempt yang sama. Kunci tidak masuk bundle frontend.
- [x] Perbarui API contract jika behavior selection berubah; uji dry-run, malformed/key invalid, konflik, import ulang, rollback, draft exclusion, soal lama immutable dan ownership. Jalankan diagnostic -> lesson/practice -> reassessment melalui app setelah import.

Demo checkpoint: **LOLOS**. Dry-run melaporkan alasan penolakan per kandidat; 24 kandidat invalid (opsi ekuivalen dengan kunci, key salah secara matematis, soal ambigu, misconception mapping rusak) dihapus dari `data/processed` dan `question_candidates` setelah audit. Import aktif: 65 approved (diaktifkan ke `questions`, 155 soal runtime: 26 diagnostic, 62 practice, 67 reassessment), 136 draft tersimpan tanpa ikut selection. Import ulang idempotent (seluruh item `unchanged`, hash identik). Verifikasi API: siswa demo baru menjalankan diagnostic dan soal import `diag-cmpdiff-001` ditawarkan dengan opsi ter-rotasi stabil; `go build`/`go vet` dan `npm run build`/`npm run lint` lolos. Review pendidikan eksternal tetap pending; label bukan klaim Mastered akademik.

Fallback: seed lama tetap runnable; kandidat bermasalah tetap draft/rejected dengan alasan eksplisit. Tambahan practice tidak otomatis menambah siklus belajar lengkap jika soal reassessment baru belum cukup. Endpoint upload/admin belum diperlukan untuk command import lokal.

## Phase 4: Data siswa -> insight guru -> intervensi [P0]

Goal: demo learning loop lengkap sampai keputusan guru.

Dependencies: Phase 3 dan checkpoint import Phase 3B yang sudah lolos; tambahkan classroom, student enrollment, dan teacher assignment sebelum endpoint data siswa guru. Seed assignment untuk demo, bukan daftar siswa hardcoded.

- [ ] Relasi assignment mendukung satu guru di beberapa kelas; query data hanya kelas yang di-assign. Uji dua kelas assigned dan satu kelas nonassigned.
- [ ] Role teacher membaca kelas yang dia ajar saja; agregasi membedakan belum dinilai dan gap, menampilkan denominator siswa yang dinilai.
- [ ] Dashboard kelas -> daftar siswa -> detail -> topic/skill mastery, bukti prasyarat, dan perubahan progress.
- [ ] Dashboard completion memakai jumlah siswa/aktivitas eligible dan status aktual; ranking completion opsional setelah aturan denominator/window disepakati.
- [ ] Agregasi visible gap dan root-gap candidate memakai definisi sama dengan hasil siswa; jangan menghitung rata-rata mastery seolah tervalidasi jika datanya belum cukup.
- [ ] Rekomendasi intervensi mengacu pada lesson/skill yang benar-benar tersedia, tanpa assignment otomatis.
- [ ] Seed beberapa pola siswa dengan label data demo untuk memperlihatkan insight kelas; progress siswa yang sedang didemokan harus data aktual.
- [ ] Uji akses student->teacher ditolak dan teacher kelas lain ditolak.

Demo checkpoint: guru membuka kelas 4A -> melihat masalah perbandingan pecahan -> membuka siswa demo yang sama -> melihat kandidat gap pecahan senilai beserta evidence -> mendapat lesson/practice recommendation. Data berubah setelah siswa reassessment.

Fallback: demo utama satu kelas dan satu guru, tetapi model/otorisasi tetap mendukung beberapa kelas; filter lintas sekolah dan assignment ditunda. Jangan mengganti insight guru dengan statistik statis.

## Phase 5: Engagement dan administrasi minimum [P1]

Goal: melengkapi must-have sekunder setelah seluruh core demo P0 selesai.

Dependencies: Phase 4; aturan reward, daily goal, zona waktu, dan role admin disepakati.

- [ ] Reward server untuk diagnostic/lesson/practice/reassessment; pisahkan XP/level dari academic mastery dan cegah duplikasi reward saat retry.
- [ ] Streak berdasarkan hari belajar dalam zona waktu yang ditetapkan, daily goal sederhana, dan achievement awal untuk diagnostic pertama serta peningkatan mastery.
- [ ] Tampilkan XP, level, streak, daily goal, achievement di home/progress dengan data aktual.
- [ ] Admin membuat akun/role serta kelas, mengatur enrollment siswa dan assignment guru (beberapa kelas); izin dan navigasi admin terpisah meskipun shell dipakai bersama. Tidak ada public self-upgrade role.
- [ ] Perubahan assignment berlaku untuk detail, agregasi, dan AI; guru yang dilepas dari kelas kehilangan akses pada request berikutnya.
- [ ] Uji reward retry, batas pergantian hari, dan perubahan role oleh akun tanpa izin.

Demo checkpoint: aktivitas memberi reward sekali -> hari belajar tercatat -> achievement muncul sesuai bukti; admin mengelola akun/enrollment lalu akun itu dapat login sesuai role. Menonaktifkan akun mencabut aksesnya.

Scope cut memerlukan catatan persetujuan karena handover menyebut fitur ini must-have. Seed akun adalah fallback demo untuk management, bukan klaim fitur management selesai.

### Slice AI: konfigurasi dan batas bersama [P1, setelah checkpoint P0]

Dependencies: Phase 4 authorization/evidence dan Phase 3B draft bank. Selaraskan scope scanner dengan `docs/PRODUCT.md` sebelum implementasi. Pilihan provider/model, biaya/kuota, aturan penyimpanan foto dan izin penggunaan data harus ditentukan; dependency baru tetap memerlukan persetujuan. Jangan memakai nama/model lama tanpa memeriksa ketersediaannya saat implementasi.

- [ ] Integrasi provider berjalan di backend Go melalui `net/http`, `context` dan timeout terkonfigurasi; secrets hanya env backend. Jika Gemini dipilih, `GEMINI_API_KEY` berada pada config dan contoh env tanpa nilai rahasia.
- [ ] Tetapkan batas input/output, timeout, retry terbatas dan rate limit sesuai kapasitas provider. Validasi struktur hasil, bukan hanya JSON yang bisa diparse.
- [ ] PII tidak dikirim ke provider. Foto dapat memuat nama atau informasi siswa: siapkan redaksi/crop sebelum pengiriman, permission, retensi/penghapusan, serta larangan foto/base64 masuk log.
- [ ] FE memiliki loading/error/empty/success, timeout dan retry dengan draft tetap tersimpan. Provider tidak tersedia harus ditampilkan jelas; fallback memakai input manual, evidence aktual atau template. Mock demo harus berlabel simulasi dan tidak disimpan sebagai OCR/evidence/nilai nyata.
- [ ] Definisikan kontrak lengkap di `docs/API_CONTRACT.md` sebelum setiap endpoint: method/path, role/owner/assignment/environment, fields, status/error aman, retry/idempotency dan aturan penyimpanan. Usulan route di bawah belum kontrak aktif.

### Slice AI-A: Generator soal dan asisten guru [P1]

- [ ] Guru/admin meminta draft soal kelas 4 dengan skill/difficulty; validasi struktur, kunci, opsi, penjelasan dan metadata, lalu review sebelum publish. Cegah publish otomatis atau bank soal aktif berubah karena output model mentah.
- [ ] Form/modal generator memilih skill, grade, difficulty dan konteks; simpan ke draft PostgreSQL melalui alur Phase 3B, bukan tombol yang langsung menerbitkan soal ke assessment.
- [ ] Guru memilih siswa di kelas assigned -> asisten mengambil evidence yang diizinkan -> menghasilkan analisis dan rekomendasi dengan rujukan evidence. Terapkan authorization pada setiap tool call; minimalkan identitas yang dikirim ke provider.
- [ ] Panel asisten pada dashboard guru membedakan bukti, dugaan misconception dan rekomendasi; materi yang direkomendasikan benar-benar tersedia. Chat/generation tidak mengubah nilai atau mastery.
- [ ] Admin dapat membantu menyiapkan draft untuk guru; jangan menganggap ini izin membaca data akademik semua siswa.
- [ ] Uji akses lintas kelas, revoked assignment, prompt/tool arguments yang mencoba melewati scope, output invalid, retry dan timeout.

Usulan route yang harus didefinisikan sebelum coding: `POST /api/ai/generate-question` untuk draft guru/admin; `POST /api/ai/teacher-assistant` untuk siswa yang boleh diakses guru. Akses admin ke analisis individual tidak diberikan secara implisit.

Checkpoint: guru assigned dua kelas dapat menganalisis siswa keduanya, ditolak untuk kelas lain; generate -> review draft -> publish yang sah; admin membantu draft. Mastery/evidence tidak berubah karena chat atau generation.

### Slice AI-B: Scanner satu lembar -> review koreksi [P1]

Goal: membantu membaca jawaban kertas, dengan hasil dan ketidakpastian yang dapat diperiksa guru. Mulai satu lembar kelas 4 sebelum batch.

- [ ] Model kunci jawaban terkonfirmasi: input manual/paket bank approved atau foto lembar kunci guru -> ekstraksi -> preview -> koreksi/konfirmasi guru. Nomor soal, skill dan versi kunci terikat pada tugas/paket, bukan key bebas yang otomatis memengaruhi mastery.
- [ ] Validasi foto JPEG/PNG/WebP berdasarkan isi dan ukuran; usulan batas awal 5 MB/file dikonfirmasi bersama batas request/dimensi. Preview orientasi/rotasi/crop tersedia sebelum kirim; ZIP bukan kebutuhan slice awal.
- [ ] Ekstrak nomor soal dan jawaban final, termasuk coretan/ralat, foto miring/gelap, tulisan sulit terbaca, nomor terlewat/duplikat/tidak berurutan. Hasil meragukan ditandai dengan teks dan alasan untuk review, bukan warna saja; threshold OCR ditentukan dari fixture, bukan angka confidence yang dianggap tervalidasi.
- [ ] Tampilkan foto/redaksi dan hasil ekstraksi berdampingan pada dashboard guru; guru dapat memperbaiki pembacaan dan mengonfirmasi kunci/jawaban sebelum finalisasi.
- [ ] Evaluasi jawaban dilakukan deterministic setelah ekstraksi. Definisikan normalisasi yang didukung, misalnya pecahan rasional 1/2 = 2/4 = 0,5; jawaban bebas/ambigu atau analisis langkah oleh AI tetap dugaan yang perlu review. Jangan menyerahkan sumber kebenaran score ke LLM.
- [ ] Simpan hasil dengan provenance, versi kunci, status review, koreksi reviewer dan audit retry. Skor sementara/final koreksi kertas dibedakan dari progress diagnostic; integrasi ke mastery memerlukan aturan evidence terpisah yang dikonfirmasi.
- [ ] Guru hanya memproses siswa kelas assigned. Fitur siswa, jika diaktifkan, hanya lembar dirinya dan tidak dapat memilih owner lain; hak admin untuk membantu kunci/draft tidak memberi akses foto/nilai semua siswa.
- [ ] Uji fixture tulisan nyata dengan izin, ambiguity/low confidence, pecahan ekuivalen, MIME/payload invalid, timeout, ownership, kunci diubah dan retry finalisasi. Hasil tanpa provider memakai koreksi manual yang jelas, bukan hasil scan buatan.

Usulan route: `POST /api/ai/extract-answer-key` dan `POST /api/ai/scan-correction`. Tentukan resource tugas/siswa dan lifecycle review dalam API contract; jangan meneruskan contoh payload lama berbasis student_name sebagai identitas/otorisasi.

Checkpoint: guru assigned memilih paket/input kunci -> konfirmasi kunci -> upload satu lembar -> periksa ekstraksi -> perbaiki bagian meragukan -> finalisasi koreksi sekali -> reload melihat hasil/provenance. Guru kelas lain ditolak; core learning tetap bekerja ketika AI gagal.

### Slice AI-C: Batch koreksi dan rekap [P1, setelah AI-B lolos]

- [ ] Tambahkan multi-file upload dan mapping setiap lembar ke siswa/tugas yang diizinkan. Batas jumlah file/total ukuran dan concurrency worker ditentukan dari pengujian kuota; angka 30+ file dan 3–5 worker dari plan lama adalah usulan, bukan jaminan kapasitas.
- [ ] Antrean terkontrol dengan job/item ID, status tiap lembar dan progress tersimpan; reload, retry item gagal dan pembatalan tidak menggandakan biaya atau finalisasi. Uji partial failure tanpa menghilangkan hasil item sukses.
- [ ] Dashboard koreksi: kunci -> upload -> progress -> tabel rekap -> review per lembar. Status membedakan menunggu review, selesai dan gagal; skor rendah tidak disamakan dengan kegagalan OCR.
- [ ] Rekap kelas dan ekspor CSV hanya hasil/kelas yang diizinkan, dengan status review terlihat. Ekspor Excel, upload ZIP dan integrasi buku nilai menjadi perluasan setelah alur CSV/finalisasi terbukti; jangan mengklaim buku nilai sudah tersedia.
- [ ] Uji batch dengan quota/timeout, revoked assignment selama job berjalan, file duplikat, restart worker/reload dan retry. Authorization diperiksa lagi saat proses, akses hasil, finalisasi dan ekspor.

Checkpoint: beberapa lembar siswa assigned -> progress tiap item -> satu item gagal dapat diulang -> review -> rekap/export hasil terkonfirmasi -> reload tidak mengulang item sukses. Hasil batch tidak otomatis mengubah mastery.

## Phase 6: Reliability dan verifikasi demo [P2]

Goal: memperkuat semua flow yang sudah berjalan. Keamanan dasar dan feedback UI tidak menunggu fase ini.

- [ ] Recovery dari network error/reload/double submit, expired session, empty class, dan data parsial; hindari request race mengubah jawaban.
- [ ] Jalankan penuh keyboard, focus, contrast, touch targets, matematika terbaca, reduced motion; siswa mobile/desktop dan guru tablet/desktop.
- [ ] Periksa batas payload, authorization seluruh route, no-secret image/repo, cookie/CSRF sesuai auth, safeguards auth/write endpoints, dan safe logs.
- [ ] CI build web/typecheck sesuai scripts dan Go vet/build; Docker build tiap app dan full-stack smoke check.
- [ ] Dokumentasikan setup clean machine, env, migrations/seed, demo reset terarah, restart Go, dan demo script siswa/guru kelas 4.

Demo checkpoint: clean startup -> full demo -> restart/resume -> common failure/retry tanpa corrupt progress; tidak ada akses silang siswa/kelas. Catat command dan hasil verifikasi, bukan hanya checklist.

## Phase 7: Polish dan perluasan opsional [P3]

Dependencies: Phase 1–6 lolos; konfirmasi masing-masing tambahan sebelum dependency/integrasi.

- [ ] Polish visual/motion yang mendukung pembelajaran; hindari dekorasi yang mengganggu soal.
- [ ] AI hint/explanation dengan input minimum, timeout, validasi output dan fallback reviewed; tidak mengubah scoring. Jika agent diwajibkan, re-plan sesuai kebutuhan judging.
- [ ] Assignment memakai enrollment/authorization yang ada; perintah guru eksplisit.
- [ ] Tambahkan slice kelas 5–6 lalu 7–9 satu per satu dengan seed/rubric reviewed dan checkpoint seperti kelas 4. Skenario aljabar kelas 8 diuji setelah kontennya tersedia.
- [ ] Expo mobile sebagai client API Go setelah web demo stabil; buat plan mobile terpisah, jangan memindahkan backend/domain rules ke client.

Checkpoint tiap tambahan: tunjukkan perubahan end-to-end dengan data nyata dan bukti fallback. Memilih grade 5–9 tidak boleh menampilkan dukungan palsu sebelum kontennya siap; beri informasi cakupan yang jelas.

## Drop First If Time Is Short

1. Expo, leaderboard, animasi dekoratif, assignment, ZIP/Excel/buku nilai dan batch scanner.
2. Jika integrasi AI dipotong, pertahankan draft/manual review dan evidence aktual; jangan menggantinya dengan mock yang menyamar sebagai hasil nyata.
3. Perluasan domain/kelas di luar slice kelas 4 dan question levels 3–5.
4. Dengan persetujuan perubahan MVP: AI eksternal/scanner, achievement tambahan, daily goal, administrasi UI; seed terkontrol sebagai fallback.

Jangan memotong diagnostic evidence, prerequisite check, lesson/practice/reassessment, insight guru, auth/ownership, atau persistence. Deadline pendek berarti re-plan, bukan mengklaim fase selesai tanpa checkpoint.

## Handoff untuk sesi AI berikutnya

Baca `AGENTS.md`, guide yang relevan, `docs/PRODUCT.md`, lalu plan ini. Verifikasi status repo dan keputusan pending. Mulai dari fase tertinggi yang belum selesai, batasi perubahan ke slice aktif, catat checkpoint dan blockers, lalu lanjutkan hanya jika checkpoint lolos. Jangan commit/push/deploy tanpa instruksi yang sesuai.
