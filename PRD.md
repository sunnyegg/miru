# Product Requirement Document (PRD)

## 1. Executive Summary
**Miru (見る)** adalah aplikasi pemutar dan pengunduh media anime desktop berbasis *portable*, serba guna, dan berkinerja tinggi. Aplikasi ini dibangun menggunakan kombinasi **Wails (Go + React)** untuk memberikan pengalaman pengguna yang responsif dengan konsumsi memori (RAM & CPU) yang sangat efisien. 

Tidak seperti aplikasi media manager anime konvensional yang memerlukan *client* eksternal rumit, **Miru** menyediakan modul BitTorrent bawaan secara *native*, deteksi otomatis pemutar MPV eksternal (dengan opsi *multi-path* & *custom file picker*), serta otomatisasi penyinkronan riwayat tontonan ke **AniList** secara *real-time*.

---

## 2. System Architecture & Tech Stack

```text
┌────────────────────────────────────────────────────────┐
│                      Miru Client                       │
│                                                        │
│  ┌──────────────────┐            ┌──────────────────┐  │
│  │  React Frontend  │ ◄─Wails──► │    Go Backend    │  │
│  │ (Tailwind, Vite) │   Bridge   │  (Engine Core)   │  │
│  └──────────────────┘            └────────┬─────────┘  │
└───────────────────────────────────────────┼────────────┘
                                            │
         ┌──────────────────┬───────────────┼───────────────┬─────────────────┐
         ▼                  ▼               ▼               ▼                 ▼
  ┌─────────────┐    ┌─────────────┐  ┌───────────┐  ┌─────────────┐   ┌─────────────┐
  │ SQLite DB   │    │ Embedded/   │  │ MPV IPC   │  │ AniList API │   │ GitHub CI   │
  │ (modernc)   │    │ anacrolix   │  │ Socket    │  │ (GraphQL)   │   │ Auto-Update │
  └─────────────┘    └─────────────┘  └───────────┘  └─────────────┘   └─────────────┘

```

* **Core Framework:** Wails v2 (Go 1.22+ & React 18+ via Vite).
* **Media Player Engine:** External MPV Executable via JSON-IPC Inter-Process Communication.
* **Torrent Engine:** `anacrolix/torrent` (Pure Go BitTorrent Implementation).
* **Database Layer:** `modernc.org/sqlite` (CGO-free Pure Go Driver).
* **Metadata & Auth:** AniList GraphQL API v2 + OAuth2 Implicit Grant Flow.
* **File Name Parser:** `github.com/nssteinbrenner/anitogo` (Anitomy C++ port to Go).
* **Packaging & CI/CD:** Native Portable Binary (`.exe`, `.app`, biner Linux) + GitHub Actions.

---

## 3. Core Features & Functional Requirements

### 3.1 MPV Execution & Detection Engine

* **Automated Multi-Path Detection:** Memindai biner MPV di `PATH` sistem, lokasi instalasi umum (Program Files, Homebrew, `/usr/bin`), dan konfigurasi manual.
* **Custom File Picker:** Menyediakan dialog *file picker* OS untuk pengguna yang menempatkan `mpv.exe` portabel di folder khusus.
* **JSON-IPC Integration:** Mengendalikan MPV di latar belakang, memantau *watch progress* (persentase durasi tonton), dan membaca status *playback*.

### 3.2 Integrated Torrent & Seeding Management

* **In-App BitTorrent Downloader:** Memuat magnet link atau file torrent langsung tanpa membutuhkan client eksternal (seperti qBittorrent).
* **Default Seeding Policy:** Otomatis membatasi unggahan data (*seeding*) hingga mencapai rasio **0.5x** dari ukuran file yang diunduh.
* **Bandwidth Throttling:** Kontrol batas kecepatan unduh (*Download Speed Limit*) dan unggah (*Upload Speed Limit*) langsung melalui UI.
* **Multi-Source RSS Indexing:** Mendukung pencarian dan feed otomatis dari Nyaa.si, Tokyo Toshokan, dan endpoint fansub langsung.

