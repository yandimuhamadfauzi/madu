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
- Jalur jebakan khusus (decoy): halaman `wp-login.php` palsu untuk memancing bot login.
- Klasifikasi otomatis berdasarkan aturan di database (path, query, dan user-agent).
- Blokir otomatis (auto-block): IP publik yang melakukan 5 serangan tingkat tinggi dalam 10 menit diblokir selama 1 jam pada jalur jebakan. Ambang dapat diatur lewat variabel `AUTOBLOCK_*`. IP lokal, privat, dan sejenisnya tidak pernah diblokir, setiap blokir otomatis tercatat di audit log, dan admin dapat menjadikannya permanen.
- CRUD aturan jebakan, blocklist IP, dan pengguna.
- Dashboard interaktif:
  - Ringkasan statistik dan jenis serangan.
  - Grafik tren serangan: Terintegrasi dengan Chart.js untuk menampilkan serangan per jam (24 jam terakhir).
  - Peta GeoIP interaktif: menggunakan peta bertema gelap Esri (CSP ketat) yang menampilkan persebaran penyerang dengan informasi ISP di popup, pengelompokan titik (marker clustering), dan batas pergerakan (zoom/pan).
  - Jalur dan IP teratas: menampilkan batas hingga 50 IP teratas dengan scrollbox agar tata letak tetap rapi.
- Log serangan dengan filter jenis, pencarian, dan rentang tanggal.
- Laporan: ekspor CSV yang mengikuti filter yang sedang dipakai (termasuk rentang tanggal).
- Audit log aktivitas pengguna.
- Pemantauan uptime: UptimeRobot memanggil rute `/healthz` tiap 5 menit agar instance gratis tidak tidur.

## Teknologi

