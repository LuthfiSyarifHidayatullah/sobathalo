package session

import (
	"sync"
	"time"
)

// State merepresentasikan posisi pengguna dalam pohon menu.
//
// Path adalah daftar nomor pilihan dari root ke node saat ini.
// Contoh:
//   []            -> di menu utama
//   ["1"]         -> di Bidang Kesejahteraan
//   ["1","8"]     -> di sub-kategori e-Kinerja
//   ["1","8","3"] -> di pelayanan Rekon Absensi (menampilkan submenu info)
type State struct {
	// Path menyimpan jejak navigasi (stack) berupa key nomor pilihan
	Path []string

	// Waktu terakhir interaksi
	LastActive time.Time
}

// Manager mengelola sesi pengguna berdasarkan nomor WhatsApp secara thread-safe.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*State
}

// NewManager membuat session manager baru.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*State),
	}
}

// Get mengambil salinan state pengguna. Jika belum ada, dibuat state baru di menu utama.
func (m *Manager) Get(userID string) *State {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.sessions[userID]
	if !exists {
		state = &State{
			Path:       []string{},
			LastActive: time.Now(),
		}
		m.sessions[userID] = state
	}

	// Kembalikan salinan agar caller tidak memodifikasi data internal tanpa lock
	pathCopy := make([]string, len(state.Path))
	copy(pathCopy, state.Path)
	return &State{Path: pathCopy, LastActive: state.LastActive}
}

// SetPath menyimpan path baru untuk pengguna.
func (m *Manager) SetPath(userID string, path []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pathCopy := make([]string, len(path))
	copy(pathCopy, path)
	m.sessions[userID] = &State{
		Path:       pathCopy,
		LastActive: time.Now(),
	}
}

// Reset mengembalikan state pengguna ke menu utama.
func (m *Manager) Reset(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[userID] = &State{
		Path:       []string{},
		LastActive: time.Now(),
	}
}

// CleanupExpired membersihkan sesi yang sudah tidak aktif melebihi maxInactive.
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
