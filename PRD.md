# Product Requirement Document (PRD)

> **Status audit:** 2026-08-28 — ~~dicoret~~ = sudah ada di codebase. Entri tanpa coretan = belum / sebagian. Teks yang diperbarui mencerminkan implementasi aktual.

## 1. Executive Summary
**Miru (見る)** adalah aplikasi pemutar dan pengunduh media anime desktop berbasis *portable*, serba guna, dan berkinerja tinggi. Aplikasi ini dibangun menggunakan kombinasi **Wails (Go + React)** untuk memberikan pengalaman pengguna yang responsif dengan konsumsi memori (RAM & CPU) yang sangat efisien.

Tidak seperti aplikasi media manager anime konvensional yang memerlukan *client* eksternal rumit, **Miru** menyediakan modul BitTorrent bawaan secara *native*, deteksi otomatis pemutar MPV eksternal (dengan opsi *custom file picker*), serta penyinkronan riwayat tontonan ke **AniList** saat sesi MPV berakhir.

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
  │ (modernc)   │    │ anacrolix   │  │ Socket    │  │ (GraphQL)   │   │ + Releases  │
  └─────────────┘    └─────────────┘  └───────────┘  └─────────────┘   └─────────────┘
                                                                               │
                                                                        ┌──────┴──────┐
                                                                        │ In-app      │
                                                                        │ auto-update │
                                                                        └─────────────┘
