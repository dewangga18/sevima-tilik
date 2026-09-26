# Tilik API Contract

Dokumen ini menjelaskan endpoint yang sudah diimplementasikan untuk Phase 1. Phase 2 menambah progressive diagnostic dan rekomendasi path; Phase 3 menyediakan lesson/practice/reassessment dan progress siswa; intervensi guru belum tersedia. Ubah dokumen ini bersama perubahan endpoint sesuai `AGENTS.md`.

## Transport dan session

- Base URL berasal dari konfigurasi publik frontend. Default development Compose: `http://localhost:8080`.
- JSON request memakai `Content-Type: application/json`.
- Endpoint terlindungi menerima `Authorization: Bearer <token>` atau cookie `tilik_session`. Header Bearer diprioritaskan jika keduanya dikirim.
- Session berlangsung tujuh hari, diperiksa backend, dan dicabut saat logout. Cookie login memakai HttpOnly, SameSite=Lax, dan Path=/.
- Origin browser yang diizinkan berasal dari `ALLOWED_ORIGIN`. Preflight `OPTIONS` dikembalikan `204` oleh middleware CORS.
- Production image frontend wajib menerima `VITE_API_URL` saat build; env nginx runtime tidak mengubah bundle.

## Bentuk response

Endpoint `/api/*` sukses:

```json
{"success": true, "data": {}}
```

Endpoint `/api/*` terdaftar yang gagal:

```json
{"success": false, "error": "Pesan aman untuk pengguna"}
```

`data` dapat berupa object atau array. Saat latest assessment belum ada, response adalah `{"success":true}` karena `data` nil dihilangkan serializer. Client menormalisasikan data yang tidak ada menjadi `null`; keadaan ini berbeda dari kegagalan HTTP/network.

Route yang tidak terdaftar, termasuk demo-login di production, memakai respons router default `404` dengan plain text `404 page not found`, bukan envelope JSON.

Semua endpoint terlindungi dapat mengembalikan `401` jika session tidak ditemukan, tidak valid, atau expired. Database/repository errors tidak dikirim sebagai detail ke client diagnostic; detail dicatat di server. `500` adalah kegagalan operasi, bukan hasil assessment kosong.

## Ringkasan endpoint

| Method | Path | Akses | Success |
|---|---|---|---|
| GET | `/health` | Publik | 200 |
| POST | `/api/auth/login` | Publik | 200 |
| POST | `/api/auth/demo-login` | Development saja | 200 |
| POST | `/api/auth/logout` | Token/cookie opsional | 200 |
| GET | `/api/auth/me` | Session | 200 |
| GET | `/api/skills` | Publik, seeded curriculum | 200 |
| POST | `/api/diagnostic/start` | Session siswa | 200 |
| POST | `/api/diagnostic/answer` | Session siswa + owner attempt | 200 |
| POST | `/api/diagnostic/submit` | Session siswa + owner attempt | 200 |
| GET | `/api/diagnostic/latest` | Session siswa, hanya milik user | 200 |
| GET | `/api/diagnostic/history` | Session siswa, hanya milik user | 200 |
| GET | `/api/diagnostic/{id}` | Session siswa + owner attempt | 200 |
| GET | `/api/learning/progress` | Session siswa, data sendiri | 200 |
| GET | `/api/learning/lessons/{skill_id}` | Session siswa | 200 |
| POST | `/api/learning/start` | Session siswa + owner diagnostic source | 200 |
| GET | `/api/learning/sessions/{id}` | Session siswa + owner | 200 |
| POST | `/api/learning/lesson-complete` | Session siswa + owner | 200 |
| POST | `/api/learning/answer` | Session siswa + owner | 200 |

Semua route `/api/diagnostic/*` memerlukan session dengan role `student`. Session tidak valid menghasilkan `401`; role teacher/admin menghasilkan `403` dengan pesan `Fitur ini hanya tersedia untuk siswa`, sebelum request body atau resource diproses. Ownership tetap wajib: siswa tidak dapat membaca/submit attempt siswa lain (`403 Akses ditolak`). Pembatasan berlaku di seluruh environment. Retry setelah `401/403` tidak melakukan perubahan; login dengan akun yang berhak sebelum mencoba lagi. Tidak ada endpoint untuk membaca assessment milik siswa lain atau teacher analytics pada fase ini.

