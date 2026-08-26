package handler

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"chatbot-bkpsdm/config"
	"chatbot-bkpsdm/logger"
	"chatbot-bkpsdm/session"
)

// Handler mengelola logika navigasi menu chatbot berbasis pohon (tree) rekursif.
type Handler struct {
	root     *config.Root
	sessions *session.Manager
	log      *logger.Logger
}

// NewHandler membuat Handler baru.
func NewHandler(root *config.Root, sessions *session.Manager, log *logger.Logger) *Handler {
	return &Handler{
		root:     root,
		sessions: sessions,
		log:      log,
	}
}

// HandleMessage memproses pesan masuk dan mengembalikan teks balasan.
func (h *Handler) HandleMessage(userID, pushName, messageID, messageText string) string {
	input := strings.TrimSpace(messageText)
	inputLower := strings.ToLower(input)

	// Perintah "menu" selalu kembali ke menu utama.
	if inputLower == "menu" {
		h.sessions.Reset(userID)
		h.record(userID, pushName, messageID, input, nil, "Menu utama", "terjawab")
		return h.render(h.root.Root, true)
	}

	state := h.sessions.Get(userID)

	// Node saat ini berdasarkan path tersimpan.
	current := h.nodeAt(state.Path)
	if current == nil {
		// Path tidak valid (mis. konfigurasi berubah). Reset ke menu utama.
		h.sessions.Reset(userID)
		return h.render(h.root.Root, true)
	}

	// Pilihan "0" = kembali ke node induk.
	if input == "0" {
		if len(state.Path) == 0 {
			// Sudah di menu utama, tetap tampilkan menu utama.
			h.record(userID, pushName, messageID, input, state.Path, "Menu utama", "terjawab")
			return h.render(h.root.Root, true)
		}
		parentPath := state.Path[:len(state.Path)-1]
		h.sessions.SetPath(userID, parentPath)
		parent := h.nodeAt(parentPath)
		h.record(userID, pushName, messageID, input, parentPath, "Kembali", "terjawab")
		return h.render(parent, len(parentPath) == 0)
	}

	// Penanganan berdasarkan tipe node saat ini.
	switch current.Type {
	case config.TypeMenu:
		return h.handleMenu(userID, pushName, messageID, input, state.Path, current)
	case config.TypePelayanan:
		return h.handlePelayanan(userID, pushName, messageID, input, state.Path, current)
	default:
		// FAQ tidak seharusnya menjadi node aktif (langsung ditampilkan). Reset.
		h.sessions.Reset(userID)
		return h.render(h.root.Root, true)
	}
}

// handleMenu menangani pemilihan anak pada node bertipe "menu".
func (h *Handler) handleMenu(userID, pushName, messageID, input string, path []string, node *config.Node) string {
	child, ok := node.Children[input]
	if !ok {
		h.record(userID, pushName, messageID, input, path, "", "tidak_dikenali")
		return h.root.InvalidChoice + "\n\n" + h.render(node, len(path) == 0)
	}

	newPath := append(append([]string{}, path...), input)

	switch child.Type {
	case config.TypeMenu:
		// Masuk ke submenu.
		h.sessions.SetPath(userID, newPath)
		h.record(userID, pushName, messageID, input, newPath, child.Judul, "terjawab")
		return h.render(child, false)

	case config.TypePelayanan:
		// Masuk ke pelayanan, tampilkan submenu info.
		h.sessions.SetPath(userID, newPath)
		h.record(userID, pushName, messageID, input, newPath, child.Judul, "terjawab")
		return h.renderPelayanan(child)

	case config.TypeFAQ:
		// FAQ: tampilkan jawaban langsung, TIDAK mengubah posisi (tetap di menu ini).
		h.record(userID, pushName, messageID, input, newPath, child.Judul, "terjawab")
		footer := h.navFooter(path)
		return child.Jawaban + footer
	}

	// Tipe tidak dikenal.
	h.record(userID, pushName, messageID, input, path, "", "tidak_dikenali")
	return h.root.InvalidChoice + "\n\n" + h.render(node, len(path) == 0)
}

// handlePelayanan menangani pemilihan info (syarat/cara/petugas) pada node "pelayanan".
func (h *Handler) handlePelayanan(userID, pushName, messageID, input string, path []string, node *config.Node) string {
	info := node.Info
	if info == nil {
		h.sessions.Reset(userID)
		return h.render(h.root.Root, true)
	}

	// Bangun daftar pilihan yang tersedia (hanya field yang terisi).
	options := h.pelayananOptions(info)

	selected, ok := options[input]
	if !ok {
		h.record(userID, pushName, messageID, input, path, node.Judul, "tidak_dikenali")
		return h.root.InvalidChoice + "\n\n" + h.renderPelayanan(node)
	}

	// Status pencatatan: jika memilih info petugas, tandai sebagai dialihkan ke petugas.
	status := "terjawab"
	if selected.key == "petugas" {
		status = "dialihkan_ke_petugas"
	}

	h.record(userID, pushName, messageID, input, path, node.Judul+" - "+selected.label, status)

	footer := "\n\n---\n📌 Ketik *0* untuk kembali ke pilihan pelayanan\n📌 Ketik *menu* untuk ke menu utama"
	return selected.text + footer
}

