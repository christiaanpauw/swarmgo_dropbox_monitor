package main

import (
	"fmt"
	"log"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/scheduler"
)

func main() {
	fmt.Println("Starting Dropbox Monitor...")

	// 🔹 Step 1: Test Dropbox Authentication Immediately
	err := dropbox.TestConnection()
	if err != nil {
		log.Fatalf("❌ Dropbox authentication failed: %v", err)
	}

	log.Println("✅ Successfully connected to Dropbox!")

	// 🔹 Step 2: Start the scheduler
	err = scheduler.Start()
	if err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	select {} // Keeps the program running
}