## GET /health

Tidak memakai envelope API:

```json
{"status":"ok","env":"development"}
```

Endpoint tidak memanggil AI. Database diinisialisasi saat startup server; health response sendiri tidak menjalankan query database.

## POST /api/auth/login

Request:

```json
{"email":"user@example.test","password":"<password pengguna>"}
```

- `email` dan `password` wajib nonempty; email di-trim.
- `200`: `data` berisi `user` dan `token`, serta cookie session.
- `400`: JSON invalid atau email/password kosong.
- `401`: kredensial salah. Akun demo tetap ditolak di luar development meskipun record lama masih ada.
- `500`: kegagalan operasi dengan pesan umum.
- Login ulang membuat session baru; bukan operasi idempotent.

User response:

```json
{"id":"user-id","email":"user@example.test","name":"Nama","role":"student","grade_level":4,"created_at":"2026-09-26T00:00:00Z"}
```

`role`: `student`, `teacher`, atau `admin`. `grade_level` tidak disertakan jika nol. Password/hash tidak diserialisasi.

## POST /api/auth/demo-login

Request:

```json
{"role":"student"}
```

- Hanya didaftarkan ketika `APP_ENV=development`. Di production dan environment lain endpoint menghasilkan `404`.
- `role` harus `student`, `teacher`, atau `admin`; backend memilih ID akun demo dari allowlist tetap. Client tidak memilih user ID/email/password.
- `200`: bentuk response/cookie sama dengan login biasa.
- `400`: JSON invalid atau role tidak diizinkan.
- `401`: akun demo tidak tersedia atau role stored tidak cocok.
- `500`: kegagalan session/storage dengan pesan aman.
- Setiap request membuat session resmi baru. Session demo tidak diterima di luar development, termasuk session yang sudah tersimpan sebelum environment berubah.
- Seed akun demo hanya development, memakai password acak yang tidak dibagikan. Frontend tidak memiliki password demo; tombol demo hanya tampil dalam Vite development.
- Role demo student memakai profil siswa baru `u-student-new`. Profil lama `u-student-1` dan hasilnya tetap tersimpan; login ulang tidak mereset kedua profil. Siswa baru benar-benar belum memiliki assessment pada seed awal, bukan hasil selesai yang disembunyikan frontend.

## POST /api/auth/logout

Tidak membutuhkan body. Token/cookie yang tersedia dipakai untuk menghapus session, cookie expired dikirim, dan response sukses berisi:

```json
{"success":true,"data":{"message":"Logged out successfully"}}
```

Logout berulang diperbolehkan. Saat ini handler tidak meneruskan kegagalan repository logout sebagai error response; ini batas perilaku yang belum diperkuat, bukan jaminan pencabutan saat storage gagal. Client menghapus token lokal.

## GET /api/auth/me

`200`: `data` adalah object User. `401`: tidak ada session valid, termasuk session akun demo pada production.

## GET /api/skills

`200`: `data` array Skill dari slice yang sudah di-seed, termasuk fondasi kelas sebelumnya:

```json
{"id":"frac_equiv","name":"Pecahan Senilai","domain":"Pecahan","grade_level":4,"description":"Deskripsi konsep","prereqs":["mul_basic","div_basic","frac_rep"]}
```

`prereqs` dapat dihilangkan jika kosong. Saat ini endpoint tidak menyediakan filter grade atau seluruh kurikulum kelas 4–9. `500`: `Gagal memuat daftar skill`, termasuk graph dengan cycle atau referensi prasyarat invalid. Detail graph/repository hanya dicatat server. Urutan skill/prasyarat stabil; request GET aman diulang.

## POST /api/diagnostic/start

Request opsional:

```json
{"grade_level":4}
```