// pelOption merepresentasikan satu pilihan info pada pelayanan.
type pelOption struct {
	key   string
	label string
	text  string
}

// pelayananOptions memetakan nomor pilihan ke info pelayanan yang tersedia.
func (h *Handler) pelayananOptions(info *config.PelayananInfo) map[string]pelOption {
	options := make(map[string]pelOption)
	n := 1
	if strings.TrimSpace(info.Syarat) != "" {
		options[strconv.Itoa(n)] = pelOption{key: "syarat", label: "Persyaratan", text: info.Syarat}
		n++
	}
	if strings.TrimSpace(info.CaraPengajuan) != "" {
		options[strconv.Itoa(n)] = pelOption{key: "cara_pengajuan", label: "Cara Pengajuan", text: info.CaraPengajuan}
		n++
	}
	if strings.TrimSpace(info.Petugas) != "" {
		options[strconv.Itoa(n)] = pelOption{key: "petugas", label: "Hubungi Petugas", text: info.Petugas}
		n++
	}
	return options
}

// nodeAt menelusuri pohon mengikuti path dan mengembalikan node tujuan.
// Path kosong mengembalikan node root (menu utama).
func (h *Handler) nodeAt(path []string) *config.Node {
	node := h.root.Root
	for _, key := range path {
		if node == nil || node.Type != config.TypeMenu {
			return nil
		}
		child, ok := node.Children[key]
		if !ok {
			return nil
		}
		node = child
	}
	return node
}

// render membentuk teks daftar menu untuk node bertipe "menu".
// isRoot menentukan apakah memakai teks welcome khusus.
func (h *Handler) render(node *config.Node, isRoot bool) string {
	if node == nil {
		return h.root.Welcome
	}
	if isRoot {
		return h.root.Welcome
	}

	var b strings.Builder
	if node.Deskripsi != "" {
		b.WriteString(node.Deskripsi)
	} else {
		b.WriteString("📋 *" + node.Judul + "*\n\nSilakan pilih:")
	}
	b.WriteString("\n\n")

	for _, key := range h.orderedKeys(node) {
		child := node.Children[key]
		b.WriteString(key + "️⃣ " + child.Judul + "\n")
	}

	b.WriteString("\n0️⃣ Kembali\n")
	b.WriteString("\n📌 Ketik *menu* untuk kembali ke menu utama.")
	return b.String()
}

// renderPelayanan membentuk submenu info untuk sebuah pelayanan.
func (h *Handler) renderPelayanan(node *config.Node) string {
	var b strings.Builder
	b.WriteString("📋 *" + node.Judul + "*\n\nPilih informasi yang dibutuhkan:\n\n")

	options := h.pelayananOptions(node.Info)
	// Tampilkan dalam urutan nomor.
	for i := 1; i <= len(options); i++ {
		key := strconv.Itoa(i)
		if opt, ok := options[key]; ok {
			b.WriteString(key + "️⃣ " + opt.label + "\n")
		}
	}

	b.WriteString("\n0️⃣ Kembali\n")
	b.WriteString("\n📌 Ketik *menu* untuk kembali ke menu utama.")
	return b.String()
}

// navFooter memberi petunjuk navigasi setelah menampilkan jawaban FAQ.
func (h *Handler) navFooter(path []string) string {
	if len(path) == 0 {
		return "\n\n---\n📌 Ketik nomor lain untuk pertanyaan lain, atau *menu* untuk menu utama."
	}
	return "\n\n---\n📌 Ketik nomor lain untuk pertanyaan lain\n📌 Ketik *0* untuk kembali ke menu sebelumnya\n📌 Ketik *menu* untuk ke menu utama."
}

// orderedKeys mengembalikan key anak node terurut.
// Jika node.Order diisi, gunakan itu; jika tidak, urutkan numerik.
func (h *Handler) orderedKeys(node *config.Node) []string {
	if len(node.Order) > 0 {
		return node.Order
	}
	keys := make([]string, 0, len(node.Children))
	for k := range node.Children {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, ei := strconv.Atoi(keys[i])
		nj, ej := strconv.Atoi(keys[j])
		if ei == nil && ej == nil {
			return ni < nj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// record mencatat interaksi ke logger. Bidang & Pelayanan diturunkan dari path.
func (h *Handler) record(userID, pushName, messageID, pesanAsli string, path []string, jenisInfo, status string) {
	if h.log == nil {
		return
	}

	bidang, pelayanan := h.deriveCategories(path)

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

// deriveCategories menentukan nama Bidang dan Pelayanan dari path untuk pencatatan.
// Bidang = judul node pada level 1. Pelayanan = judul node pelayanan/faq terdalam.
func (h *Handler) deriveCategories(path []string) (bidang, pelayanan string) {
	if len(path) == 0 {
		return "", ""
	}

	node := h.root.Root
	for i, key := range path {
		if node == nil || node.Type != config.TypeMenu {
			break
		}
		child, ok := node.Children[key]
		if !ok {
			break
		}
		if i == 0 {
			bidang = child.Judul
		}
		// Simpan judul pelayanan/faq terdalam sebagai "pelayanan".
		if child.Type == config.TypePelayanan || child.Type == config.TypeFAQ {
			pelayanan = child.Judul
		}
		node = child
	}
	return bidang, pelayanan
}