### 3.3 AniList Sync & Parser Metadata

* **OAuth2 Authentication:** Pengguna dapat terhubung ke akun AniList menggunakan alur login browser yang aman.
* **Anitogo File Parsing:** Memecah string nama file torrent/lokal yang berantakan menjadi judul bersih, nomor episode, resolusi, dan nama grup fansub.
* **Auto Progress Update:** Mengirim mutasi GraphQL `SaveMediaListEntry` secara otomatis ketika pemutaran MPV mencapai **>= 85%** dari total durasi video.
* **Airing Calendar:** Menampilkan tab jadwal rilis mingguan interaktif berdasarkan zona waktu lokal pengguna.

### 3.4 Customization & Portable Experience

* **Zero-Installation Portable Binary:** Berjalan tanpa pemasangan sistem (UAC Admin-free), cocok diletakkan di USB atau folder lokal.
* **Comprehensive Settings UI:** Seluruh konfigurasi (folder download, batas rasio, threshold sync, path MPV) dapat diatur melalui antarmuka React.
* **In-App Splashscreen:** Tampilan *splashscreen* berbasis React UI yang halus saat aplikasi memuat modul database dan sistem di latar belakang.

---

## 4. Local Database Schema (SQLite)

Database disimpan secara permanen di direktori konfigurasi lokal pengguna (`%LOCALAPPDATA%\miru\app_data.db` pada Windows atau `~/.config/miru/app_data.db` pada Linux/macOS).

```sql
-- 1. Pengaturan Aplikasi Terpusat
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- 2. Cache Metadata AniList
CREATE TABLE IF NOT EXISTS anime_cache (
    anilist_id INTEGER PRIMARY KEY,
    title_romaji TEXT NOT NULL,
    title_english TEXT,
    cover_image TEXT,
    total_episodes INTEGER,
    status TEXT,
    synopsis TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 3. Status Unduhan & Histori Episode
CREATE TABLE IF NOT EXISTS episode_downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anilist_id INTEGER NOT NULL,
    episode_number INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    downloaded_bytes INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('DOWNLOADING', 'COMPLETED', 'FAILED', 'PAUSED')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (anilist_id) REFERENCES anime_cache(anilist_id),
    UNIQUE(anilist_id, episode_number)
);

```

---

## 5. Non-Functional Requirements

* **Performance:** Penggunaan memori (RAM) saat idle harus berada di bawah 100 MB.
* **Cross-Platform Building:** Kode sumber Go harus bebas dari ketergantungan CGO agar dapat dikompilasi ke Windows (x86_64), macOS (Universal/ARM64), dan Linux (x86_64) tanpa hambatan.
* **Network Efficiency:** Mendukung pembatas token (*rate limiter*) agar aktivitas *seeding* di latar belakang tidak membebankan koneksi internet pengguna.

---

## 6. Release & CI/CD Strategy

Setiap pembuatan *Git Tag* baru (`v*`) akan memicu workflow **GitHub Actions** untuk merilis biner resmi:

| Target Platform | Output File Pattern | Packaging Format |
| --- | --- | --- |
| **Windows** | `miru-windows-amd64.exe` | Standalone Executable (Embedded WebView2) |
| **macOS** | `miru-mac-universal.zip` | Compressed `.app` Bundle |
| **Linux** | `miru-linux-amd64` | Standalone ELF Executable |

---

## 7. Future Roadmap (Post-v1.0)

* **Discord Rich Presence (RPC):** Menampilkan status anime yang sedang diputar di profil Discord.
* **Shader Injection (Anime4K):** Opsi otomatisasi pengaktifan shader upscaling video pada MPV.
* **Desktop Notifications:** Notifikasi lokal OS ketika unduhan episode anime selesai di latar belakang.

```