- Body kosong diperbolehkan. JSON invalid/field tak dikenal/payload >16 KiB menghasilkan 400. Grade default dari user, lalu 4 jika nol; grade selain 4 menghasilkan 400 `Diagnostic saat ini tersedia untuk kelas 4`.
- Attempt baru memakai `rule_version=progressive-demo-v1`, `target_skill_id=frac_cmp_diff_den`, `revision=0`, satu item pending. Maksimal 18 jawaban; soal berikutnya dipilih backend. Tidak mengklaim dukungan grade lain.
- `200`: Assessment `in_progress` milik user; latest attempt yang masih berlangsung dilanjutkan. Jika latest sudah selesai, dibuat attempt baru.
- Pembuatan attempt baru memvalidasi graph prasyarat dan ketersediaan soal terlebih dahulu; graph invalid atau bank kosong tidak membuat attempt kosong. Resume attempt lama tetap tersedia tanpa membuat attempt baru.
- `500`: `Gagal memulai tes diagnostik. Silakan coba lagi.`, termasuk graph invalid/bank kosong; error internal tidak dikirim ke client.
- Start/resume tidak mengubah jawaban. Start paralel untuk siswa sama diserialisasi: keduanya kembali ke attempt aktif yang sama. Attempt lama `legacy-v1` dapat diselesaikan melalui submit lama.

Assessment:

```json
{
  "id":"asm-id",
  "student_id":"user-id",
  "grade_level":4,
  "status":"in_progress",
  "started_at":"2026-09-26T00:00:00Z",
  "items":[{
    "id":"item-id",
    "assessment_id":"asm-id",
    "question_id":"q-eq-1",
    "order_index":1,
    "question":{"id":"q-eq-1","skill_id":"frac_equiv","difficulty":1,"prompt":"Pertanyaan","options":["A","B"]}
  }]
}
```

Selama assessment belum selesai, question tidak memuat `explanation`, `misconception`, atau `answer_key`. Aturan ini berlaku untuk start baru, resume, dan latest. `answer_key` tidak pernah diserialisasi, termasuk pada hasil selesai.

## POST /api/diagnostic/submit

Request:

```json
{"assessment_id":"asm-id","answers":[{"question_id":"q-eq-1","student_answer":"A"}]}
```

- `assessment_id` wajib. Backend mengambil kunci dan menghitung evidence; client tidak mengirim is_correct/mastery.
- Saat ini grading adalah kecocokan string case-insensitive setelah trim. Jawaban kosong/tidak ada tidak dihitung sebagai evidence; ID soal di luar attempt diabaikan dan duplicate ID memakai entri terakhir. Normalisasi matematis umum belum diimplementasikan.
- `200`: Assessment `completed` tersimpan, dengan `completed_at`, `student_answer`, `is_correct`, `answered_at` per item, question explanations, dan `results` per skill.
- Skill result: `skill_id`, `skill_name`, `status` (`unassessed`, `needs_practice`, `mastered`), `total_answered`, `total_correct`, `evidence_count`, `confidence` (`low`, `medium`, `high`), `is_root_gap`.
- `is_root_gap` masih false pada Phase 1; backtracking belum tersedia. Status mastery ini aturan demo awal, bukan hasil pedagogis tervalidasi.
- `400`: JSON invalid atau assessment ID kosong.
- `403`: `Akses ditolak` untuk owner berbeda.
- `404`: `Asesmen tidak ditemukan`.
- `409`: `Asesmen sudah selesai. Jawaban tidak dapat diubah.`
- `500`: `Gagal mengevaluasi asesmen. Silakan coba lagi.`

Hasil selesai immutable: submit ulang tidak mengubah jawaban, evidence, atau timestamp. Dua submit bersamaan menghasilkan paling banyak satu completion; lainnya `409`. Claim status, jawaban, dan evidence berada dalam satu transaksi; kegagalan storage me-rollback seluruh perubahan. Jika response sukses hilang di jaringan, ambil latest untuk memeriksa completion sebelum mencoba submit ulang; attempt selesai tetap mengembalikan `409`, bukan hasil baru.

