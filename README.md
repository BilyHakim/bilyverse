# Bilyverse

Portofolio satu halaman Bily Hakim Erlangga dengan palet Graphite & Emerald. Halaman memuat profil, studi kasus Bilyverse, dan formulir kontak.

- **Frontend/** berisi Svelte dan Vite. Teks halaman ada di `src/App.svelte`, sedangkan palet dan tata letak ada di `src/style.css`.
- **Backend/** berisi API Go dengan Chi, koneksi PostgreSQL, dan migrasi SQL Goose.

## Jalankan secara lokal

Siapkan Node.js 22.12+ (atau 20.19+), Go 1.25+, dan PostgreSQL. Dari root repo, perintah berikut menjalankan PostgreSQL lokal melalui Docker Compose.

```powershell
docker compose up -d --wait
```

Di terminal backend, jalankan migrasi lalu nyalakan API.

```powershell
cd Backend
$env:DATABASE_URL = 'postgres://bilyverse:bilyverse@127.0.0.1:55432/bilyverse?sslmode=disable'
go run . migrate
go run .
```

Jika menggunakan instalasi PostgreSQL sendiri, buat database kosong dan ganti `DATABASE_URL` dengan koneksi database tersebut. Kredensial di contoh adalah untuk Compose lokal.

Di terminal kedua, dari root repo, nyalakan frontend.

```powershell
cd Frontend
npm.cmd ci
npm.cmd run dev
```

Buka URL lokal yang dicetak Vite, biasanya `http://localhost:5173`. Vite meneruskan permintaan `/api` ke backend di `127.0.0.1:8080`.

## Build dan pemeriksaan

- Jalankan `npm.cmd run build` dari `Frontend/` untuk menghasilkan `Frontend/dist/`.
- Jalankan `go test ./...` dari `Backend/` untuk menguji validasi API.
- Untuk menyajikan hasil build dan API dari satu origin, set `$env:STATIC_DIR = '../Frontend/dist'` sebelum menjalankan `go run .` dari `Backend/`. Buka `http://127.0.0.1:8080`.

`GET /api/health` memeriksa koneksi database. `POST /api/contact` menerima JSON dengan kolom `name`, `email`, dan `message`. Pesan tersimpan di tabel `contact_messages` dan dapat dibaca melalui klien PostgreSQL yang memiliki akses ke database.
