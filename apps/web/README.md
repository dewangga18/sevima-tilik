# Tilik Web

Frontend web Tilik: React 19 + TypeScript + Vite. Berisi alur siswa (login, beranda, diagnostic progresif kelas 4, micro lesson, adaptive practice, reassessment, progress) dan shell dashboard guru/admin.

Konsumsi API Go via `src/services/api.ts`. Kontrak endpoint: [docs/API_CONTRACT.md](../../docs/API_CONTRACT.md).

## Menjalankan

Cara utama adalah Docker Compose dari root repository (lihat [README root](../../README.md)).

Jalankan native (opsional):

```bash
cp .env.example .env
npm install
npm run dev        # http://localhost:5173
```

## Skrip

| Perintah | Fungsi |
|---|---|
| `npm run dev` | dev server Vite |
| `npm run build` | typecheck (`tsc -b`) + build produksi |
| `npm run lint` | oxlint |
| `npm run preview` | preview hasil build |

## Environment

| Variabel | Contoh | Keterangan |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Base URL API backend |

## Struktur

```text
src/
  components/   # layar dan komponen UI
  services/     # klien API
  types/        # tipe bersama
  App.css       # design tokens (warna, spacing)
```