## GET /api/diagnostic/latest

- Tanpa body/query. Hanya latest assessment milik session user berdasarkan `started_at`.
- `200`: Assessment dengan bentuk di atas; `in_progress` tanpa solusi, `completed` dengan explanation dan results.
- `200` tanpa `data`: belum ada attempt; frontend menampilkan beranda siswa dengan status belum mulai.
- `500`: `Gagal memuat asesmen`; frontend menampilkan error dan retry, mempertahankan session, serta tidak memperlakukannya sebagai assessment kosong.
- GET/retry tidak memodifikasi attempt.

## GET /api/diagnostic/history

Tanpa body/query; hanya milik session user. `200` memiliki `data`:

```json
{"completed_count":0,"items":[]}
```

`completed_count` adalah jumlah seluruh assessment selesai milik user, tidak dibatasi jumlah history yang ditampilkan. `items` memuat maksimal lima assessment terbaru, urut `started_at` terbaru dengan ID sebagai tie-break. Setiap item memuat `id`, `grade_level`, `status`, `started_at`, `completed_at` jika selesai, `question_count`, `answered_count`, `correct_count` (hanya setelah completed; dihilangkan saat in_progress), dan `assessed_skill_count` (evidence_count > 0). Semua count berasal dari storage; record aktif tidak dianggap sudah selesai. Tidak ada soal, kunci, atau explanation pada history summary.

`401`: session invalid. `500`: `Gagal memuat riwayat belajar. Silakan coba lagi.` Client menampilkan error/retry dan tidak menggantinya dengan statistik nol. GET/retry tidak mengubah hasil.

## GET /api/diagnostic/{id}

`id` adalah ID assessment yang dipilih pada history. `200`: bentuk Assessment yang sama dengan latest; penjelasan hanya tersedia ketika selesai. `401`: session invalid; `403`: owner berbeda; `404`: assessment tidak ditemukan; `500`: `Gagal memuat asesmen. Silakan coba lagi.` Semua error storage dicatat server-side. GET tidak memodifikasi attempt atau membuat assessment baru.

## Verifikasi perubahan kontrak

Backend tests berada di dekat package. Jalankan `go test -race ./...`, `go vet ./...`, dan `go build ./...` dari `apps/api`. Untuk mengaktifkan regression tests PostgreSQL, sediakan `TEST_DATABASE_URL` ke database lokal khusus test; tiap test memakai schema terisolasi dan membersihkannya. Tanpa variable tersebut integration tests di-skip. Build/lint frontend dari `apps/web`: `npm run build` dan `npm run lint`.

## POST /api/diagnostic/answer

Session role student, owner assessment, semua environment. Request: `{"assessment_id":"asm-id","question_id":"q-id","student_answer":"<salah satu opsi>"}`. Field wajib nonempty, opsi harus persis salah satu opsi soal aktif; payload maksimal 16 KiB, tidak boleh field asing/trailing JSON.

200: Assessment terbaru setelah jawaban disimpan atomically, termasuk tepat satu pending question berikutnya atau status completed. `revision` bertambah untuk setiap jawaban baru; semua item answered memiliki `answered_at`, `probe_for_skill_id` mencatat skill induk yang memicu pemeriksaan. Saat in_progress, is_correct, results, path, explanation dan answer_key tidak dikirim. History juga menghilangkan correct_count untuk attempt aktif. Field `rule_version`, `target_skill_id`, `max_questions=18` menunjukkan aturan dan scope demo.

Retry dengan assessment/question/answer yang sama mengembalikan state terbaru tanpa evidence/next question duplikat, termasuk retry jawaban terakhir sesudah completion. Jawaban berbeda pada question yang sudah committed menghasilkan 409. Jawaban terhadap question yang belum ditawarkan menghasilkan 400; tidak boleh mengirim banyak jawaban sekaligus. Concurrent submissions diserialisasi berdasarkan revision dalam transaction; state terbaru digunakan setelah konflik. Request yang kalah dengan jawaban berbeda menerima 409, bukan menimpa evidence.

