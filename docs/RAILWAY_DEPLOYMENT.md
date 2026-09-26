# Handbook Deployment Railway: Tilik

Panduan praktis langkah-demi-langkah (step-by-step) untuk mendeploy stack aplikasi **Tilik** ke [Railway.com](https://railway.com) agar live online, selalu aktif tanpa laptop standby, dan siap dinilai juri hackathon.

---

## 🏗️ Arsitektur di Railway

Dalam 1 Project Railway, akan ada 3 komponen yang saling terhubung:

```
┌────────────────────────────────────────────────────────┐
│                   Railway Project                      │
│                                                        │
│  [ PostgreSQL Database ]                               │
│       │ (koneksi internal: ${{Postgres.DATABASE_URL}}) │
│       ▼                                                │
│  [ tilik-api ] ──────────────► https://tilik-api-xxx   │
│       ▲                                                │
│       │ (HTTP REST / CORS)                             │
│  [ tilik-web ] ──────────────► https://tilik-web-xxx   │
└────────────────────────────────────────────────────────┘
```

---

## 📋 Pra-Syarat

1. Akun di [Railway.com](https://railway.com) (login dengan akun GitHub).
2. Repositori `sevima-tilik` sudah di-push ke GitHub.

---

## 🚀 Langkah 1: Buat Project & PostgreSQL Database

1. Buka dashboard Railway, klik tombol **New Project**.
2. Pilih **Provision PostgreSQL**.
3. Tunggu ~10 detik hingga kartu database muncul di canvas Railway.
4. *(Opsional)* Klik kartu database tersebut -> tab **Settings** -> ganti namanya menjadi `Postgres` jika belum.

> **Catatan**: Railway otomatis membuat database dan variabel koneksi `${{Postgres.DATABASE_URL}}`. Tidak perlu ubah apa pun di database.

---

## ⚙️ Langkah 2: Deploy Backend API (`tilik-api`)

1. Di canvas project yang sama, klik **+ New** (di pojok kanan atas canvas) -> pilih **GitHub Repo** -> pilih repositori `sevima-tilik`.
2. Klik service yang baru dibuat, masuk ke tab **Settings**:
   - **Service Name**: Ubah menjadi `tilik-api`.
   - **Root Directory**: Isi `apps/api` lalu klik **Save**.
   *(Railway akan mendeteksi `apps/api/Dockerfile` secara otomatis)*.
3. Buka tab **Variables**, klik **+ New Variable** (atau **Raw Editor**) dan masukkan:

```env
APP_ENV=production
PORT=8080
DATABASE_URL=${{Postgres.DATABASE_URL}}
ALLOWED_ORIGIN=*
```

> **Tips `DATABASE_URL`**: Saat mengetik `${{` di Railway, akan muncul dropdown otomatis. Pilih **Postgres** -> **DATABASE_URL**.

4. Buka tab **Settings** -> scroll ke bagian **Networking**:
   - Klik tombol **Generate Domain**.
   - Railway akan memberikan domain publik, contohnya: `https://tilik-api-production.up.railway.app`.
   - **Simpan/Salin URL ini** untuk Langkah 3.

---

## 🌐 Langkah 3: Deploy Frontend Web (`tilik-web`)

1. Di canvas project yang sama, klik **+ New** -> pilih **GitHub Repo** -> pilih repositori yang sama (`sevima-tilik`).
2. Klik service baru tersebut, masuk ke tab **Settings**:
   - **Service Name**: Ubah menjadi `tilik-web`.
   - **Root Directory**: Isi `apps/web` lalu klik **Save**.
3. Masih di tab **Settings**, scroll ke bagian **Networking**:
   - Cari pengaturan **Port** (atau Service Port): isi dengan `80` lalu simpan (karena Nginx di Dockerfile expose port 80).
   - Klik tombol **Generate Domain**.
   - Railway akan memberikan domain publik web, contohnya: `https://tilik-web-production.up.railway.app`.
   - **Simpan/Salin URL ini** untuk sinkronisasi CORS di Langkah 4.
4. Buka tab **Variables**, tambahkan variabel berikut:

```env
VITE_API_URL=https://tilik-api-production.up.railway.app
```
*(Ganti nilai di atas dengan URL API asli yang didapat dari Langkah 2)*.

> **Penting**: `apps/web/Dockerfile` membutuhkan `VITE_API_URL` saat tahap build (`ARG VITE_API_URL`). Railway otomatis menyuntikkan variabel ini saat menjalankan `docker build`.

---

## 🔒 Langkah 4: Kunci Keamanan CORS di API

1. Kembali ke service **tilik-api** di dashboard Railway.
2. Buka tab **Variables**.
3. Ubah `ALLOWED_ORIGIN` dari `*` menjadi URL frontend web dari Langkah 3:

```env
ALLOWED_ORIGIN=https://tilik-web-production.up.railway.app
```

4. Railway akan otomatis me-redeploy service `tilik-api` dalam hitungan detik.

---

## ✅ Langkah 5: Verifikasi & Smoke Test

Lakukan pengujian cepat setelah build selesai:

1. **Uji Healthcheck Backend**:
   Buka di browser:
   ```
   https://<domain-api-kamu>.up.railway.app/health
   ```
   Harus menghasilkan:
   ```json
   {"env":"production","status":"ok"}
   ```

2. **Uji Frontend Web**:
   Buka di browser:
   ```
   https://<domain-web-kamu>.up.railway.app
   ```
   Pastikan halaman utama Tilik terbuka dengan sempurna (tampilan login responsif, asset CSS dan icon termuat).

3. **Uji Database Persistence**:
   - Login dan coba jalankan 1 sesi asesmen.
   - Sesi dan jawaban tersimpan di database PostgreSQL Railway.

---

## 🛠️ Panduan Troubleshooting

| Gejala Masalah | Penyebab Umum | Solusi |
|---|---|---|
| **Build Web Gagal: `VITE_API_URL build argument is required`** | Variabel `VITE_API_URL` belum diisi di tab Variables `tilik-web`. | Tambahkan `VITE_API_URL` di tab Variables service web, lalu klik **Redeploy**. |
| **CORS Error di Console Browser** | `ALLOWED_ORIGIN` di API tidak cocok dengan domain Web. | Samakan `ALLOWED_ORIGIN` di service `tilik-api` persis seperti domain web (perhatikan `https://` dan tanpa garis miring `/` di ujung). |
| **502 Bad Gateway di Frontend** | Port Nginx tidak terdeteksi oleh Railway. | Di Settings `tilik-web` -> Networking -> set **Port** manual ke `80`. |
| **API Error: Connection to database failed** | Format `DATABASE_URL` belum benar. | Pastikan `DATABASE_URL` menggunakan syntax `${{Postgres.DATABASE_URL}}` atau copy langsung connection string dari tab Connect database Postgres. |

---

*Handbook ini siap dijadikan referensi saat kamu melakukan deploy ke Railway.*
