# Madu: Honeypot Dashboard

Aplikasi web berbasis honeypot yang menyediakan jalur jebakan (misalnya `/.env`, `/wp-login.php`),
mencatat permintaan mencurigakan, mengklasifikasikannya, dan menampilkannya di dashboard.
Dibuat oleh Yandi Muhamad Fauzi (Universitas Kuningan) untuk Tugas Project mata kuliah Cyber Security.

- URL aplikasi: <ISI URL APLIKASI SETELAH DEPLOY>
- Repository: https://github.com/yandimuhamadfauzi/madu
- Akun uji untuk dosen dicantumkan di laporan PDF, bukan di repository ini.

## Latar belakang

Server yang terhubung ke internet terus dipindai bot dan alat otomatis. Pemilik situs kecil jarang tahu
serangan apa yang menimpa mereka. Madu menjawab hal itu dengan mengumpulkan, mengklasifikasikan, dan
menampilkan serangan tersebut agar polanya mudah dipahami.

## Fitur

- Registrasi dan login. Akun baru berstatus menunggu dan harus disetujui admin.
- Peran admin dan analis. Analis hanya membaca, admin mengelola pengguna, aturan, dan blocklist.
- Jalur jebakan: semua permintaan ke jalur yang tidak ada di aplikasi asli dicatat dan dijawab dengan respons umpan.
- Klasifikasi otomatis berdasarkan aturan di database (path, query, dan user-agent).
- CRUD aturan jebakan, blocklist IP, dan pengguna.
- Dashboard (ringkasan, grafik per jam, jenis serangan, jalur dan IP teratas) dan log serangan dengan filter.
- Laporan: ekspor CSV.
- Audit log aktivitas pengguna.

## Teknologi

| Bagian | Teknologi |
|---|---|
| Bahasa dan framework | Go, Gin |
| ORM dan database | GORM, MySQL |
| Tampilan | `html/template`, CSS sendiri (tanpa CDN) |
| Hosting | <ISI HOSTING APLIKASI> |
| Database cloud | <ISI PENYEDIA DATABASE> |
| Kontainer | Docker (lihat `Dockerfile`) |

Struktur singkat:
cmd/server/ titik masuk aplikasi
internal/config/ pembacaan environment variable
internal/models/ model database
internal/handlers/ handler HTTP
internal/middleware/ autentikasi, peran, CSRF, header keamanan, audit
internal/services/ password, sesi, klasifikasi, statistik, cache
web/templates/ template HTML
web/static/ CSS dan ikon

## Penerapan keamanan

| Kontrol | Penerapan |
|---|---|
| Hash password | bcrypt, kebijakan password (8+ karakter, huruf besar, kecil, angka) |
| Sesi | Token acak 256 bit, hanya hash-nya disimpan di database, cookie HttpOnly, Secure, SameSite=Lax, masa berlaku 8 jam |
| Kontrol akses | Peran admin dan analis diperiksa di server (middleware), bukan hanya di tampilan |
| CSRF | Token per sesi untuk form setelah login, token double-submit cookie untuk form sebelum login |
| Rate limiting | Login (per IP dan per username), pendaftaran, pencatatan jebakan (per IP dan global) |
| Validasi dan sanitasi | Validasi di server, query berparameter (anti SQL injection), escape otomatis `html/template` (anti XSS), netralisasi formula injection di CSV |
| Header keamanan | CSP ketat, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, HSTS (produksi), Cache-Control no-store untuk panel |
| Secret | Environment variable, tidak ada di repository (dipindai dengan gitleaks) |
| Data penyerang | Hanya metadata (waktu, IP, method, path+query, user-agent). Body request dan password yang dicoba tidak disimpan |
| Ketahanan server | Timeout koneksi, batas ukuran header dan body, batas pencatatan, cache aturan |
| Privasi | Retensi log (bawaan 90 hari), data uji ditandai terpisah dari data nyata |
| Audit | Aksi pengguna dan percobaan login dicatat tanpa isi form |

## Menjalankan secara lokal

Prasyarat: Go (versi di `go.mod`), MySQL 8.

