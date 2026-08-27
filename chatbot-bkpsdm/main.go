package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"chatbot-bkpsdm/config"
	"chatbot-bkpsdm/handler"
	"chatbot-bkpsdm/logger"
	"chatbot-bkpsdm/session"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"

	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Muat environment variables dari file .env (opsional, tidak error jika tidak ada)
	_ = godotenv.Load()

	// ===== MUAT KONFIGURASI RESPONS =====
	root, err := config.Load()
	if err != nil {
		fmt.Printf("Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[INFO] Konfigurasi respons berhasil dimuat")

	// ===== SETUP LOGGER =====
	scriptURL := os.Getenv("GOOGLE_SCRIPT_URL")
	scriptToken := os.Getenv("GOOGLE_SCRIPT_TOKEN")
	timeoutStr := os.Getenv("GOOGLE_SCRIPT_TIMEOUT")
	csvPath := os.Getenv("CSV_BACKUP_PATH")

	if csvPath == "" {
		csvPath = "./data/log_backup.csv"
	}

	timeout := 30 * time.Second
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(t) * time.Second
		}
	}

	log, err := logger.NewLogger(scriptURL, scriptToken, timeout, csvPath)
	if err != nil {
		fmt.Printf("Gagal membuat logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()
	fmt.Println("[INFO] Logger berhasil diinisialisasi")

	// ===== SETUP SESSION MANAGER =====
	sessionMgr := session.NewManager()

	// ===== SETUP HANDLER =====
	msgHandler := handler.NewHandler(root, sessionMgr, log)

	// ===== SETUP WHATSMEOW =====
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/whatsapp.db"
	}

	// Pastikan direktori data ada
	if err := os.MkdirAll("./data", 0755); err != nil {
		fmt.Printf("Gagal membuat direktori data: %v\n", err)
		os.Exit(1)
	}

	// Context utama aplikasi
	ctx := context.Background()

	// Inisialisasi database store
	dbLog := waLog.Stdout("Database", "WARN", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:"+dbPath+"?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000", dbLog)
	if err != nil {
		fmt.Printf("Gagal membuat database store: %v\n", err)
		os.Exit(1)
	}

	// Ambil device store (sesi pertama yang tersimpan, atau buat baru)
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		fmt.Printf("Gagal mengambil device store: %v\n", err)
		os.Exit(1)
	}

	// Buat client WhatsApp
	clientLog := waLog.Stdout("Client", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Register event handler
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			handleIncomingMessage(client, msgHandler, v)
		}
	})

	// Koneksi ke WhatsApp
	if client.Store.ID == nil {
		// Belum login, tampilkan QR code
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			fmt.Printf("Gagal koneksi: %v\n", err)
			os.Exit(1)
		}

		for evt := range qrChan {
			switch evt.Event {
			case "code":
				fmt.Println("\n========================================")
				fmt.Println("  SCAN QR CODE DI BAWAH INI")
				fmt.Println("  dengan WhatsApp > Linked Devices > Link a Device")
				fmt.Println("========================================\n")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			case "login":
				fmt.Println("\n[INFO] Berhasil login!")
			}
		}
	} else {
		// Sudah login, langsung koneksi
		err = client.Connect()
		if err != nil {
			fmt.Printf("Gagal koneksi: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[INFO] Terhubung ke WhatsApp")
	}

	fmt.Println("[INFO] Chatbot BKPSDM Bengkayang aktif dan siap menerima pesan!")
	fmt.Println("[INFO] Tekan Ctrl+C untuk menghentikan...")

	// Bersihkan sesi expired setiap 1 jam
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			sessionMgr.CleanupExpired(24 * time.Hour)
		}
	}()

	// Tunggu sinyal untuk berhenti
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n[INFO] Mematikan chatbot...")
	client.Disconnect()
	fmt.Println("[INFO] Chatbot berhasil dihentikan.")
}

// handleIncomingMessage memproses pesan masuk
func handleIncomingMessage(client *whatsmeow.Client, msgHandler *handler.Handler, msg *events.Message) {
	// Abaikan pesan dari bot sendiri
	if msg.Info.IsFromMe {
		return
	}

	// Abaikan pesan dari grup
	if msg.Info.IsGroup {
		return
	}

	// Ambil teks pesan
	messageText := ""
	if msg.Message.GetConversation() != "" {
		messageText = msg.Message.GetConversation()
	} else if msg.Message.GetExtendedTextMessage() != nil {
		messageText = msg.Message.GetExtendedTextMessage().GetText()
	}

	// Abaikan pesan kosong
	if messageText == "" {
		return
	}

	// Ambil informasi pengirim
	senderJID := msg.Info.Sender
	pushName := msg.Info.PushName
	messageID := msg.Info.ID

	// Proses pesan dan dapatkan respons
	userID := senderJID.String()
	response := msgHandler.HandleMessage(userID, pushName, messageID, messageText)

	// Kirim balasan
	sendReply(client, senderJID, response)
}

// sendReply mengirim pesan balasan ke pengguna
func sendReply(client *whatsmeow.Client, to types.JID, text string) {
	// Hapus device part dari JID (WhatsMeow terbaru pakai LID, harus tanpa :device)
	to = to.ToNonAD()

	msg := &waProto.Message{
		Conversation: proto.String(text),
	}

	_, err := client.SendMessage(context.Background(), to, msg)
	if err != nil {
		fmt.Printf("[ERROR] Gagal mengirim pesan ke %s: %v\n", to.String(), err)
	}
}
