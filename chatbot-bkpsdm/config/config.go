package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ServiceInfo berisi informasi detail pelayanan
type ServiceInfo struct {
	Persyaratan string `json:"persyaratan"`
	Prosedur    string `json:"prosedur"`
	Waktu       string `json:"waktu"`
	Formulir    string `json:"formulir"`
	Status      string `json:"status"`
	Kendala     string `json:"kendala"`
}

// Pelayanan berisi data pelayanan
type Pelayanan struct {
	Nama string      `json:"nama"`
	Menu string      `json:"menu"`
	Info ServiceInfo `json:"info"`
}

// Bidang berisi data bidang
type Bidang struct {
	Nama      string               `json:"nama"`
	Menu      string               `json:"menu"`
	Pelayanan map[string]Pelayanan `json:"pelayanan"`
}

// Responses berisi seluruh konfigurasi respons chatbot
type Responses struct {
	Welcome       string            `json:"welcome"`
	InvalidChoice string            `json:"invalid_choice"`
	Bidang        map[string]Bidang `json:"bidang"`
}

// LoadResponses membaca file responses.json
func LoadResponses() (*Responses, error) {
	// Cari file responses.json relatif terhadap executable atau working directory
	paths := []string{
		"config/responses.json",
		"./config/responses.json",
	}

	// Tambahkan path relatif terhadap file source (untuk development)
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

	var responses Responses
	if err := json.Unmarshal(data, &responses); err != nil {
		return nil, fmt.Errorf("gagal parsing responses.json: %w", err)
	}

	return &responses, nil
}
