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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogEntry berisi data yang akan dicatat (format ringkas & mudah dibaca)
type LogEntry struct {
	Waktu     string `json:"waktu"`
	Pengguna  string `json:"pengguna"`
	Bidang    string `json:"bidang"`
	Pelayanan string `json:"pelayanan"`
	InfoDiminta string `json:"info_diminta"`
	Status    string `json:"status"`
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

// sendToGoogleSheets mengirim data ke Google Apps Script Web App
func (l *Logger) sendToGoogleSheets(entry LogEntry) {
	if l.scriptURL == "" {
		return
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("[LOGGER] Gagal marshal JSON: %v\n", err)
		return
	}

	// Custom HTTP client yang mengikuti redirect sambil tetap POST
	// (Google Apps Script me-redirect ke googleusercontent.com)
	client := &http.Client{
		Timeout: l.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Pertahankan method POST dan body saat redirect
			if len(via) >= 10 {
				return fmt.Errorf("terlalu banyak redirect")
			}
			req.Method = via[0].Method
			req.Header.Set("Content-Type", "application/json")
			if via[0].GetBody != nil {
				body, err := via[0].GetBody()
				if err == nil {
					req.Body = body
				}
			}
			return nil
		},
	}

	req, err := http.NewRequest("POST", l.scriptURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("[LOGGER] Gagal membuat request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	// Simpan body agar bisa dipakai ulang saat redirect
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(payload)), nil
	}

	if l.scriptToken != "" {
		req.Header.Set("Authorization", "Bearer "+l.scriptToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[LOGGER] Gagal mengirim ke Google Sheets: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[LOGGER] Google Sheets response status: %d\n", resp.StatusCode)
	}
}

// Close menutup file CSV
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.csvWriter.Flush()
	return l.csvFile.Close()
}
