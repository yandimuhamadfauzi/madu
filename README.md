# Madu: Honeypot Dashboard

Aplikasi web berbasis honeypot yang menyediakan jalur jebakan (misalnya `/.env`, `/wp-login.php`),
mencatat permintaan mencurigakan, mengklasifikasikannya, dan menampilkannya di dashboard.
Dibuat oleh Yandi Muhamad Fauzi (Universitas Kuningan) untuk Tugas Project mata kuliah Cyber Security.

- URL aplikasi utama: <https://madu.yandimfauzi.my.id>
- URL alternatif (Render): <https://madu-s5cl.onrender.com>
- Repository: <https://github.com/yandimuhamadfauzi/madu>
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
- Dashboard (ringkasan, grafik per jam, jenis serangan, jalur dan IP teratas).
- Log serangan dengan filter jenis, pencarian, dan rentang tanggal.
- Laporan: ekspor CSV yang mengikuti filter yang sedang dipakai (termasuk rentang tanggal).
- Audit log aktivitas pengguna.
- Pemantauan uptime: UptimeRobot memanggil rute `/healthz` tiap 5 menit agar instance gratis tidak tidur.

## Teknologi

| Bagian | Teknologi |
|---|---|
| Bahasa dan framework | Go, Gin |
| ORM dan database | GORM, MySQL 8.4 terkelola di Aiven |
| Tampilan | `html/template`, CSS sendiri (tanpa CDN) |
| Hosting aplikasi | Render (web service Docker) |
| Domain dan DNS | `yandimfauzi.my.id`, DNS dikelola lewat InfinityFree, sertifikat TLS otomatis |
| Pemantauan uptime | UptimeRobot |
| Kontainer | Docker (lihat `Dockerfile`) |
| Prototipe deteksi anomali (terpisah) | Python, Pandas, scikit-learn (Isolation Forest) |

Struktur singkat:

```
cmd/server/          titik masuk aplikasi
internal/config/     pembacaan environment variable
internal/models/     model database
internal/handlers/   handler HTTP
internal/middleware/ autentikasi, peran, CSRF, header keamanan, audit
internal/services/   password, sesi, klasifikasi, statistik, cache
web/templates/       template HTML
web/static/          CSS dan ikon
```

## Penerapan keamanan

| Kontrol | Penerapan |
|---|---|
| Hash password | bcrypt, kebijakan password (8+ karakter, huruf besar, kecil, angka) |
| Sesi | Token acak 256 bit, hanya hash-nya disimpan di database, cookie HttpOnly, Secure, SameSite=Lax, masa berlaku 8 jam |
| Kontrol akses | Peran admin dan analis diperiksa di server (middleware), bukan hanya di tampilan |
| CSRF | Token per sesi untuk form setelah login, token double-submit cookie untuk form sebelum login |
| Rate limiting | Login (per IP dan per username), pendaftaran, pencatatan jebakan (per IP dan global) |
| Validasi dan sanitasi | Validasi di server (termasuk parameter URL dan tanggal), query berparameter (anti SQL injection), escape otomatis `html/template` (anti XSS), netralisasi formula injection di CSV |
| Header keamanan | CSP ketat, X-Frame-Optiazons, X-Content-Type-Options, Referrer-Policy, HSTS (produksi), Cache-Control no-store untuk panel |
| Transport | HTTPS dengan sertifikat TLS otomatis; koneksi ke database memakai TLS dengan verifikasi sertifikat CA |
| Secret | Environment variable, tidak ada di repository (dipindai dengan gitleaks) |
| Data penyerang | Hanya metadata (waktu, IP, method, path+query, user-agent). Body request dan password yang dicoba tidak disimpan |
| IP klien di balik proxy | Header IP platform hanya dipercaya bila diaktifkan lewat `CLIENT_IP_HEADER`, dan nilainya divalidasi sebagai alamat IP |
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

## Prototipe deteksi anomali (terpisah dari aplikasi)

Prototipe berbasis Python (Pandas dan scikit-learn, algoritma Isolation Forest) untuk mendeteksi pola
permintaan yang tidak biasa pada data log yang diekspor dari Madu. Ini deteksi anomali tanpa label,
bukan klasifikasi jenis serangan, dan tidak terhubung ke aplikasi yang berjalan di produksi.

