# Tilik

**Temukan pijakan belajar berikutnya.**

Aplikasi diagnostic learning dan adaptive numeracy untuk siswa kelas 4 SD hingga kelas 9 SMP. Tilik menemukan konsep dan prasyarat yang belum dipahami siswa, menentukan pembelajaran berikutnya, dan membantu guru memilih intervensi berdasarkan bukti belajar.

Tilik menelusuri jawaban siswa untuk menjawab dua pertanyaan: **apa yang perlu dipelajari berikutnya?** dan **mengapa siswa kesulitan?**

Contohnya, siswa kesulitan membandingkan pecahan. Tilik memeriksa pemahaman pecahan senilai sebagai salah satu prasyarat. Jika jawaban menunjukkan kemungkinan gap di konsep tersebut, siswa diarahkan ke micro lesson dan latihan yang sesuai, kemudian dinilai kembali. Guru dapat melihat bukti kesulitan dan rekomendasi pembelajaran berikutnya.

Scope produk: [docs/PRODUCT.md](docs/PRODUCT.md). Rencana implementasi dan checkpoint per fase: [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md).

Kontrak endpoint yang sudah tersedia: [docs/API_CONTRACT.md](docs/API_CONTRACT.md).

MVP pertama memprioritaskan kelas 4 dan pengalaman web. React Native + Expo menjadi perluasan setelah demo web selesai.

Status:

- Berjalan: login/session, beranda siswa, riwayat/statistik, diagnostic progresif kelas 4 (Go/PostgreSQL/Docker).
- Attempt baru memakai aturan demo `progressive-demo-v1` dengan pemeriksaan prasyarat dan rekomendasi review; hasil legacy tetap tersedia.
- Micro lesson, adaptive practice, reassessment, dan progress/path tersedia untuk enam skill kelas 4.
- Dashboard guru/admin masih shell tanpa data kelas.
- Aturan score masih demo awal; mastery jangka panjang belum tersedia.

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
| Frontend | React + TypeScript + Vite |
| Backend | Go (net/http) |
| Mobile (optional after web MVP) | React Native + Expo |
| Database | PostgreSQL 16 |
| Auth | Session-based (PostgreSQL), bcrypt password hashing |
| Rate Limiting | Token bucket per-IP (10 req/min auth, 100 req/min API) |
| Local Runtime | Docker Compose |
| Production | Docker + HTTPS reverse proxy (Caddy/nginx) |
| Deployment | Container-ready, cloud-agnostic |

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

Prasyarat: Docker CLI dan Docker Compose. Jika memakai Colima pada macOS, jalankan terlebih dahulu:

```bash
colima start
```

### Quick Start (Development)

```bash
# Otomatis setup dan jalankan development environment
./dev-start.sh

# Atau manual:
cp .env.example .env
docker compose up --build
```

Kemudian buka:

```text
Frontend: http://localhost:5173
API:      http://localhost:8080
Health:   http://localhost:8080/health
```

**Demo accounts (development only):**
- Student: `student@tilik.local` / `student123`
- Teacher: `teacher@tilik.local` / `teacher123`
- Admin: `admin@tilik.local` / `admin123`

Hentikan stack dengan:

```bash
docker compose down
```

### Production Deployment

Untuk production deployment dengan HTTPS, rate limiting, dan security hardening:

📖 **[Lihat Deployment Guide](guides/DEPLOYMENT.md#production-deployment)**

Quick production deploy:
```bash
# 1. Setup environment
cp .env.production.example .env.production
# Edit .env.production dengan values production

# 2. Deploy dengan HTTPS (Caddy auto SSL)
./prod-deploy.sh
```

Setup native Node/Go bersifat opsional dan didokumentasikan di `guides/DEPLOYMENT.md`.

## Environment Variables

Di development, tombol akun demo memakai `POST /api/auth/demo-login` dengan role yang diizinkan backend; frontend tidak menyimpan password demo. Endpoint dan seed akun demo hanya aktif pada `APP_ENV=development`. Production menolak akun/session demo yang tersisa. Akun production dibuat lewat proses administratif, bukan demo-login.

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

Jangan commit secret asli.

## Security Features

Production-ready security measures:

- ✅ **Session-based auth** dengan secure, httpOnly cookies (auto-enabled di production)
- ✅ **bcrypt password hashing** dengan proper cost factor
- ✅ **Rate limiting** per IP address (auth endpoints + API endpoints)
- ✅ **CORS protection** dengan strict origin validation
- ✅ **Demo account lockout** otomatis di production
- ✅ **SQL injection protection** via parameterized queries
- ✅ **HTTPS enforcement** melalui reverse proxy (Caddy/nginx)
- ✅ **Database SSL** support untuk managed DB

Lihat [SECURITY.md](guides/SECURITY.md) untuk detail lengkap.

## Runtime & Deployment

Development menggunakan Docker Compose dengan hot-reload. Production menggunakan optimized container builds dengan multi-stage Dockerfile.

**Development:**
```bash
./dev-start.sh  # Vite dev server + Go hot-reload
```

**Production:**
```bash
./prod-deploy.sh  # Optimized builds + HTTPS + rate limiting
```

Cloud hosting bersifat cloud-agnostic (AWS/GCP/Azure/VPS). Container setup dan deployment guide: **[guides/DEPLOYMENT.md](guides/DEPLOYMENT.md)**

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

