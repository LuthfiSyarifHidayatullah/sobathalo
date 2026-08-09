# Chatbot WhatsApp BKPSDM Kabupaten Bengkayang

Chatbot WhatsApp berbasis menu angka untuk layanan informasi BKPSDM (Badan Kepegawaian dan Pengembangan Sumber Daya Manusia) Kabupaten Bengkayang. Dibangun menggunakan bahasa Go dan library [WhatsMeow](https://github.com/tulir/whatsmeow).

## Fitur

- Menu interaktif berbasis angka (3 level kedalaman)
- 3 bidang pelayanan dengan 6 jenis layanan
- State/sesi per pengguna (mengetahui posisi menu)
- Pencatatan otomatis ke Google Spreadsheet
- Backup CSV lokal jika koneksi ke Google Sheets gagal
- Penyamaran nomor WhatsApp (hashing SHA-256)
- Konfigurasi jawaban terpisah dari logika (mudah diperbarui)
- Thread-safe session management

## Struktur Folder

```
chatbot-bkpsdm/
├── main.go                  # Entry point aplikasi
├── go.mod                   # Go module definition
├── .env.example             # Template environment variables
├── config/
│   ├── config.go            # Loader konfigurasi respons
│   └── responses.json       # Isi jawaban pelayanan (EDIT DI SINI)
├── handler/
│   └── handler.go           # Logika menu dan pemrosesan pesan
├── session/
│   └── session.go           # Manajemen state/sesi pengguna
├── logger/
│   └── logger.go            # Pencatatan ke Google Sheets & CSV
├── scripts/
│   └── google_apps_script.js # Kode Google Apps Script
├── data/                    # (dibuat otomatis saat runtime)
│   ├── whatsapp.db          # Database sesi WhatsApp
│   └── log_backup.csv       # Backup log CSV
└── README.md                # Dokumentasi ini
```

## Struktur Menu

```
MENU UTAMA
├── 1. Bidang Kesejahteraan dan Informasi Kepegawaian
│   ├── 1. Pengajuan Cuti
│   │   ├── 1. Persyaratan pelayanan
│   │   ├── 2. Prosedur pengajuan
│   │   ├── 3. Waktu penyelesaian
│   │   ├── 4. Formulir atau tautan pengajuan
│   │   ├── 5. Cek status pengajuan
│   │   ├── 6. Kendala dan hubungi petugas
│   │   └── 0. Kembali
│   ├── 2. Pencantuman Gelar
│   │   └── (submenu sama seperti di atas)
│   └── 0. Kembali ke Menu Utama
├── 2. Bidang Pengadaan dan Mutasi
│   ├── 1. Kenaikan Pangkat
│   ├── 2. Mutasi PNS
│   └── 0. Kembali ke Menu Utama
├── 3. Bidang Pengembangan SDM Aparatur
│   ├── 1. Tugas Belajar
│   ├── 2. Kenaikan Jabatan Fungsional
│   └── 0. Kembali ke Menu Utama
```

## Panduan Instalasi di Windows

### Prasyarat

1. **Go** versi 1.22 atau lebih baru
   - Download: https://go.dev/dl/
   - Setelah install, buka Command Prompt dan ketik `go version` untuk verifikasi

2. **GCC** (diperlukan untuk SQLite driver)
   - Download TDM-GCC: https://jmeubank.github.io/tdm-gcc/
   - Atau gunakan MSYS2: https://www.msys2.org/
   - Pastikan `gcc` tersedia di PATH

3. **Git** (opsional, untuk clone repository)
   - Download: https://git-scm.com/download/win

### Langkah Instalasi

```batch
REM 1. Buka Command Prompt atau PowerShell

REM 2. Masuk ke folder project
cd chatbot-bkpsdm

REM 3. Salin file konfigurasi environment
copy .env.example .env

REM 4. Edit file .env dengan Notepad atau editor lain
notepad .env

REM 5. Download dependencies
go mod tidy

REM 6. Build aplikasi
go build -o chatbot.exe .

REM 7. Jalankan aplikasi
chatbot.exe
```

### Langkah Pertama Kali (Login WhatsApp)

1. Jalankan `chatbot.exe`
2. QR Code akan muncul di terminal
3. Buka WhatsApp di HP > **Linked Devices** > **Link a Device**
4. Scan QR Code yang tampil di terminal
5. Setelah berhasil, bot akan otomatis aktif
6. Sesi tersimpan di `data/whatsapp.db` (tidak perlu scan ulang)

### Menjalankan Setelah Login

```batch
REM Cukup jalankan:
chatbot.exe
```

Bot akan otomatis terhubung menggunakan sesi yang tersimpan.

## Panduan Menghubungkan Google Spreadsheet

### Langkah 1: Buat Google Spreadsheet

1. Buka https://sheets.google.com
2. Buat spreadsheet baru
3. Beri nama misalnya "Log Chatbot BKPSDM"

### Langkah 2: Buat Google Apps Script

1. Di spreadsheet, buka menu **Extensions** > **Apps Script**
2. Hapus semua kode default
3. Salin dan tempelkan seluruh isi file `scripts/google_apps_script.js`
4. Simpan (Ctrl+S), beri nama project "Chatbot BKPSDM Logger"

### Langkah 3: Deploy sebagai Web App

1. Klik **Deploy** > **New deployment**
2. Klik ikon gear di samping "Select type" > pilih **Web app**
3. Isi konfigurasi:
   - Description: "Chatbot Logger"
   - Execute as: **Me**
   - Who has access: **Anyone**
4. Klik **Deploy**
5. **Salin URL** Web App yang muncul

### Langkah 4: Set Token Rahasia (Opsional tapi Disarankan)

1. Di Apps Script, buka **Project Settings** (ikon gear)
2. Scroll ke bagian **Script Properties**
3. Klik **Add script property**
4. Key: `SECRET_TOKEN`, Value: (buat token acak, misalnya `my-secret-token-2024`)
5. Klik Save

### Langkah 5: Konfigurasi .env

Edit file `.env` di folder project:

```env
GOOGLE_SCRIPT_URL=https://script.google.com/macros/s/XXXXX/exec
GOOGLE_SCRIPT_TOKEN=my-secret-token-2024
GOOGLE_SCRIPT_TIMEOUT=10
```

### Verifikasi

Buka URL Web App di browser. Jika muncul response JSON `{"status":"active",...}` maka konfigurasi berhasil.

## Mengisi Jawaban Resmi Pelayanan

### File yang Perlu Diedit

Edit file **`config/responses.json`** untuk mengubah isi jawaban setiap pelayanan.

### Struktur File

```json
{
  "bidang": {
    "kesejahteraan": {
      "pelayanan": {
        "cuti": {
          "info": {
            "persyaratan": "Isi persyaratan resmi di sini...",
            "prosedur": "Isi prosedur resmi di sini...",
            "waktu": "Isi waktu penyelesaian resmi...",
            "formulir": "Isi link formulir resmi...",
            "status": "Isi cara cek status resmi...",
            "kendala": "Isi kontak petugas resmi..."
          }
        }
      }
    }
  }
}
```

### Pelayanan yang Tersedia

| Bidang | Key Bidang | Pelayanan | Key Pelayanan |
|--------|-----------|-----------|--------------|
| Kesejahteraan dan Informasi Kepegawaian | `kesejahteraan` | Pengajuan Cuti | `cuti` |
| Kesejahteraan dan Informasi Kepegawaian | `kesejahteraan` | Pencantuman Gelar | `gelar` |
| Pengadaan dan Mutasi | `pengadaan` | Kenaikan Pangkat | `pangkat` |
| Pengadaan dan Mutasi | `pengadaan` | Mutasi PNS | `mutasi` |
| Pengembangan SDM Aparatur | `pengembangan` | Tugas Belajar | `tubel` |
| Pengembangan SDM Aparatur | `pengembangan` | Kenaikan Jabatan Fungsional | `fungsional` |

### Cara Edit

1. Buka `config/responses.json` dengan text editor
2. Cari pelayanan yang ingin diisi (misal `"cuti"`)
3. Ganti teks placeholder di dalam `"info"` dengan informasi resmi
4. Simpan file
5. Restart chatbot agar perubahan berlaku

### Tips Format Teks WhatsApp

- `*teks*` → **bold**
- `_teks_` → _italic_
- `~teks~` → ~~strikethrough~~
- `\n` → baris baru

## Contoh Percakapan

```
PENGGUNA: Halo
BOT: Halo! Selamat datang di Layanan Informasi *BKPSDM Kabupaten Bengkayang* 🏛️

     Silakan pilih menu dengan mengetik angka:

     1️⃣ Bidang Kesejahteraan dan Informasi Kepegawaian
     2️⃣ Bidang Pengadaan dan Mutasi
     3️⃣ Bidang Pengembangan SDM Aparatur

     📌 Ketik *menu* kapan saja untuk kembali ke menu utama.
     📌 Ketik *0* untuk kembali ke menu sebelumnya.

PENGGUNA: 1
BOT: 📋 *Bidang Kesejahteraan dan Informasi Kepegawaian*

     Pilih layanan:
     1️⃣ Pengajuan Cuti
     2️⃣ Pencantuman Gelar
     0️⃣ Kembali ke Menu Utama

PENGGUNA: 1
BOT: 📋 *Pengajuan Cuti*

     Pilih informasi yang dibutuhkan:
     1️⃣ Persyaratan pelayanan
     2️⃣ Prosedur pengajuan
     3️⃣ Waktu penyelesaian
     4️⃣ Formulir atau tautan pengajuan
     5️⃣ Cek status pengajuan
     6️⃣ Kendala dan hubungi petugas
     0️⃣ Kembali

PENGGUNA: 1
BOT: 📝 *Persyaratan Pengajuan Cuti:*

     1. Surat permohonan cuti yang ditandatangani pemohon
     2. Fotokopi SK terakhir
     3. Sisa cuti tahun berjalan masih tersedia
     4. Persetujuan atasan langsung
     5. Formulir cuti yang telah diisi lengkap

     _(Informasi ini adalah placeholder. Isi resmi akan diperbarui oleh admin BKPSDM.)_

     ---
     📌 Ketik *0* untuk kembali ke menu pelayanan
     📌 Ketik *menu* untuk ke menu utama

PENGGUNA: 0
BOT: 📋 *Pengajuan Cuti*

     Pilih informasi yang dibutuhkan:
     1️⃣ Persyaratan pelayanan
     ...

PENGGUNA: menu
BOT: Halo! Selamat datang di Layanan Informasi *BKPSDM Kabupaten Bengkayang* 🏛️
     ...
```

## Keamanan

- ✅ Nomor WhatsApp pengguna disamarkan dengan hash SHA-256 sebelum dicatat
- ✅ URL webhook dan token disimpan di environment variable (bukan di source code)
- ✅ Bot tidak meminta data sensitif (KTP, PIN, password)
- ✅ Timeout pada pengiriman data ke Google Sheets
- ✅ Mutex untuk akses concurrent ke session dan CSV
- ✅ Kegagalan Google Sheets tidak menghentikan chatbot

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| QR Code tidak muncul | Pastikan terminal mendukung karakter Unicode. Coba Windows Terminal |
| Error "gcc not found" | Install TDM-GCC dan pastikan ada di PATH |
| Bot tidak membalas | Cek apakah pesan dikirim ke chat pribadi (bukan grup) |
| Google Sheets tidak terisi | Cek URL di .env, pastikan Web App sudah di-deploy |
| "Gagal memuat konfigurasi" | Pastikan file `config/responses.json` ada dan valid JSON |

## Lisensi

Proyek ini dibuat untuk kebutuhan internal BKPSDM Kabupaten Bengkayang.