```

* **Core Framework:** Wails v2 (Go 1.25 & React 18+ via Vite).
* **Media Player Engine:** External MPV executable via JSON-IPC.
* **Torrent Engine:** `anacrolix/torrent` (Pure Go BitTorrent implementation).
* **Database Layer:** `modernc.org/sqlite` (CGO-free Pure Go driver).
* **Metadata & Auth:** AniList GraphQL API v2 + OAuth2 **authorization code** flow (browser login + local callback server; token disimpan di keyring dengan fallback file).
* **File Name Parser:** `github.com/nssteinbrenner/anitogo` (Anitomy C++ port to Go).
* **Packaging & CI/CD:** Portable binary (`.exe`, `.app`, ELF Linux) + GitHub Actions (`ci.yml` untuk PR/push, `release.yml` untuk tag `v*`) + pemeriksaan pembaruan in-app dari GitHub Releases.

---

## 3. Core Features & Functional Requirements

### 3.1 MPV Execution & Detection Engine

* ~~**Automated Multi-Path Detection:** Memindai biner MPV di `PATH` sistem, lokasi instalasi umum (Program Files, Homebrew, `/usr/bin`), dan konfigurasi manual.~~
  * **Implementasi:** `PATH` + lokasi umum **Linux** (`/usr/bin`, `/usr/local/bin`, `~/.local/bin`, Snap). Windows/macOS mengandalkan `PATH` + path manual di Settings.
* ~~**Custom File Picker:** Menyediakan dialog *file picker* OS untuk pengguna yang menempatkan biner MPV portabel di folder khusus.~~
* ~~**JSON-IPC Integration:** Mengendalikan MPV, memantau *watch progress* (persentase durasi tonton), dan membaca status *playback*.~~
  * **Implementasi:** MPV diluncurkan dengan jendela sendiri (`--force-window=yes`), bukan headless. Progress dipoll via IPC; posisi resume disimpan ke SQLite saat MPV ditutup.

### 3.2 Integrated Torrent & Seeding Management

* ~~**In-App BitTorrent Downloader:** Memuat magnet link atau file `.torrent` langsung tanpa client eksternal.~~
  * **Implementasi:** Magnet, URL `.torrent`, dan file picker OS. **Inspect → pilih file** sebelum unduh (`InspectTorrent`, `StartTorrent`).
* ~~**Default Seeding Policy:** Otomatis membatasi unggahan (*seeding*) hingga rasio target (default **0.5×**) dari ukuran file yang diunduh.~~
  * **Implementasi:** Rasio bisa diatur di Settings → Downloads (`seed_ratio`, rentang 0–10). 0 = stop seeding segera setelah unduh selesai.
* ~~**Bandwidth Throttling:** Kontrol batas kecepatan unduh dan unggah langsung melalui UI.~~
  * **Implementasi:** Juga ada batas **maksimum unduhan bersamaan** dan antrian `QUEUED`.
* **Multi-Source RSS Indexing:** Pencarian on-demand dari Nyaa.si dan Tokyo Toshokan (RSS sebagai API pencarian).
  * ~~Nyaa.si~~ dan ~~Tokyo Toshokan~~ sudah ada.
  * **Belum:** feed RSS otomatis / langganan background, endpoint fansub langsung.

### 3.3 AniList Sync & Parser Metadata

* ~~**OAuth2 Authentication:** Pengguna dapat terhubung ke akun AniList menggunakan alur login browser yang aman.~~
* ~~**Anitogo File Parsing:** Memecah nama file torrent/lokal menjadi judul, nomor episode, resolusi, dan grup fansub.~~
  * **Implementasi:** Metadata parse disimpan di backend; UI Library belum menampilkan resolusi/grup fansub.
* ~~**Auto Progress Update:** Mengirim mutasi GraphQL `SaveMediaListEntry` ketika pemutaran mencapai threshold (default **≥ 85%**, bisa diatur di Settings).~~
  * **Implementasi:** Sync terjadi **saat MPV ditutup**, bukan real-time di tengah pemutaran. Dedup via tabel `sync_events`. Mapping episode multi-season via AniList season API.
* ~~**Airing Calendar:** Tab jadwal rilis mingguan berdasarkan zona waktu lokal.~~

**Di luar PRD asli, sudah ada:**

* ~~Tab **Watching** — kelola entri AniList (CURRENT/COMPLETED/PLANNING/…), edit skor & progress manual.~~
* ~~**Library lokal** — impor file video, bind ke AniList (manual atau auto-match), poster grid, daftar episode, auto-ingest dari torrent selesai.~~
* ~~Overlay progress AniList pada episode lokal; strip "Watching" di Library.~~

### 3.4 Customization & Portable Experience

* ~~**Zero-Installation Portable Binary:** Berjalan tanpa pemasangan sistem, cocok di USB atau folder lokal.~~
* ~~**Comprehensive Settings UI:** Folder download, batas kecepatan, rasio seeding, threshold sync AniList, path MPV, jaringan (system/direct/SOCKS5), pembaruan otomatis.~~
  * **Belum di Settings:** proxy HTTP/HTTPS (hanya SOCKS5).
* ~~**In-App Splashscreen:** Splash React saat bootstrap database dan modul backend.~~

---

## 4. Local Database Schema (SQLite)

Database disimpan di direktori konfigurasi lokal (`%LOCALAPPDATA%\miru\app_data.db` pada Windows atau `~/.config/miru/app_data.db` pada Linux/macOS). Schema aktual (v6):

```sql
-- Pengaturan aplikasi (key-value)
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Cache metadata AniList
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

-- Episode lokal (impor / unduhan selesai)
CREATE TABLE IF NOT EXISTS episode_downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anilist_id INTEGER,              -- nullable sampai di-bind
    episode_number INTEGER,          -- nullable sampai di-parse/bind
    file_path TEXT NOT NULL UNIQUE,
    display_title TEXT,
    downloaded_bytes INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('DOWNLOADING', 'COMPLETED', 'FAILED', 'PAUSED')),
    resume_position REAL NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (anilist_id) REFERENCES anime_cache(anilist_id)
);
CREATE UNIQUE INDEX episode_unique
    ON episode_downloads(anilist_id, episode_number)
    WHERE anilist_id IS NOT NULL AND episode_number IS NOT NULL;

