package logger

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogEntry berisi data yang akan dicatat
type LogEntry struct {
	Timestamp   string `json:"timestamp"`
	MessageID   string `json:"message_id"`
	UserIDHash  string `json:"user_id_hash"`
	PushName    string `json:"push_name"`
	Bidang      string `json:"bidang"`
	Pelayanan   string `json:"pelayanan"`
	JenisInfo   string `json:"jenis_info"`
	PesanAsli   string `json:"pesan_asli"`
	Status      string `json:"status"` // "terjawab", "tidak_dikenali", "dialihkan_ke_petugas"
}

// Logger mengelola pencatatan ke Google Sheets dan CSV
type Logger struct {
	mu            sync.Mutex
	scriptURL     string
	scriptToken   string
	timeout       time.Duration
	csvPath       string
	csvFile       *os.File
	csvWriter     *csv.Writer
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
			"Timestamp", "MessageID", "UserIDHash", "PushName",
			"Bidang", "Pelayanan", "JenisInfo", "PesanAsli", "Status",
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

// HashUserID menyamarkan nomor WhatsApp dengan SHA-256
func HashUserID(userID string) string {
	hash := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(hash[:8]) // Ambil 8 byte pertama (16 karakter hex)
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
		entry.Timestamp,
		entry.MessageID,
		entry.UserIDHash,
		entry.PushName,
		entry.Bidang,
		entry.Pelayanan,
		entry.JenisInfo,
		entry.PesanAsli,
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

	client := &http.Client{Timeout: l.timeout}

	req, err := http.NewRequest("POST", l.scriptURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("[LOGGER] Gagal membuat request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
