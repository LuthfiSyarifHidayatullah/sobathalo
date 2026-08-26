@echo off
REM ============================================================
REM  STOP - Menghentikan Chatbot BKPSDM yang sedang berjalan
REM ============================================================

title Stop Chatbot BKPSDM
echo Menghentikan proses chatbot.exe...

taskkill /IM chatbot.exe /F >nul 2>nul
if %errorlevel% equ 0 (
    echo [OK] Chatbot berhasil dihentikan.
) else (
    echo [INFO] Tidak ada proses chatbot.exe yang sedang berjalan.
)

echo.
pause