400 `Jawaban atau permintaan tidak valid` untuk JSON/field/opsi/mode invalid; 401 session invalid; 403 role/owner invalid; 404 assessment tidak ada; 409 `Jawaban sudah tersimpan dan tidak dapat diubah` untuk perubahan committed; 500 `Jawaban belum bisa disimpan. Silakan coba lagi.` dengan detail server-only. Retry 500 menggunakan payload sama; frontend mempertahankan pilihan untuk retry.

Completion menyimpan evidence dan learning_path bersama jawaban terakhir. Stop reason: `evidence_complete`, `insufficient_evidence` (bank soal/prasyarat tidak cukup), atau `question_limit`. Status skill: `strong_evidence` (3/3), `needs_practice` (0–1/3), `inconclusive` (2/3 atau bukti <3), `unassessed` (0 jawaban). Ini klasifikasi demo, bukan mastery tervalidasi. Kandidat root gap hanya dari rantai prerequisite yang punya evidence langsung; provenance `related_target_skill_id` tersedia. `learning_path` urutan prerequisite dahulu, memuat skill_id/name/reason; enam skill slice kelas 4 kini dapat membuka lesson melalui learning/start.

## Compatibility Phase 1

Bulk `/api/diagnostic/submit` hanya untuk `legacy-v1`. Progressive attempt ditolak 400 `Gunakan pengiriman satu jawaban untuk diagnostic progresif`; tidak ada perubahan evidence. Hasil legacy tetap mempertahankan rules/label lama dan dapat dibuka tanpa diproses ulang. Client memilih quiz berdasarkan rule_version.

## Learning Phase 3: shared access/errors

Bank konten tambahan dikelola oleh command backend lokal, bukan endpoint publik. Hanya kandidat approved yang disalin ke tabel questions aktif; draft/rejected dan metadata/key tetap server-only. Aktivasi pertama dibatasi enam skill kelas 4 dengan lesson tersedia, level 1–2; graph/kelas baru tidak otomatis aktif. Import ID/content immutable tidak mengubah item, opsi, jawaban atau hasil attempt yang sudah tersimpan. Retry import tidak membuat duplikasi. Bentuk response/auth endpoint diagnostic dan learning tetap sama.

Semua `/api/learning/*` memakai student middleware: 401 session invalid, 403 nonstudent `Fitur ini hanya tersedia untuk siswa`; resource owner berbeda 403 `Akses ditolak`. Semua environment. JSON write maksimal 16 KiB, field asing/trailing JSON/non-object invalid 400 `Permintaan belajar tidak valid`. Detail storage hanya log server. 404 `Aktivitas atau materi tidak ditemukan`; 409 `Aktivitas atau jawaban sudah berubah. Buka kembali dari beranda.` untuk immutable answer/stage conflict; 409 `Soal baru untuk skill ini belum cukup. Pilih rekomendasi lain atau coba setelah materi ditambah.` untuk bank exhaustion; 500 `Data belajar belum bisa diproses. Silakan coba lagi.`. Retry read aman; retry write memakai payload yang sama.

## GET /api/learning/lessons/{skill_id}

200 Lesson: id, skill_id, title, estimated_minutes (2), steps [{heading,body,example}], hint. Tidak memuat kunci soal. Materi adalah konten demo awal yang diperiksa untuk konsistensi matematika, belum review kurikulum eksternal. Unknown skill 404. Tidak mengubah progress.

## POST /api/learning/start

Request: `{source_assessment_id,skill_id,request_id}` string wajib. Source adalah diagnostic progressive-demo-v1 kelas 4 yang completed milik siswa; skill harus bagian results source dan lesson tersedia. 400 jika source/mode/status/skill invalid. Request ID maksimal 100 karakter; unique per siswa. Retry request ID yang sama dengan source/skill berbeda 409; retry sama mengembalikan session yang sama, termasuk setelah completed. Start paralel diserialisasi per siswa dan resume active session untuk skill tersebut. Request ID baru yang meresume session dicatat sebagai alias, sehingga retry sesudah session completed tetap mengembalikan session yang sama. Sebelum session baru dibuat, bank harus memiliki sedikitnya 3 soal practice dan 3 reassessment yang belum pernah ditawarkan kepada siswa itu. Tidak mereset progress/history.

