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
$env:HTTP_ADDR = '127.0.0.1:8081'
go run .
```

Jika menggunakan instalasi PostgreSQL sendiri, buat database kosong dan ganti `DATABASE_URL` dengan koneksi database tersebut. Kredensial di contoh adalah untuk Compose lokal.

Di terminal kedua, dari root repo, nyalakan frontend.

```powershell
cd Frontend
npm.cmd ci
npm.cmd run dev
```

Buka URL lokal yang dicetak Vite, biasanya `http://localhost:5173`. Vite meneruskan permintaan `/api` ke backend di `127.0.0.1:8081`.

## Dashboard konten

Setelah `DATABASE_URL` diatur, jalankan `go run . migrate` dari `Backend/`.
Set `CMS_ADMIN_KEY` di terminal backend ke kunci acak minimal 32 karakter, lalu mulai ulang API dengan `go run .`.
Contoh untuk membuat kunci di PowerShell:

```powershell
$cmsKeyBytes = New-Object byte[] 32
$cmsRng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$cmsRng.GetBytes($cmsKeyBytes)
$cmsRng.Dispose()
$env:CMS_ADMIN_KEY = [Convert]::ToBase64String($cmsKeyBytes)
```

Salin nilai kunci dari terminal sendiri dengan `$env:CMS_ADMIN_KEY` dan simpan di pengelola kata sandi. Buka URL frontend dengan akhiran `/#admin` dan masukkan kunci.

Portofolio, Sertifikat, dan Blog mendukung hingga 50 entri per bagian. Tambahkan judul, ringkasan, detail, tanggal, dan tautan sesuai kebutuhan. Centang **Tampilkan di website** pada entri yang siap diterbitkan, lalu pilih **Simpan perubahan**. Entri lain tetap menjadi draf. Konten dari editor lama muncul sebagai entri awal. Gunakan HTTPS jika dashboard diakses di luar localhost.

Setiap entri Portofolio, Sertifikat, dan Blog dapat memuat hingga delapan gambar. Pilih beberapa file JPEG, PNG, atau WebP sekaligus (maksimal 10 MB per file). Setelah pratinjau muncul, tambahkan teks alternatif dan pilih **Jadikan sampul** untuk mengatur gambar utama. Pilih **Simpan perubahan** untuk menyimpan galeri. Batas total satu penyimpanan adalah 32 MiB setelah optimasi dan pengodean gambar.

## Build dan pemeriksaan

- Jalankan `npm.cmd run build` dari `Frontend/` untuk menghasilkan `Frontend/dist/`.
- Jalankan `go test ./...` dari `Backend/` untuk menguji validasi API.
- Untuk menyajikan hasil build dan API dari satu origin, set `$env:STATIC_DIR = '../Frontend/dist'` sebelum menjalankan `go run .` dari `Backend/`. Buka alamat yang dicetak backend.

`GET /api/health` memeriksa koneksi database. `POST /api/contact` menerima JSON dengan kolom `name`, `email`, dan `message`. Pesan tersimpan di tabel `contact_messages` dan dapat dibaca melalui klien PostgreSQL yang memiliki akses ke database.
