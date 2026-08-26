package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Tipe node menu:
//   "menu"      -> node yang punya anak (bidang, sub-kategori). Menampilkan daftar pilihan.
//   "pelayanan" -> layanan berkas. Menampilkan submenu info (syarat, cara pengajuan, petugas).
//   "faq"       -> tanya-jawab. Langsung menampilkan jawaban tanpa submenu.
const (
	TypeMenu      = "menu"
	TypePelayanan = "pelayanan"
	TypeFAQ       = "faq"
)

// Node adalah unit menu rekursif. Bisa berupa menu (punya Children),
// pelayanan (punya Info), atau FAQ (punya Jawaban).
type Node struct {
	// Tipe node: "menu", "pelayanan", atau "faq"
	Type string `json:"type"`

	// Judul yang ditampilkan di daftar menu induk dan dipakai untuk pencatatan
	Judul string `json:"judul"`

	// Deskripsi opsional yang ditampilkan di atas daftar pilihan (khusus type "menu")
	Deskripsi string `json:"deskripsi,omitempty"`

	// Children untuk type "menu": dipetakan berdasarkan nomor pilihan ("1", "2", ...)
	Children map[string]*Node `json:"children,omitempty"`

	// Order menentukan urutan tampil anak berdasarkan key nomor ("1","2","3"...)
	// Jika kosong, urutan diambil dari sorting numerik key Children.
	Order []string `json:"order,omitempty"`

	// Info untuk type "pelayanan": bagian syarat, cara pengajuan, petugas, dan tautan
	Info *PelayananInfo `json:"info,omitempty"`

	// Jawaban untuk type "faq": teks jawaban langsung
	Jawaban string `json:"jawaban,omitempty"`
}

// PelayananInfo berisi detail sebuah pelayanan berkas.
// Field kosong ("") tidak akan ditampilkan sebagai pilihan submenu.
type PelayananInfo struct {
	Syarat        string `json:"syarat,omitempty"`
	CaraPengajuan string `json:"cara_pengajuan,omitempty"`
	Petugas       string `json:"petugas,omitempty"`
}

// Root adalah konfigurasi keseluruhan chatbot.
type Root struct {
	Welcome       string `json:"welcome"`
	InvalidChoice string `json:"invalid_choice"`
	// Menu utama adalah node bertipe "menu"
	Root *Node `json:"root"`
}

// Load membaca file responses.json dari beberapa kemungkinan lokasi.
func Load() (*Root, error) {
	paths := []string{
		"config/responses.json",
		"./config/responses.json",
	}

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		dir := filepath.Dir(filename)
		paths = append(paths, filepath.Join(dir, "responses.json"))
	}

	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("gagal membaca responses.json: %w", err)
	}

	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("gagal parsing responses.json: %w", err)
	}
	if root.Root == nil {
		return nil, fmt.Errorf("responses.json tidak memiliki node 'root'")
	}

	return &root, nil
}