200 LearningSession: id, student_id, source_assessment_id, skill_id/name, rule_version=learning-demo-v1, stage (lesson/practice/reassessment/completed/exhausted), revision, started_at, lesson_completed_at?, completed_at?, before_score?, score?, outcome?, stop_reason?, review_skill_id?, lesson, items[]. Item: id, question_id, stage, order_index, question, student_answer?, answered_at?, is_correct?. Question answer_key tidak pernah dikirim. Pending practice tanpa explanation/is_correct; answered practice memberi feedback setelah commit. Reassessment correctness/solutions disembunyikan sampai session completed/exhausted. lesson completion tidak mengubah score.

## GET /api/learning/sessions/{id}

200 session owned sesuai bentuk di atas. GET tidak membuat aktivitas/jawaban atau reward. Session snapshot dibaca konsisten; digunakan resume setelah reload, termasuk untuk memeriksa response write yang hilang.

## POST /api/learning/lesson-complete

Request `{session_id}`. Transisi lesson -> practice menandai lesson_completed_at dan menawarkan satu soal practice secara atomik; revision naik. Retry setelah transisi mengembalikan state terkini tanpa menambahkan soal/mereset timer. Jika bank habis, stage exhausted dan stop_reason insufficient_questions; tidak membuat score baru. 200 session terbaru.

## POST /api/learning/answer

Request `{session_id,question_id,student_answer}` wajib; answer harus salah satu opsi pending question. 400 untuk opsi/soal belum ditawarkan, 409 untuk jawaban committed berbeda. Retry answer yang sama mengembalikan state terbaru tanpa duplikasi, termasuk setelah completed. Status/revision claim, jawaban, next question, transisi stage, dan progress commit bersama. Konflik concurrent answer menggunakan state terbaru; jawaban kalah tidak menimpa evidence.

Practice 3 jawaban -> reassessment 3 jawaban berbeda dari diagnostic/practice dan belum pernah ditawarkan kepada siswa. Difficulty menggunakan level tersedia: naik setelah benar/turun setelah salah, tie-break jarak level/level/ID; tidak mengulang soal untuk mastery. Dua kesalahan practice berturut-turut dapat memberi review_skill_id prasyarat sebagai rekomendasi review, bukan root-gap diagnosis. Reassessment completed memperbarui score dari tiga jawaban terakhir (rounded percentage), outcome strong_evidence 3/3, inconclusive 2/3, needs_practice 0–1/3. Ini score/evidence demo, bukan mastery pendidikan tervalidasi. before_score mengambil bukti terakhir sebelum sesi; score bisa naik/tetap/turun. Klik lesson selesai dan jawaban practice tidak mengubah score akademis. Stage exhausted tidak mengubah progress.

## GET /api/learning/progress

200 `{source_assessment_id?,skills:[],learning_path:[],active_sessions:[],recent_sessions:[],completed_count}`. Skills: skill_id/name, status, score? (absent untuk unassessed), evidence_count, correct_count, source (diagnostic/reassessment/unassessed), updated_at?. Ambil latest completed progressive diagnostic dan overlay reassessment yang lebih baru; jangan merusak snapshot hasil diagnostic lama. Learning_path topological berisi weak/inconclusive skills dan reason menyebut reassessment jika itulah evidence terbaru, tanpa mengunci skill unassessed seolah gagal. active_sessions memiliki summary id/skill/stage/date; recent_sessions maksimal 5, completed_count seluruh learning completions owner. Loading/error tidak disamakan dengan empty progress. Tidak memuat jawaban/solusi. No data: arrays [], count 0; skill map boleh berisi unassessed untuk slice tersedia.
