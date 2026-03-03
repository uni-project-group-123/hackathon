package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	BackendURL  string `json:"backend_url"`
	ZoneID      int    `json:"zone_id"`
	AccessPoint string `json:"access_point"`
}

type AuthRequest struct {
	CardID    string `json:"card_id"`
	ZoneID    int    `json:"zone_id"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Success   bool   `json:"success"`
}

type AuthResponse struct {
	Success  bool   `json:"success"`
	UserID   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
	Message  string `json:"message,omitempty"`
	Action   string `json:"action,omitempty"`
}

var config Config

func loadConfig() {
	config = Config{
		BackendURL:  "http://localhost:5000",
		ZoneID:      1,
		AccessPoint: "main-door",
	}

	// Try to load from config file
	file, err := os.Open("config.json")
	if err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(&config)
	}

	// Override with environment variables
	if url := os.Getenv("BACKEND_URL"); url != "" {
		config.BackendURL = url
	}
	if zoneID := os.Getenv("ZONE_ID"); zoneID != "" {
		fmt.Sscanf(zoneID, "%d", &config.ZoneID)
	}
	log.Printf("Config: Backend=%s, ZoneID=%d, AccessPoint=%s",
		config.BackendURL, config.ZoneID, config.AccessPoint)
}

func authenticate(cardID string) AuthResponse {
	authReq := AuthRequest{
		CardID:    cardID,
		ZoneID:    config.ZoneID,
		Method:    "card",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(authReq)

	req, _ := http.NewRequest("POST", config.BackendURL+"/api/authenticate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Auth request failed: %v", err)
		return AuthResponse{Success: false, Message: "Connection error"}
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)
	return authResp
}

func logAccess(authReq AuthRequest, success bool) {
	jsonData, _ := json.Marshal(authReq)

	req, _ := http.NewRequest("POST", config.BackendURL+"/api/access-log", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	client.Do(req)
}

func controlDoor(grant bool) {
	// This would interface with GPIO in production
	// For now, just log
	if grant {
		log.Println("🔓 DOOR OPENED")
		// GPIO 17 - set HIGH to open door
		// gpioWrite(17, 1)
		// time.Sleep(3 * time.Second)
		// gpioWrite(17, 0)
	} else {
		log.Println("🔒 DOOR DENIED")
		// Play denied sound
	}
}

func simulateNFC() {
	// Simulate NFC card reads for testing
	// In production, use nfcpy or libnfc
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n=== NFC Reader Simulator ===")
	fmt.Println("Enter card ID to test (or 'q' to quit):")
	fmt.Println("Test cards: STU001, STU002, STU003, FAC001")

	for {
		fmt.Print("> ")
		cardID, _ := reader.ReadString('\n')
		cardID = strings.TrimSpace(cardID)

		if cardID == "q" || cardID == "quit" {
			break
		}

		if cardID == "" {
			continue
		}

		log.Printf("Card detected: %s", cardID)

		authReq := AuthRequest{
			CardID:    cardID,
			ZoneID:    config.ZoneID,
			Method:    "card",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		resp := authenticate(cardID)

		if resp.Success {
			controlDoor(true)
			authReq.Action = "granted"
		} else {
			controlDoor(false)
			authReq.Action = "denied"
		}

		logAccess(authReq, resp.Success)
		fmt.Printf("Result: %s - %s\n", resp.Message, resp.UserName)
	}
}

func readFromFile(filename string) {
	// Watch a file for card reads (for USB NFC readers that write to file)
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Could not open %s: %v", filename, err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		cardID := strings.TrimSpace(line)
		if cardID != "" {
			log.Printf("Card from file: %s", cardID)
			resp := authenticate(cardID)
			controlDoor(resp.Success)
		}
	}
}

func main() {
	loadConfig()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║   Sentinel Access Controller v1.0     ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Printf("Zone: %d | Backend: %s\n", config.ZoneID, config.BackendURL)
	fmt.Println()

	// Check backend connectivity
	resp, err := http.Get(config.BackendURL + "/api/health")
	if err != nil {
		log.Printf("Warning: Cannot connect to backend: %v", err)
	} else {
		resp.Body.Close()
		log.Println("Connected to backend")
	}

	// Run in simulation mode
	simulateNFC()
}
