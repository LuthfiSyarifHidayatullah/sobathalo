@echo off
REM ============================================================
REM  RUN AS ADMINISTRATOR - Chatbot BKPSDM Bengkayang
REM ------------------------------------------------------------
REM  Menjalankan chatbot dengan hak Administrator.
REM  Script otomatis meminta izin UAC (tidak perlu klik kanan
REM  "Run as administrator" manual).
REM
REM  Gunakan ini HANYA jika run.bat biasa bermasalah karena izin
REM  (mis. gagal menulis file/folder atau firewall).
REM ============================================================

REM ---------- Cek apakah sudah running sebagai admin ----------
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Meminta izin Administrator...
    REM Self-elevate: jalankan ulang file ini dengan hak admin via UAC
    powershell -NoProfile -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
    exit /b
)

REM ---------- Sudah admin, lanjut jalankan run.bat ----------
cd /d "%~dp0"
title Chatbot BKPSDM (Administrator)

echo ============================================================
echo   MODE ADMINISTRATOR AKTIF
echo ============================================================
echo.

call "%~dp0run.bat"
