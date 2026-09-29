package logger

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogEntry berisi data yang akan dicatat (format ringkas & mudah dibaca)
type LogEntry struct {
	Waktu       string `json:"waktu"`
	Pengguna    string `json:"pengguna"`
	Bidang      string `json:"bidang"`
	Pelayanan   string `json:"pelayanan"`
	InfoDiminta string `json:"info_diminta"`
	Status      string `json:"status"`
}

// Logger mengelola pencatatan ke Google Sheets dan CSV
type Logger struct {
	mu          sync.Mutex
	scriptURL   string
	scriptToken string
	timeout     time.Duration
	csvPath     string
	csvFile     *os.File
	csvWriter   *csv.Writer
}

// NewLogger membuat Logger baru
func NewLogger(scriptURL, scriptToken string, timeout time.Duration, csvPath string) (*Logger, error) {
	// Pastikan direktori CSV ada
	dir := filepath.Dir(csvPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori CSV: %w", err)
	}

	// Buka atau buat file CSV
	fileExists := false
	if _, err := os.Stat(csvPath); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file CSV: %w", err)
	}

	writer := csv.NewWriter(file)

	// Tulis header jika file baru
	if !fileExists {
		header := []string{
			"Waktu", "Pengguna", "Bidang", "Pelayanan", "Info Diminta", "Status",
		}
		if err := writer.Write(header); err != nil {
			file.Close()
			return nil, fmt.Errorf("gagal menulis header CSV: %w", err)
		}
		writer.Flush()
	}

	return &Logger{
		scriptURL:   scriptURL,
		scriptToken: scriptToken,
		timeout:     timeout,
		csvPath:     csvPath,
		csvFile:     file,
		csvWriter:   writer,
	}, nil
}

// HashUserID menyamarkan nomor WhatsApp dengan SHA-256 (dipakai internal, tidak ditampilkan)
func HashUserID(userID string) string {
	hash := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(hash[:8])
}

// FormatStatus mengubah status teknis menjadi label emoji yang mudah dibaca
func FormatStatus(status string) string {
	switch status {
	case "terjawab":
		return "✅ Terjawab"
	case "tidak_dikenali":
		return "❓ Tidak dikenali"
	case "dialihkan_ke_petugas":
		return "📞 Dialihkan ke petugas"
	default:
		return status
	}
}

// SingkatBidang menghilangkan kata "Bidang" di depan nama bidang agar lebih ringkas
func SingkatBidang(bidang string) string {
	bidang = strings.TrimPrefix(bidang, "Bidang ")
	return bidang
}

// Log mencatat pesan ke Google Sheets dan CSV backup
func (l *Logger) Log(entry LogEntry) {
	// Catat ke CSV terlebih dahulu (sebagai backup)
	l.writeCSV(entry)

	// Kirim ke Google Sheets secara async (non-blocking)
	go l.sendToGoogleSheets(entry)
}

// writeCSV menulis entry ke file CSV
func (l *Logger) writeCSV(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	record := []string{
		entry.Waktu,
		entry.Pengguna,
		entry.Bidang,
		entry.Pelayanan,
		entry.InfoDiminta,
		entry.Status,
	}

	if err := l.csvWriter.Write(record); err != nil {
		fmt.Printf("[LOGGER] Gagal menulis ke CSV: %v\n", err)
		return
	}
	l.csvWriter.Flush()
}

// sendToGoogleSheets mengirim data ke Google Apps Script Web App.
//
// Catatan penting tentang Google Apps Script:
//   - Web App membalas dengan HTTP 302 redirect ke script.googleusercontent.com.
//     http.Client default Go akan mengikuti redirect ini secara otomatis.
//   - Custom header (mis. Authorization) TIDAK diteruskan/dapat dibaca oleh
//     Apps Script. Karena itu token dikirim sebagai bagian dari payload JSON.
func (l *Logger) sendToGoogleSheets(entry LogEntry) {
	if l.scriptURL == "" {
		return
	}

	// Bungkus entry + token dalam satu payload. Token dibaca oleh Apps Script
	// dari field "token" (lihat scripts/google_apps_script.js).
	payloadObj := map[string]interface{}{
		"waktu":        entry.Waktu,
		"pengguna":     entry.Pengguna,
		"bidang":       entry.Bidang,
		"pelayanan":    entry.Pelayanan,
		"info_diminta": entry.InfoDiminta,
		"status":       entry.Status,
	}
	if l.scriptToken != "" {
		payloadObj["token"] = l.scriptToken
	}

	payload, err := json.Marshal(payloadObj)
	if err != nil {
		fmt.Printf("[LOGGER] Gagal marshal JSON: %v\n", err)
		return
	}

	client := &http.Client{Timeout: l.timeout}

	// Content-Type text/plain menghindari CORS preflight dan diterima Apps Script
	// (e.postData.contents tetap berisi string JSON yang kita kirim).
	req, err := http.NewRequest("POST", l.scriptURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("[LOGGER] Gagal membuat request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain;charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[LOGGER] Gagal mengirim ke Google Sheets: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[LOGGER] Google Sheets HTTP %d: %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
		return
	}

	// Apps Script membalas JSON {status: success|error, ...}. Deteksi error logis.
	if strings.Contains(string(body), "\"status\":\"error\"") ||
		strings.Contains(string(body), "\"status\": \"error\"") {
		fmt.Printf("[LOGGER] Google Sheets menolak data: %s\n", strings.TrimSpace(string(body)))
		return
	}
}

// TestConnection mengirim GET ke script URL untuk memverifikasi Web App aktif.
// Dipanggil saat startup agar masalah konfigurasi terlihat lebih awal.
func (l *Logger) TestConnection() {
	if l.scriptURL == "" {
		fmt.Println("[LOGGER] GOOGLE_SCRIPT_URL kosong — pencatatan hanya ke CSV lokal.")
		return
	}
	if _, err := url.ParseRequestURI(l.scriptURL); err != nil {
		fmt.Printf("[LOGGER] GOOGLE_SCRIPT_URL tidak valid: %v\n", err)
		return
	}

	client := &http.Client{Timeout: l.timeout}
	resp, err := client.Get(l.scriptURL)
	if err != nil {
		fmt.Printf("[LOGGER] Tidak bisa menghubungi Google Script: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))

	if resp.StatusCode == http.StatusOK && strings.Contains(string(body), "active") {
		fmt.Println("[LOGGER] ✅ Koneksi Google Spreadsheet OK (Web App aktif).")
	} else {
		fmt.Printf("[LOGGER] ⚠️  Google Script membalas status %d. Pastikan deploy sebagai Web App dengan akses 'Anyone'.\n", resp.StatusCode)
	}
}

// Close menutup file CSV
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.csvWriter.Flush()
	return l.csvFile.Close()
}
