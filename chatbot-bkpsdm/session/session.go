package session

import (
	"sync"
	"time"
)

// State merepresentasikan posisi menu pengguna saat ini
type State struct {
	// Level menu: "main", "bidang", "pelayanan", "info"
	Level string

	// Key bidang yang dipilih: "kesejahteraan", "pengadaan", "pengembangan"
	Bidang string

	// Key pelayanan yang dipilih: "cuti", "gelar", "pangkat", "mutasi", "tubel", "fungsional"
	Pelayanan string

	// Waktu terakhir interaksi
	LastActive time.Time
}

// Manager mengelola sesi pengguna berdasarkan nomor WhatsApp
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*State
}

// NewManager membuat session manager baru
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*State),
	}
}

// Get mengambil state pengguna, buat baru jika belum ada
func (m *Manager) Get(userID string) *State {
	m.mu.RLock()
	state, exists := m.sessions[userID]
	m.mu.RUnlock()

	if !exists {
		state = &State{
			Level:      "main",
			LastActive: time.Now(),
		}
		m.mu.Lock()
		m.sessions[userID] = state
		m.mu.Unlock()
	}

	return state
}

// Set menyimpan state pengguna
func (m *Manager) Set(userID string, state *State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state.LastActive = time.Now()
	m.sessions[userID] = state
}

// Reset mengembalikan state pengguna ke menu utama
func (m *Manager) Reset(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[userID] = &State{
		Level:      "main",
		LastActive: time.Now(),
	}
}

// CleanupExpired membersihkan sesi yang sudah tidak aktif (opsional)
func (m *Manager) CleanupExpired(maxInactive time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, state := range m.sessions {
		if now.Sub(state.LastActive) > maxInactive {
			delete(m.sessions, id)
		}
	}
}
