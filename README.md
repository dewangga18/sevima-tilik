# Tilik

**Temukan pijakan belajar berikutnya.**

Aplikasi diagnostic learning dan adaptive numeracy untuk siswa kelas 4 SD hingga kelas 9 SMP. Tilik membantu menemukan konsep dan prasyarat yang belum dipahami siswa, menentukan pembelajaran berikutnya, dan membantu guru memilih intervensi berdasarkan bukti belajar.

Tilik menelusuri jawaban siswa untuk membantu menjawab dua pertanyaan: **apa yang perlu dipelajari berikutnya?** dan **mengapa siswa kesulitan?**

Contohnya, siswa kesulitan membandingkan pecahan. Tilik memeriksa pemahaman pecahan senilai sebagai salah satu prasyarat. Jika jawaban menunjukkan kemungkinan gap di konsep tersebut, siswa diarahkan ke micro lesson dan latihan yang sesuai, kemudian dinilai kembali. Guru dapat melihat bukti kesulitan dan rekomendasi pembelajaran berikutnya.

Scope produk: [docs/PRODUCT.md](docs/PRODUCT.md). Rencana implementasi dan checkpoint per fase: [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md).

Kontrak endpoint yang sudah tersedia: [docs/API_CONTRACT.md](docs/API_CONTRACT.md).

MVP pertama memprioritaskan kelas 4 dan pengalaman web. React Native + Expo menjadi perluasan setelah demo web selesai.

Status: login/session, beranda siswa, riwayat/statistik, dan diagnostic progresif kelas 4 sudah berjalan melalui Go/PostgreSQL/Docker. Attempt baru memakai aturan demo progressive-demo-v1 dengan pemeriksaan prasyarat serta rekomendasi review; hasil legacy tetap tersedia. Guru/admin memiliki shell dashboard, tetapi data kelas/administrasi belum terhubung. Micro lesson, adaptive practice, reassessment, dan progress/path terbaru sudah tersedia untuk enam skill kelas 4. Aturan score masih demo awal; mastery jangka panjang belum tersedia.

## Core Features

- Diagnostic assessment dan pemeriksaan prerequisite gap.
- Learning path, micro lesson, adaptive practice, dan mastery progress.
- Insight guru tentang kesulitan siswa dan intervensi berikutnya.

Alur siswa yang sudah berjalan untuk slice kelas 4:

```text
Diagnostic -> Learning gap -> Pemeriksaan prasyarat
-> Micro lesson -> Adaptive practice -> Reassessment -> Progress konsep
```

Evaluasi jawaban dan mastery menggunakan aturan deterministic. AI dapat membantu penjelasan atau hint sebagai fitur opsional.

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React + TypeScript |
| Backend | Go |
| Mobile (optional after web MVP) | React Native + Expo |
| Database | PostgreSQL |
| Auth | Session backend PostgreSQL, bcrypt; demo-login development saja |
| Local Runtime | Docker Compose |
| Deployment | Container-ready, provider TBD |

Panduan batas frontend/backend berada di [guides/ARCHITECTURE.md](guides/ARCHITECTURE.md).

## Repository Structure

```text
apps/
  web/      # React + TypeScript
  api/      # Go API
docs/       # scope produk dan implementation plan
guides/     # project and agent rules
docker-compose.yml
AGENTS.md
README.md
```

## Local Setup

Docker adalah runtime lokal acuan. Siapkan Docker CLI, Docker Compose, dan runtime container. Jika menggunakan Colima pada macOS, jalankan terlebih dahulu:

```bash
colima start
```

Dari root repository:

```bash
cp .env.example .env
docker compose up --build
```

Then open:

```text
Frontend: http://localhost:5173
API:      http://localhost:8080
Health:   http://localhost:8080/health
```

Stop the stack with:

```bash
docker compose down
```

Native Node/Go setup is optional and documented in `guides/DEPLOYMENT.md`.

## Environment Variables

Di development, tombol akun demo memakai `POST /api/auth/demo-login` dengan role yang diizinkan backend; frontend tidak menyimpan password demo. Endpoint dan seed akun demo hanya aktif pada `APP_ENV=development`. Production menolak akun/session demo yang tersisa dan tidak menampilkan tombol tersebut. Akun pengguna production perlu disediakan melalui proses pengelolaan akun yang sesuai.

Frontend example:

```env
VITE_API_URL=http://localhost:8080
```

Backend example:

```env
PORT=8080
APP_ENV=development
DATABASE_URL=
ALLOWED_ORIGIN=http://localhost:5173
```

Never commit real secrets.

## Runtime & Deployment

The current priority is a reproducible Docker-based local environment. Cloud hosting is selected only after the MVP is stable.

Container setup and future deployment rules: `guides/DEPLOYMENT.md`.

## Project Guides

| File | Scope |
|---|---|
| `guides/PRD.md` | discovery, scope, P0-P3 prioritization, phased implementation plan |
| `guides/ARCHITECTURE.md` | stack, boundaries, data flow, structure |
| `guides/CLEAN_CODE.md` | implementation quality rules |
| `guides/DESIGN_SYSTEM.md` | reusable UI component structure |
| `guides/DESIGN_DIRECTION.md` | project-specific visual direction |
| `guides/SECURITY.md` | security floor |
| `guides/DEPLOYMENT.md` | Docker-first local runtime, env, CI/CD, deployment portability |
| `guides/GIT_CONVENTION.md` | commit format |

## Referensi desain

Arah UI Tilik mengambil inspirasi dari sapaan personal, statistik ringkas, serta aksen hijau gelap dan lime pada [referensi Gamified Learning App](docs/assets/gamified-learning-reference.png). Gambar ini adalah referensi eksternal yang diberikan pengguna, bukan screenshot Tilik. Panduan penerapannya berada di [Design Direction](guides/DESIGN_DIRECTION.md); elemen spin dan coins tidak termasuk scope MVP.
