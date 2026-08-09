package handler

import (
	"chatbot-bkpsdm/config"
	"chatbot-bkpsdm/logger"
	"chatbot-bkpsdm/session"
	"strings"
	"time"
)

// Handler mengelola logika menu chatbot
type Handler struct {
	responses *config.Responses
	sessions  *session.Manager
	log       *logger.Logger
}

// NewHandler membuat Handler baru
func NewHandler(responses *config.Responses, sessions *session.Manager, log *logger.Logger) *Handler {
	return &Handler{
		responses: responses,
		sessions:  sessions,
		log:       log,
	}
}

// bidangKeys memetakan nomor pilihan ke key bidang
var bidangKeys = map[string]string{
	"1": "kesejahteraan",
	"2": "pengadaan",
	"3": "pengembangan",
}

// pelayananKeys memetakan nomor pilihan ke key pelayanan per bidang
var pelayananKeys = map[string]map[string]string{
	"kesejahteraan": {
		"1": "cuti",
		"2": "gelar",
	},
	"pengadaan": {
		"1": "pangkat",
		"2": "mutasi",
	},
	"pengembangan": {
		"1": "tubel",
		"2": "fungsional",
	},
}

// infoKeys memetakan nomor pilihan ke key informasi
var infoKeys = map[string]string{
	"1": "persyaratan",
	"2": "prosedur",
	"3": "waktu",
	"4": "formulir",
	"5": "status",
	"6": "kendala",
}

// infoLabels label untuk pencatatan
var infoLabels = map[string]string{
	"1": "Persyaratan pelayanan",
	"2": "Prosedur pengajuan",
	"3": "Waktu penyelesaian",
	"4": "Formulir atau tautan pengajuan",
	"5": "Cek status pengajuan",
	"6": "Kendala dan hubungi petugas",
}

// HandleMessage memproses pesan masuk dan mengembalikan respons
func (h *Handler) HandleMessage(userID, pushName, messageID, messageText string) string {
	input := strings.TrimSpace(messageText)
	inputLower := strings.ToLower(input)

	state := h.sessions.Get(userID)

	// Perintah "menu" selalu kembali ke menu utama
	if inputLower == "menu" {
		h.sessions.Reset(userID)
		h.logMessage(userID, pushName, messageID, input, "", "", "Menu utama", "terjawab")
		return h.responses.Welcome
	}

	// Proses berdasarkan level state
	switch state.Level {
	case "main":
		return h.handleMainMenu(userID, pushName, messageID, input)
	case "bidang":
		return h.handleBidangMenu(userID, pushName, messageID, input, state)
	case "pelayanan":
		return h.handlePelayananMenu(userID, pushName, messageID, input, state)
	default:
		h.sessions.Reset(userID)
		return h.responses.Welcome
	}
}

// handleMainMenu menangani pilihan di menu utama
func (h *Handler) handleMainMenu(userID, pushName, messageID, input string) string {
	bidangKey, exists := bidangKeys[input]
	if !exists {
		h.logMessage(userID, pushName, messageID, input, "", "", "", "tidak_dikenali")
		return h.responses.InvalidChoice + "\n\n" + h.responses.Welcome
	}

	bidang, exists := h.responses.Bidang[bidangKey]
	if !exists {
		h.logMessage(userID, pushName, messageID, input, "", "", "", "tidak_dikenali")
		return h.responses.InvalidChoice + "\n\n" + h.responses.Welcome
	}

	// Update state ke level bidang
	h.sessions.Set(userID, &session.State{
		Level:  "bidang",
		Bidang: bidangKey,
	})

	h.logMessage(userID, pushName, messageID, input, bidang.Nama, "", "Pilih bidang", "terjawab")
	return bidang.Menu
}

