@echo off
REM ============================================================
REM  BUAT SHORTCUT DESKTOP - Chatbot BKPSDM Bengkayang
REM ------------------------------------------------------------
REM  Membuat ikon "Chatbot BKPSDM" di Desktop yang langsung
REM  menjalankan run.bat. Cukup dijalankan SEKALI.
REM ============================================================

cd /d "%~dp0"
title Buat Shortcut Desktop - Chatbot BKPSDM
color 0B

echo ============================================================
echo   MEMBUAT SHORTCUT DI DESKTOP
echo ============================================================
echo.

set "TARGET=%~dp0run.bat"
set "ICONFILE=%~dp0chatbot.exe"
set "WORKDIR=%~dp0"
set "SHORTCUT=%USERPROFILE%\Desktop\Chatbot BKPSDM.lnk"

REM Buat shortcut menggunakan PowerShell (tersedia di semua Windows modern)
powershell -NoProfile -Command ^
  "$ws = New-Object -ComObject WScript.Shell;" ^
  "$sc = $ws.CreateShortcut('%SHORTCUT%');" ^
  "$sc.TargetPath = '%TARGET%';" ^
  "$sc.WorkingDirectory = '%WORKDIR%';" ^
  "if (Test-Path '%ICONFILE%') { $sc.IconLocation = '%ICONFILE%,0' };" ^
  "$sc.Description = 'Jalankan Chatbot WhatsApp BKPSDM Kabupaten Bengkayang';" ^
  "$sc.Save()"

if %errorlevel% equ 0 (
    color 0A
    echo   [OK] Shortcut berhasil dibuat di Desktop dengan nama:
    echo        "Chatbot BKPSDM"
    echo.
    echo   Sekarang Anda cukup double-click ikon tersebut di Desktop
    echo   untuk menjalankan chatbot.
) else (
    color 0C
    echo   [ERROR] Gagal membuat shortcut.
    echo           Coba jalankan file ini sebagai Administrator.
)

echo.
pause
