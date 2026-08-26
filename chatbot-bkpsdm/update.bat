@echo off
setlocal
REM ============================================================
REM  UPDATE - Perbarui library WhatsMeow ke versi TERBARU
REM ------------------------------------------------------------
REM  Jalankan file ini jika muncul error:
REM    "Client outdated (405) connect failure"
REM
REM  WhatsApp sering memperbarui protokolnya. Error 405 berarti
REM  versi library WhatsMeow sudah usang dan perlu di-update.
REM  Script ini mengambil versi TERBARU lalu build ulang.
REM ============================================================

cd /d "%~dp0"
title Update WhatsMeow - Chatbot BKPSDM
color 0B

echo ============================================================
echo   UPDATE LIBRARY WHATSMEOW KE VERSI TERBARU
echo ============================================================
echo.

REM ---------- Cek Go ----------
where go >nul 2>nul
if %errorlevel% neq 0 goto no_go

echo [1/4] Mengambil WhatsMeow versi terbaru...
go get go.mau.fi/whatsmeow@latest
if %errorlevel% neq 0 goto err_get
echo       OK.
echo.

echo [2/4] Merapikan dependency (go mod tidy)...
go mod tidy
if %errorlevel% neq 0 goto err_tidy
echo       OK.
echo.

echo [3/4] Menghapus build lama...
if exist "chatbot.exe" del /f /q "chatbot.exe"
echo       OK.
echo.

echo [4/4] Build ulang aplikasi...
go build -o chatbot.exe .
if %errorlevel% neq 0 goto err_build
echo       OK - Build berhasil.
echo.

color 0A
echo ============================================================
echo   UPDATE SELESAI!
echo   Sekarang jalankan run.bat untuk memakai versi terbaru.
echo ============================================================
pause
goto end

:no_go
color 0C
echo [ERROR] Go tidak ditemukan. Install dari https://go.dev/dl/
pause
goto end

:err_get
color 0C
echo [ERROR] Gagal mengambil WhatsMeow terbaru. Cek koneksi internet.
pause
goto end

:err_tidy
color 0C
echo [ERROR] go mod tidy gagal. Cek koneksi internet.
pause
goto end

:err_build
color 0C
echo [ERROR] Build gagal. Pastikan GCC (TDM-GCC) terpasang.
pause
goto end

:end
endlocal
