# Chatbot WhatsApp BKPSDM Kabupaten Bengkayang

Chatbot WhatsApp berbasis menu angka untuk layanan informasi BKPSDM (Badan Kepegawaian dan Pengembangan Sumber Daya Manusia) Kabupaten Bengkayang. Dibangun menggunakan bahasa Go dan library [WhatsMeow](https://github.com/tulir/whatsmeow).

## Fitur

- Menu interaktif berbasis angka, bertingkat dinamis (tree)
- 3 bidang dengan puluhan layanan & FAQ resmi (29 pelayanan + 16 FAQ)
- Tiga tipe konten: pelayanan (berkas), FAQ (tanya-jawab), dan sub-kategori
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

Menu bersifat **bertingkat dinamis** (tree). Setiap item bisa berupa salah satu dari tiga tipe:

- **`menu`** — grup/sub-kategori yang berisi pilihan lain.
- **`pelayanan`** — layanan berkas; menampilkan submenu info: *Persyaratan → Cara Pengajuan → Hubungi Petugas* (hanya bagian yang terisi yang ditampilkan).
- **`faq`** — tanya-jawab; jawaban langsung ditampilkan tanpa submenu.

```
MENU UTAMA
├── 1. Bidang Kesejahteraan dan Informasi Kepegawaian
│   ├── 1. Surat Keterangan Bebas Tindakan Disiplin   [pelayanan]
│   ├── 2. Penghargaan Satya Lencana Karya Satya      [pelayanan]
│   ├── 3. Pengajuan Klaim Tapera                     [pelayanan]
│   ├── 4. Kapan dana klaim Tapera masuk?             [faq]
│   ├── 5. Berapa nominal besaran Tapera?             [faq]
│   ├── 6. SKP / Sasaran Kinerja Pegawai              [menu] (4 FAQ)
│   ├── 7. Angka Kredit                               [menu] (5 FAQ)
│   ├── 8. e-Kinerja Bengkayang                       [menu] (2 FAQ + 2 pelayanan)
│   └── 9. DMS SIASN                                  [menu] (3 FAQ)
├── 2. Bidang Pengadaan dan Mutasi Pegawai
│   ├── 1. Pengajuan Permintaan Pensiun BUP           [pelayanan]
│   ├── 2. Pengajuan Permintaan Pensiun Janda/Duda    [pelayanan]
│   ├── 3. Permohonan Mutasi Keluar Pemkab            [pelayanan]
│   ├── 4. Permohonan Mutasi Masuk Pemkab             [pelayanan]
│   ├── 5. Permohonan Mutasi Antar Perangkat Daerah   [pelayanan]
│   ├── 6. Usul Kenaikan Gaji Berkala (KGB)           [pelayanan]
│   ├── 7. Pengajuan Taspen                           [pelayanan]
│   ├── 8. Pengajuan Peninjauan Masa Kerja (PMK)      [pelayanan]
│   └── 9. Usul Kenaikan Pangkat                      [menu] (9 pelayanan)
├── 3. Bidang Pengembangan Sumber Daya Manusia
│   ├── 1. Rekomendasi Seleksi Tugas Belajar          [pelayanan]
│   ├── 2. Berkas Administrasi Calon Tugas Belajar    [pelayanan]
│   ├── 3. Keterangan Menyelesaikan Pendidikan        [pelayanan]
│   ├── 4. Pengangkatan Pertama Jabatan Fungsional    [pelayanan]
│   ├── 5. Penetapan Kenaikan Jabatan Fungsional      [pelayanan]
│   ├── 6. Pengangkatan JF Perpindahan Jabatan Lain   [pelayanan]
│   └── 7. Pemberhentian JF Karena Mutasi             [pelayanan]
```

Navigasi: ketik **`0`** untuk kembali ke menu sebelumnya, dan **`menu`** untuk kembali ke menu utama.

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

### Struktur File (berbasis pohon/tree)

Seluruh menu didefinisikan sebagai pohon `node`. Ada **tiga tipe node**:

**1. Node `menu`** — grup berisi pilihan lain. Setiap anak diberi nomor (`"1"`, `"2"`, ...) berurutan:

```json
{
  "type": "menu",
  "judul": "Judul yang tampil di menu induk",
  "deskripsi": "📋 *Judul*\n\nSilakan pilih:",
  "children": {
    "1": { ... node anak ... },
    "2": { ... node anak ... }
  }
}
```

**2. Node `pelayanan`** — layanan berkas. Isi bagian yang relevan; bagian yang dikosongkan (`""`) tidak akan ditampilkan sebagai pilihan:

```json
{
  "type": "pelayanan",
  "judul": "Nama Pelayanan",
  "info": {
    "syarat": "📝 *Syarat:*\n\n1. ...\n2. ...",
    "cara_pengajuan": "📋 *Cara Pengajuan:*\n\n• ...",
    "petugas": "📞 *Kontak Petugas:*\n\n👤 Nama\n📱 08xx"
  }
}
```

**3. Node `faq`** — tanya-jawab, jawaban langsung tampil:

```json
{
  "type": "faq",
  "judul": "Pertanyaan singkat",
  "jawaban": "❓ *Pertanyaan?*\n\n✅ Jawaban..."
}
```

### Cara Edit / Menambah Pelayanan Baru

1. Buka `config/responses.json` dengan text editor.
2. **Mengubah isi:** cari node berdasarkan `"judul"`, lalu ubah teks di `syarat`, `cara_pengajuan`, `petugas` (untuk pelayanan) atau `jawaban` (untuk FAQ).
3. **Menambah item baru:** tambahkan entri di `children` node induk dengan **nomor urut berikutnya** (harus berurutan tanpa lompat, mis. setelah `"7"` gunakan `"8"`).
4. Pastikan JSON tetap valid (koma, kurung). Uji dengan: `python -m json.tool config/responses.json`.
5. Restart chatbot agar perubahan berlaku.

> ⚠️ **Penting:** Key anak pada setiap `menu` harus berurutan mulai dari `"1"`. Nomor inilah yang diketik pengguna di WhatsApp.

### Tips Format Teks WhatsApp

- `*teks*` → **bold**
- `_teks_` → _italic_
- `~teks~` → ~~strikethrough~~
- `\n` → baris baru

## Contoh Percakapan

### Contoh 1 — Pelayanan berkas (dari menu utama sampai info)

```
PENGGUNA: Halo
BOT: Halo! 👋 Selamat datang di *Layanan Informasi BKPSDM Kabupaten Bengkayang* 🏛️

     Silakan pilih bidang layanan dengan mengetik angka:
     1️⃣ Bidang Kesejahteraan dan Informasi Kepegawaian
     2️⃣ Bidang Pengadaan dan Mutasi Pegawai
     3️⃣ Bidang Pengembangan Sumber Daya Manusia

     📌 Ketik *0* untuk kembali ke menu sebelumnya.
     📌 Ketik *menu* kapan saja untuk kembali ke menu utama.

PENGGUNA: 1
BOT: 📋 *Bidang Kesejahteraan dan Informasi Kepegawaian*

     Silakan pilih layanan:
     1️⃣ Surat Keterangan Bebas Tindakan Disiplin
     2️⃣ Penghargaan Satya Lencana Karya Satya
     3️⃣ Pengajuan Klaim Tapera
     4️⃣ Kapan dana klaim Tapera masuk ke rekening?
     5️⃣ Berapa nominal besaran Tapera saya?
     6️⃣ SKP / Sasaran Kinerja Pegawai
     7️⃣ Angka Kredit
     8️⃣ e-Kinerja Bengkayang
     9️⃣ DMS SIASN

     0️⃣ Kembali
     📌 Ketik *menu* untuk kembali ke menu utama.

PENGGUNA: 1
BOT: 📋 *Surat Keterangan Bebas Tindakan Disiplin*

     Pilih informasi yang dibutuhkan:
     1️⃣ Persyaratan
     2️⃣ Cara Pengajuan
     3️⃣ Hubungi Petugas

     0️⃣ Kembali
     📌 Ketik *menu* untuk kembali ke menu utama.

PENGGUNA: 1
BOT: 📝 *Syarat Surat Keterangan Bebas Tindakan Disiplin:*

     1. SK PNS
     2. SK CPNS
     3. Surat Rekomendasi Berjenjang
     4. Surat Bebas Tindak Pidana dari Pengadilan Negeri

     ---
     📌 Ketik *0* untuk kembali ke pilihan pelayanan
     📌 Ketik *menu* untuk ke menu utama
```

### Contoh 2 — Sub-kategori dan FAQ (jawaban langsung)

```
PENGGUNA: 1        (dari menu utama → Bidang Kesejahteraan)
PENGGUNA: 7        (→ Angka Kredit, sebuah sub-kategori)
BOT: 📋 *Angka Kredit*

     Silakan pilih pertanyaan:
     1️⃣ Pesan "File TTD Basah Belum Diunggah"
     2️⃣ Angka Kredit dari SIASN tidak terakumulasi
     3️⃣ Angka Kredit yang diklaim tidak terakumulasi
     4️⃣ Tombol Tambah PAK tidak muncul
     5️⃣ Error "Request failed with status code 419"

     0️⃣ Kembali
     📌 Ketik *menu* untuk kembali ke menu utama.

PENGGUNA: 5
BOT: ❓ *Kenapa muncul pesan error "Request failed with status code 419"?*

     ✅ Silakan logout lalu login lagi.

     ---
     📌 Ketik nomor lain untuk pertanyaan lain
     📌 Ketik *0* untuk kembali ke menu sebelumnya
     📌 Ketik *menu* untuk ke menu utama.
```

> Catatan: pada FAQ, setelah jawaban tampil pengguna **tetap** di sub-kategori yang sama, sehingga bisa langsung mengetik nomor pertanyaan lain.

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