| Bagian | Teknologi |
|---|---|
| Bahasa dan framework | Go, Gin |
| ORM dan database | GORM, MySQL 8.4 terkelola di Aiven |
| Tampilan | `html/template`, CSS sendiri, Chart.js, dan Leaflet.js (dengan MarkerCluster) |
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
internal/services/   password, sesi, klasifikasi, statistik, cache, blokir otomatis
web/templates/       template HTML
web/static/          CSS, peta, grafik, dan ikon
madu-ml/             prototipe deteksi anomali (Python, terpisah dari aplikasi)
```

## Penerapan keamanan

| Kontrol | Penerapan |
|---|---|
| Hash password | bcrypt, kebijakan password (8+ karakter, huruf besar, kecil, angka) |
| Sesi | Token acak 256 bit, hanya hash-nya disimpan di database, cookie HttpOnly, Secure, SameSite=Lax, masa berlaku 8 jam |
| Kontrol akses | Peran admin dan analis diperiksa di server (middleware), bukan hanya di tampilan |
| CSRF | Token per sesi untuk form setelah login, token double-submit cookie untuk form sebelum login |
| Rate limiting | Login (per IP dan per username), pendaftaran, pencatatan jebakan (per IP dan global) |
| Blokir otomatis | IP publik dengan 5 serangan tingkat tinggi dalam 10 menit diblokir 1 jam pada jalur jebakan; IP lokal, privat, dan CGNAT tidak pernah diblokir |
| Validasi dan sanitasi | Validasi di server (termasuk parameter URL dan tanggal), query berparameter (anti SQL injection), escape otomatis `html/template` (anti XSS), netralisasi formula injection di CSV |
| Header keamanan | CSP, Permissions-Policy, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, HSTS (produksi), Cache-Control no-store untuk panel |
| Transport | HTTPS dengan sertifikat TLS otomatis; koneksi ke database memakai TLS dengan verifikasi sertifikat CA |
| Secret | Environment variable, tidak ada di repository (dipindai dengan gitleaks) |
| Data penyerang | Hanya metadata (waktu, IP, method, path+query, user-agent). Body request dan password yang dicoba tidak disimpan |
| IP klien di balik proxy | Header IP platform hanya dipercaya bila diaktifkan lewat `CLIENT_IP_HEADER`, dan nilainya divalidasi sebagai alamat IP |
| Ketahanan server | Timeout koneksi, batas ukuran header dan body, batas pencatatan, cache aturan |
| Privasi | Retensi log (bawaan 90 hari), data uji ditandai terpisah dari data nyata |
| Audit | Aksi pengguna, percobaan login, dan blokir otomatis dicatat tanpa isi form |

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
| `AUTOBLOCK_ENABLED` | `false` untuk mematikan blokir otomatis (bawaan menyala) |
| `AUTOBLOCK_THRESHOLD`, `AUTOBLOCK_WINDOW_MINUTES`, `AUTOBLOCK_DURATION_MINUTES` | Ambang serangan (bawaan 5), jendela waktu (10 menit), dan lama blokir (60 menit) |

Variabel lain (misalnya untuk fitur peta) tercantum di `.env.example`.

## Prototipe deteksi anomali (terpisah dari aplikasi)

Prototipe berbasis Python (Pandas dan scikit-learn, algoritma Isolation Forest) untuk menandai pola
permintaan yang tidak biasa pada data log yang diekspor dari Madu. Ini deteksi anomali tanpa label,
bukan klasifikasi jenis serangan, dan tidak terhubung ke aplikasi yang berjalan di produksi.

- Lokasi kode: `madu-ml/`
- Cara menjalankan: unduh data log CSV dari panel admin, ubah namanya menjadi `madu-serangan.csv`, simpan di folder yang sama dengan skrip, lalu jalankan `python deteksi.py`. Dependensi: Python 3, pandas, scikit-learn (`pip install pandas scikit-learn`). Berkas CSV berisi IP asli sehingga tidak di-commit ke repository.
- Fitur yang dipakai: panjang path, metode HTTP, tingkat keparahan (severity), dan frekuensi permintaan per IP.
- Hasil singkat: pada pengujian awal dengan log nyata (lebih dari 200 entri), entri yang ditandai sebagai anomali antara lain permintaan dari pemindai massal (user-agent seperti OpenBash-Surface dan l9scan/leakix) dan akses ke berkas sensitif seperti `/.env` dan `/.git/config`.
- Keterbatasan: Isolation Forest tidak memakai label sehingga tidak ada ukuran akurasi, dan hasilnya perlu diverifikasi manual. Data nyata masih sedikit, sehingga hasilnya belum dapat digeneralisasi.

## Penggunaan AI agent

- Alat: Google Gemini dan Claude (Anthropic).
- Bagian yang dibantu: perancangan fitur dan arsitektur, penulisan kode Go dan template, konfigurasi domain kustom (CNAME), perancangan prototipe deteksi anomali (Python), debugging, penulisan dokumentasi, dan penyusunan skenario uji keamanan.
- Peninjauan dan verifikasi: kode ditinjau dan dijalankan sendiri, terutama autentikasi, sesi, CSRF, dan query database. Setiap skenario keamanan diuji manual dengan `curl` dan browser. Pemindaian dependensi dengan `govulncheck` dan pemindaian secret dengan `gitleaks`.
- Kepatuhan: tidak ada API key, password produksi, atau data pribadi pihak lain yang dimasukkan ke prompt. Satu kali password database untuk lingkungan lokal (bukan produksi) sempat tertulis dalam percakapan dengan AI; password itu hanya berlaku untuk database lokal dan sudah diganti pada Oktober 2026.

## Hasil self-testing keamanan (ringkasan)

Hasil lengkap, termasuk tangkapan layar sebelum dan sesudah perbaikan, ada di laporan PDF.

| Skenario | Hasil |
|---|---|
| Password disimpan sebagai hash bcrypt | Lolos (lokal): kolom `password_hash` berawalan `$2a$10$` |
| Rate limiting login dan variasi huruf username | Lokal: Lolos (lima kali 401 lalu 429). Publik: Lolos (429 setelah percobaan login gagal berulang) |
| Kontrol akses berbasis peran (analis ke halaman admin) | Lokal: Lolos (analis 403, admin 200). Publik: Lolos (tanpa login 303 ke halaman masuk; analis 403 pada `/panel/pengguna` dan `/panel/audit`; admin 200) |
| CSRF (aksi tanpa token) | Lokal: Lolos (403 tanpa token). Publik: Lolos (403 tanpa token) |
| SQL injection dan XSS pada form dan parameter URL | Publik: Lolos. Login dengan payload SQL injection 401; parameter URL berbahaya di `/panel/log` 403 (diblokir WAF Cloudflare sebelum sampai ke aplikasi), parameter wajar 200; payload script lewat User-Agent tampil ter-escape (`&lt;script&gt;`), tidak ada yang tampil mentah | |
| Body request dan password yang dicoba tidak disimpan | Lokal: Lolos pada jalur jebakan umum. Decoy `wp-login.php`: Lolos (hanya metadata yang disimpan, isi form tidak) |
| Formula injection pada ekspor CSV | Lolos (lokal): sel User-Agent berawalan `=` diawali tanda petik |
| Pemalsuan header IP di balik proxy | Lokal: Lolos (nilai bukan IP dibuang). Publik: Lolos (IP tercatat sama dengan IP publik penguji; `CF-Connecting-IP`, `True-Client-IP`, dan `X-Forwarded-For` palsu diabaikan) |
| Header keamanan dan HTTPS (securityheaders.com, SSL Labs) | Lolos. Domain kustom `madu.yandimfauzi.my.id` (9 November 2026): securityheaders.com A+, SSL Labs A+ pada kedua server. Domain Render `madu-s5cl.onrender.com` (3 November 2026): securityheaders.com A+, SSL Labs A+ pada kedua server |
| Dependensi (`govulncheck`) | Temuan awal diperbaiki lewat pembaruan toolchain dan modul; hasil akhir: tidak ada kerentanan ditemukan |
| Secret di riwayat repository (`gitleaks`) | Lolos pada 2 November (20 commit dipindai). Pindai ulang sebelum pengumpulan: Lolos (45 commit dipindai, tidak ada secret ditemukan) |
| Timeout koneksi lambat (slowloris) | Lokal: Lolos (koneksi lambat ditutup server) |
| Blokir otomatis | Lolos (permintaan ke-3 sampai ke-5 dari IP publik berubah dari 200 menjadi 403; IP privat tidak diblokir) |
| Validasi filter tanggal | Publik: Lolos (format salah dan rentang terbalik masing-masing menampilkan pesan galat) |
| Atribut cookie sesi | Publik: Lolos (`HttpOnly; Secure; SameSite=Lax`, masa berlaku 8 jam) |
| Logout mencabut sesi | Publik: Lolos (cookie lama setelah logout dialihkan 303 ke halaman masuk) |

## Keterbatasan dan rencana perbaikan

- Pembatas login dan penghitung blokir otomatis disimpan di memori, sehingga hilang saat server restart dan tidak dibagi antar-instance (blokir yang sudah tersimpan di database tetap berlaku).
- Klasifikasi berbasis pola sederhana, sehingga bisa meleset. Pengembangan lanjutan: klasifikasi dan deteksi anomali dengan machine learning pada data yang lebih banyak.
- Blokir otomatis hanya berlaku pada jalur jebakan dan dapat memblokir sementara IP yang sah yang kebetulan mengakses jalur jebakan; login dan panel tidak terpengaruh. Ketepatannya bergantung pada pembacaan IP klien yang benar.
- Peta GeoIP bergantung pada sumber data lokasi: layanan pihak ketiga (REST API eksternal) dengan mekanisme in-memory caching untuk meminimalkan latensi dan request berulang. Bila memakai layanan pihak ketiga, IP penyerang dikirim ke layanan tersebut.
- Hosting gratis bisa tidur saat sepi, dan database gratis bisa dimatikan saat tidak aktif.

## Etika dan privasi

Aplikasi hanya mencatat secara pasif dan tidak membalas atau menyerang balik. IP dapat dianggap data
pribadi, sehingga data dibatasi masa simpannya dan hanya metadata yang dicatat.
