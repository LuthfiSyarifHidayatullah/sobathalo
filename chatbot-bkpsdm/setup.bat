@echo off
setlocal enabledelayedexpansion
REM ============================================================
REM  SETUP - Chatbot WhatsApp BKPSDM Kabupaten Bengkayang
REM ------------------------------------------------------------
REM  Jalankan SEKALI saja untuk persiapan awal:
REM    - Cek Go & GCC
REM    - Membuat .env
REM    - Mengunduh dependency
REM    - Build aplikasi
REM  Setelah setup selesai, cukup gunakan run.bat untuk menjalankan.
REM ============================================================

cd /d "%~dp0"
title Setup Chatbot BKPSDM
color 0B

echo ============================================================
echo   SETUP AWAL CHATBOT BKPSDM BENGKAYANG
echo ============================================================
echo.

set "SETUP_FAIL="

REM ---------- Cek Go ----------
echo [Cek] Go...
where go >nul 2>nul
if %errorlevel% neq 0 goto go_missing
for /f "tokens=3" %%v in ('go version') do set "GOVER=%%v"
echo   [OK] Go !GOVER!
goto cek_gcc

:go_missing
color 0C
echo   [BELUM ADA] Install Go dari: https://go.dev/dl/
set "SETUP_FAIL=1"

:cek_gcc
echo.
echo [Cek] GCC (diperlukan untuk database SQLite)...
where gcc >nul 2>nul
if %errorlevel% neq 0 goto gcc_missing
echo   [OK] GCC terpasang
goto cek_selesai

:gcc_missing
color 0E
echo   [BELUM ADA] Install TDM-GCC dari: https://jmeubank.github.io/tdm-gcc/
echo               (Diperlukan agar build tidak gagal)

:cek_selesai
echo.
if defined SETUP_FAIL goto setup_gagal

REM ---------- Buat .env ----------
echo [Konfigurasi] File .env...
if exist ".env" goto env_ada
if not exist ".env.example" goto env_lanjut
copy ".env.example" ".env" >nul
echo   File .env dibuat. Membuka untuk diedit...
notepad .env
goto env_lanjut

:env_ada
echo   [OK] .env sudah ada

:env_lanjut
echo.

REM ---------- Folder data ----------
if not exist "data" mkdir data
echo [OK] Folder data siap.
echo.

REM ---------- Dependency ----------
echo [Proses] Mengunduh dependency (go mod tidy)...
go mod tidy
if %errorlevel% neq 0 goto err_deps
echo   [OK] Dependency siap.
echo.

REM ---------- Build ----------
echo [Proses] Build aplikasi...
go build -o chatbot.exe .
if %errorlevel% neq 0 goto err_build
echo   [OK] Build berhasil.
echo.

color 0A
echo ============================================================
echo   SETUP SELESAI!
echo   Untuk menjalankan chatbot, double-click file: run.bat
echo ============================================================
pause
goto end

REM ==================== BAGIAN ERROR ====================

:setup_gagal
echo ------------------------------------------------------------
echo   Lengkapi dulu prasyarat di atas, lalu jalankan setup.bat lagi.
echo ------------------------------------------------------------
pause
goto end

:err_deps
color 0C
echo   [ERROR] Gagal mengunduh dependency. Cek koneksi internet.
pause
goto end

:err_build
color 0C
echo   [ERROR] Build gagal. Pastikan GCC terpasang.
pause
goto end

:end
endlocal