-- Job torrent (unduh + seeding)
CREATE TABLE IF NOT EXISTS torrent_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    info_hash TEXT,
    source TEXT NOT NULL,
    dest_dir TEXT NOT NULL,
    name TEXT,
    status TEXT NOT NULL CHECK(status IN (
        'QUEUED', 'DOWNLOADING', 'PAUSED', 'SEEDING',
        'COMPLETED', 'FAILED', 'CANCELLED'
    )),
    bytes_completed INTEGER DEFAULT 0,
    bytes_total INTEGER DEFAULT 0,
    bytes_uploaded INTEGER DEFAULT 0,
    files_json TEXT NOT NULL DEFAULT '',  -- pilihan file per job
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Dedup sync AniList per episode
CREATE TABLE IF NOT EXISTS sync_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anilist_id INTEGER NOT NULL,
    episode_number INTEGER NOT NULL,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(anilist_id, episode_number)
);

-- Cache respons API (AniList, dll.)
CREATE TABLE IF NOT EXISTS api_cache (
    cache_key TEXT PRIMARY KEY,
    payload TEXT NOT NULL,
    fetched_at INTEGER NOT NULL
);
```

**Perilaku restart:** `DOWNLOADING`/`QUEUED` dipulihkan ke antrian; job `PAUSED` ditandai `FAILED`. Seeding yang tersimpan perlu **Resume** manual di tab Downloads.

---

## 5. Non-Functional Requirements

* **Performance:** Penggunaan memori (RAM) saat idle di bawah 100 MB — **belum diverifikasi** (tidak ada benchmark di repo).
* ~~**Cross-Platform Building:** Kode Go bebas CGO; target Windows (amd64), macOS (universal), Linux (amd64).~~
* ~~**Network Efficiency:** Rate limiter unduh/unggah torrent; mode jaringan system/direct/SOCKS5.~~

---

## 6. Release & CI/CD Strategy

~~Setiap Git Tag baru (`v*`) memicu workflow GitHub Actions untuk merilis biner resmi:~~

| Target Platform | Output File Pattern | Packaging Format |
| --- | --- | --- |
| **Windows** | `miru-{version}-windows-amd64.exe` | Standalone executable (embedded WebView2) |
| **macOS** | `miru-{version}-mac-universal.zip` | Compressed `.app` bundle (unsigned) |
| **Linux** | `miru-{version}-linux-amd64` | Standalone ELF executable |

~~**CI PR/push** (`ci.yml`): `make fmt`, golangci-lint, `make test`, `make lint-fe`, `make typecheck`.~~

~~**In-app auto-update:** cek & terapkan pembaruan dari GitHub Releases (Settings → Updates).~~

---

## 7. UI Screens (implementasi aktual)

| Tab | Status | Fungsi |
| --- | --- | --- |
| Library | ~~done~~ | Poster grid, episode list, impor lokal, bind AniList, play |
| Watching | ~~done~~ | Kelola list AniList, edit skor/progress |
| Search | ~~done~~ | Nyaa / Tokyo Toshokan, inspect & pilih file |
| Downloads | ~~done~~ | Magnet, `.torrent`, antrian, pause/resume, seeding |
| Airing | ~~done~~ | Kalender rilis mingguan |
| Settings | ~~done~~ | MPV, download, AniList, jaringan, updates |

---

## 8. Future Roadmap (Post-v1.0)

* **Discord Rich Presence (RPC):** Menampilkan status anime yang sedang diputar di profil Discord.
* **Shader Injection (Anime4K):** Opsi otomatisasi pengaktifan shader upscaling video pada MPV.
* **Desktop Notifications:** Notifikasi lokal OS ketika unduhan episode selesai di latar belakang.
* **Feed RSS otomatis:** Langganan & polling feed fansub/indexer di background.
* **MPV detection Windows/macOS:** Scan path instalasi umum (Program Files, Homebrew).
* **Proxy HTTP/HTTPS** selain SOCKS5.
* **Sync AniList real-time** (mutasi saat threshold tercapai, tanpa menunggu MPV ditutup).
* **Benchmark RAM idle** & dokumentasi hasil.