- Lokasi kode: madu-ml
- Cara menjalankan: Unduh data log CSV dari panel admin, ubah namanya menjadi madu-serangan.csv dan simpan di folder yang sama dengan skrip Python. Jalankan program dengan perintah python deteksi.py
- Fitur yang dipakai dan hasil singkat: Model ini menganalisis fitur seperti panjang path, metode HTTP, tingkat keparahan (severity), dan frekuensi request per IP. Berdasarkan pengujian awal (contoh log nyata dari >200 entri), model mendeteksi anomali bot pemindai massal (seperti OpenBash-Surface dan l9scan/leakix) serta anomali akses pada berkas sensitif /.env dan /.git/config.
- Keterbatasan: data nyata masih sedikit, sehingga hasilnya belum dapat digeneralisasi.

## Penggunaan AI agent

- Alat: Google Gemini dan Claude (Anthropic).
- Bagian yang dibantu: perancangan fitur dan arsitektur, penulisan kode Go dan template, konfigurasi domain kustom (CNAME), perancangan prototipe deteksi anomali (Python), debugging, penulisan dokumentasi, dan penyusunan skenario uji keamanan.
- Contoh prompt utama:
  - "bantu saya membuat tugas project ini, dan ide project apa yang bagus"
  - "buatkan dulu prototype nya untuk yang honeypot dashboard"
  - "bantu saya mengerjakan panduan project MADU nya"
  - "untuk desain ui web nya bisa di perbagus ga?"
  - "bagaimana cara menghubungkan yandimfauzi.my.id dengan web ini"
  - "saya ingin menambahkan deteksi anomali menggunakan python"
- Peninjauan dan verifikasi: kode ditinjau dan dijalankan sendiri, terutama autentikasi, sesi, CSRF, dan query database. Setiap skenario keamanan diuji manual dengan `curl` dan browser. Pemindaian dependensi dengan `govulncheck` dan pemindaian secret dengan `gitleaks`.
- Kepatuhan: tidak ada API key, password produksi, atau data pribadi pihak lain yang dimasukkan ke prompt. Satu kali password database untuk lingkungan lokal (bukan produksi) sempat tertulis dalam percakapan dengan AI; password itu hanya berlaku untuk database lokal dan sudah diamankan ulang pada Oktober 2026.

## Hasil self-testing keamanan (ringkasan)

Hasil lengkap, termasuk tangkapan layar sebelum dan sesudah perbaikan, ada di laporan PDF.

| Skenario | Hasil |
|---|---|
| Password disimpan sebagai hash bcrypt | Lolos (diuji lokal) |
| Rate limiting login | Lolos di lokal (lima kali 401 lalu 429); aplikasi publik: Lolos (berfungsi aktif di produksi) |
| Kontrol akses berbasis peran (analis ke halaman admin) | Lolos di lokal (analis 403, admin 200); aplikasi publik: Lolos (403 teruji di publik) |
| CSRF (aksi tanpa token) | Lolos di lokal (403 tanpa token) |
| SQL injection dan XSS pada form dan parameter URL | Lolos (dinetralkan parameterized query & html/template escape) |
| Body request dan password yang dicoba tidak disimpan | Lolos (diuji lokal) |
| Formula injection pada ekspor CSV | Lolos (diuji lokal) |
| Pemalsuan header IP di balik proxy | Lolos (ditangani aman sesuai konfigurasi Gin) |
| Header keamanan dan HTTPS (securityheaders.com, SSL Labs) | Lolos (domain kustom aman dengan TLS aktif) |
| Dependensi (`govulncheck`) | Temuan awal diperbaiki lewat pembaruan toolchain dan modul; hasil akhir: Bersih tanpa celah kerentanan kritis setelah pembaruan modul |
| Secret di riwayat repository (`gitleaks`) | Lolos (20 commit dipindai, tidak ada kebocoran) |
| Timeout koneksi lambat (slowloris) | Lolos (server Go dikonfigurasi Read/Write timeout) |

## Keterbatasan dan rencana perbaikan

- Pembatas login disimpan di memori, sehingga hilang saat server restart dan tidak dibagi antar-instance.
- Klasifikasi berbasis pola sederhana, sehingga bisa meleset. Pengembangan lanjutan: klasifikasi dan deteksi anomali dengan machine learning pada data yang lebih banyak.
- Hosting gratis bisa tidur saat sepi, dan database gratis bisa dimatikan saat tidak aktif.

## Etika dan privasi

Aplikasi hanya mencatat secara pasif dan tidak membalas atau menyerang balik. IP dapat dianggap data
pribadi, sehingga data dibatasi masa simpannya dan hanya metadata yang dicatat.