1. Buat database dan user:

   ```sql
   CREATE DATABASE madu CHARACTER SET utf8mb4;
   CREATE USER 'madu_app'@'localhost' IDENTIFIED BY '<password-kuat>';
   GRANT ALL PRIVILEGES ON madu.* TO 'madu_app'@'localhost';
   ```

2. Salin konfigurasi dan isi nilainya (jangan commit `.env`):

   ```bash
   cp .env.example .env
   ```

3. Jalankan:

   ```bash
   go run ./cmd/server
   ```

   Aplikasi berjalan di `http://localhost:8080`. Untuk pengujian lokal tanpa HTTPS, set `COOKIE_SECURE=false`.

4. Atau dengan Docker:

   ```bash
   docker build -t madu .
   docker run --env-file .env -p 8080:8080 madu
   ```

## Konfigurasi environment variable

Contoh tanpa nilai ada di `.env.example`.

| Variabel | Keterangan |
|---|---|
| `PORT` | Port server (bawaan 8080) |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Koneksi database |
| `DB_TLS`, `DB_CA_CERT` | Koneksi terenkripsi ke database cloud (`DB_CA_CERT` berisi sertifikat CA format PEM) |
| `ADMIN_USERNAME`, `ADMIN_PASSWORD` | Admin pertama, dibuat saat server dimulai bila belum ada |
| `ANALIS_USERNAME`, `ANALIS_PASSWORD` | Akun analis pertama |
| `COOKIE_SECURE` | `false` hanya untuk pengujian lokal |
| `CLIENT_IP_HEADER` | Header IP klien dari platform hosting. Isi hanya jika platform menimpa header itu |
| `EVENT_RETENTION_DAYS` | Masa simpan log serangan (bawaan 90) |

## Penggunaan AI agent

- Alat: Claude (Anthropic).
- Bagian yang dibantu: perancangan fitur dan arsitektur, penulisan kode Go dan template, debugging, penulisan dokumentasi, dan penyusunan skenario uji keamanan.
- Contoh prompt utama:
  - "bantu saya membuat tugas project ini, dan ide project apa yang bagus"
  - "buatkan dulu prototype nya untuk yang honeypot dashboard"
  - "bantu saya mengerjakan panduan project MADU nya"
  - "untuk desain ui web nya bisa di perbagus ga?"
- Peninjauan dan verifikasi: kode ditinjau dan dijalankan sendiri, terutama autentikasi, sesi, CSRF, dan query database. Setiap skenario keamanan diuji manual dengan `curl` dan browser. Pemindaian dependensi dengan `govulncheck` dan pemindaian secret dengan `gitleaks`. Tidak ada secret atau data nyata yang dimasukkan ke prompt.

## Hasil self-testing keamanan (ringkasan)

Hasil lengkap, termasuk tangkapan layar sebelum dan sesudah perbaikan, ada di laporan PDF.

| Skenario | Hasil |
|---|---|
| SQL injection dan XSS pada form dan parameter URL | <ISI HASIL> |
| Kontrol akses (analis ke halaman admin, IDOR) | <ISI HASIL> |
| Rate limiting login dan variasi huruf username | <ISI HASIL> |
| CSRF (aksi tanpa token) | <ISI HASIL> |
| Header keamanan dan HTTPS (securityheaders.com, SSL Labs) | <ISI HASIL> |
| Dependensi (`govulncheck`) dan secret (`gitleaks`) | Bersih setelah perbaikan |
| Timeout koneksi lambat (slowloris) | <ISI HASIL> |

## Keterbatasan dan rencana perbaikan

- Pembatas login disimpan di memori, sehingga hilang saat server restart dan tidak dibagi antar-instance.
- Klasifikasi berbasis pola sederhana, sehingga bisa meleset. Pengembangan lanjutan: klasifikasi dengan machine learning.
- Hosting gratis bisa tidur saat sepi, dan database gratis bisa dimatikan saat tidak aktif.

## Etika dan privasi

Aplikasi hanya mencatat secara pasif dan tidak membalas atau menyerang balik. IP dapat dianggap data
pribadi, sehingga data dibatasi masa simpannya dan hanya metadata yang dicatat.
- **Filter Rentang Tanggal:** Pencarian log dan ekspor CSV kini mendukung penyaringan berdasarkan rentang tanggal spesifik untuk mempermudah analisis.
