@echo off
setlocal enabledelayedexpansion
REM ============================================================
REM  RUN ALL - Chatbot WhatsApp BKPSDM Kabupaten Bengkayang
REM ------------------------------------------------------------
REM  Cukup DOUBLE-CLICK file ini untuk menjalankan chatbot.
REM  Script akan otomatis:
REM    1. Cek Go terpasang
REM    2. Membuat file .env dari .env.example (jika belum ada)
REM    3. Membuat folder data
REM    4. Mengunduh dependency (go mod tidy)
REM    5. Build aplikasi
REM    6. Menjalankan chatbot
REM ============================================================

REM Pindah ke folder tempat file .bat ini berada
cd /d "%~dp0"

title Chatbot BKPSDM Bengkayang
color 0A

echo ============================================================
echo   CHATBOT WHATSAPP BKPSDM KABUPATEN BENGKAYANG
echo ============================================================
echo.

REM ---------- 1. Cek Go ----------
echo [1/6] Memeriksa instalasi Go...
where go >nul 2>nul
if %errorlevel% neq 0 goto no_go
for /f "tokens=3" %%v in ('go version') do set "GOVER=%%v"
echo       OK - Go !GOVER! terpasang.
echo.

REM ---------- 2. Cek / buat file .env ----------
echo [2/6] Memeriksa file konfigurasi .env...
if exist ".env" goto env_ready
if not exist ".env.example" goto env_missing_example

copy ".env.example" ".env" >nul
echo       File .env dibuat dari .env.example.
echo.
color 0E
echo ------------------------------------------------------------
echo   PERHATIAN: Silakan edit file .env untuk mengisi
echo   GOOGLE_SCRIPT_URL dan GOOGLE_SCRIPT_TOKEN.
echo   (Chatbot tetap bisa jalan tanpa itu, data disimpan ke CSV.)
echo ------------------------------------------------------------
color 0A
echo.
choice /c YT /n /m "Buka file .env sekarang untuk diedit? (Y=Ya, T=Tidak): "
if errorlevel 2 goto env_done
notepad .env
goto env_done

:env_missing_example
echo       [WARNING] .env.example tidak ditemukan, melewati langkah ini.
goto env_done

:env_ready
echo       OK - File .env sudah ada.

:env_done
echo.

REM ---------- 3. Buat folder data ----------
echo [3/6] Menyiapkan folder data...
if not exist "data" mkdir data
echo       OK - Folder data siap.
echo.

REM ---------- 4. Unduh dependency ----------
echo [4/6] Mengunduh dependency (go mod tidy)...
echo       (Proses ini mungkin memakan waktu saat pertama kali dijalankan)
go mod tidy
if %errorlevel% neq 0 goto err_deps
echo       OK - Dependency siap.
echo.

REM ---------- 5. Build aplikasi ----------
echo [5/6] Membangun aplikasi (build)...
go build -o chatbot.exe .
if %errorlevel% neq 0 goto err_build
echo       OK - Build berhasil (chatbot.exe).
echo.

REM ---------- 6. Jalankan ----------
echo [6/6] Menjalankan chatbot...
echo ============================================================
echo   Jika ini pertama kali, akan muncul QR Code.
echo   Scan lewat WhatsApp ^> Perangkat Tertaut ^> Tautkan Perangkat
echo.
echo   Tekan Ctrl+C untuk menghentikan chatbot.
echo ============================================================
echo.
chatbot.exe

echo.
echo ============================================================
echo   Chatbot berhenti.
echo ============================================================
pause
goto end

REM ==================== BAGIAN ERROR ====================

:no_go
color 0C
echo.
echo [ERROR] Go tidak ditemukan di sistem Anda.
echo         Silakan install Go terlebih dahulu dari:
echo         https://go.dev/dl/
echo.
echo         Setelah install, TUTUP jendela ini dan jalankan lagi run.bat
echo.
pause
goto end

:err_deps
color 0C
echo.
echo [ERROR] Gagal mengunduh dependency.
echo         Pastikan koneksi internet aktif, lalu coba lagi.
echo         Jika ada error soal 'gcc', install TDM-GCC dari:
echo         https://jmeubank.github.io/tdm-gcc/
echo.
pause
goto end

:err_build
color 0C
echo.
echo [ERROR] Gagal build aplikasi.
echo         Jika error menyebut 'gcc', install TDM-GCC dari:
echo         https://jmeubank.github.io/tdm-gcc/
echo         lalu jalankan run.bat lagi.
echo.
pause
goto end

:end
endlocal