// handleBidangMenu menangani pilihan di submenu bidang
func (h *Handler) handleBidangMenu(userID, pushName, messageID, input string, state *session.State) string {
	// Pilihan 0 = kembali ke menu utama
	if input == "0" {
		h.sessions.Reset(userID)
		h.logMessage(userID, pushName, messageID, input, "", "", "Kembali", "terjawab")
		return h.responses.Welcome
	}

	pelKeys, exists := pelayananKeys[state.Bidang]
	if !exists {
		h.sessions.Reset(userID)
		return h.responses.Welcome
	}

	pelKey, exists := pelKeys[input]
	if !exists {
		bidang := h.responses.Bidang[state.Bidang]
		h.logMessage(userID, pushName, messageID, input, bidang.Nama, "", "", "tidak_dikenali")
		return h.responses.InvalidChoice + "\n\n" + bidang.Menu
	}

	bidang := h.responses.Bidang[state.Bidang]
	pelayanan, exists := bidang.Pelayanan[pelKey]
	if !exists {
		h.logMessage(userID, pushName, messageID, input, bidang.Nama, "", "", "tidak_dikenali")
		return h.responses.InvalidChoice + "\n\n" + bidang.Menu
	}

	// Update state ke level pelayanan
	h.sessions.Set(userID, &session.State{
		Level:     "pelayanan",
		Bidang:    state.Bidang,
		Pelayanan: pelKey,
	})

	h.logMessage(userID, pushName, messageID, input, bidang.Nama, pelayanan.Nama, "Pilih pelayanan", "terjawab")
	return pelayanan.Menu
}

// handlePelayananMenu menangani pilihan di submenu pelayanan (info detail)
func (h *Handler) handlePelayananMenu(userID, pushName, messageID, input string, state *session.State) string {
	// Pilihan 0 = kembali ke menu bidang
	if input == "0" {
		bidang := h.responses.Bidang[state.Bidang]
		h.sessions.Set(userID, &session.State{
			Level:  "bidang",
			Bidang: state.Bidang,
		})
		h.logMessage(userID, pushName, messageID, input, bidang.Nama, "", "Kembali", "terjawab")
		return bidang.Menu
	}

	bidang := h.responses.Bidang[state.Bidang]
	pelayanan := bidang.Pelayanan[state.Pelayanan]

	infoKey, exists := infoKeys[input]
	if !exists {
		h.logMessage(userID, pushName, messageID, input, bidang.Nama, pelayanan.Nama, "", "tidak_dikenali")
		return h.responses.InvalidChoice + "\n\n" + pelayanan.Menu
	}

	// Ambil teks informasi berdasarkan key
	var infoText string
	switch infoKey {
	case "persyaratan":
		infoText = pelayanan.Info.Persyaratan
	case "prosedur":
		infoText = pelayanan.Info.Prosedur
	case "waktu":
		infoText = pelayanan.Info.Waktu
	case "formulir":
		infoText = pelayanan.Info.Formulir
	case "status":
		infoText = pelayanan.Info.Status
	case "kendala":
		infoText = pelayanan.Info.Kendala
	}

	// Tentukan status pencatatan
	logStatus := "terjawab"
	if infoKey == "kendala" {
		logStatus = "dialihkan_ke_petugas"
	}

	jenisInfo := infoLabels[input]
	h.logMessage(userID, pushName, messageID, input, bidang.Nama, pelayanan.Nama, jenisInfo, logStatus)

	// Tambahkan footer navigasi
	footer := "\n\n---\n📌 Ketik *0* untuk kembali ke menu pelayanan\n📌 Ketik *menu* untuk ke menu utama"
	return infoText + footer
}

// logMessage mencatat pesan ke logger
func (h *Handler) logMessage(userID, pushName, messageID, pesanAsli, bidang, pelayanan, jenisInfo, status string) {
	if h.log == nil {
		return
	}

	entry := logger.LogEntry{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		MessageID:  messageID,
		UserIDHash: logger.HashUserID(userID),
		PushName:   pushName,
		Bidang:     bidang,
		Pelayanan:  pelayanan,
		JenisInfo:  jenisInfo,
		PesanAsli:  pesanAsli,
		Status:     status,
	}

	h.log.Log(entry)
}
