package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/scheduler"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command-line arguments
	checkNow := flag.Bool("check-now", false, "Run an ad-hoc Dropbox check")
	listFolders := flag.Bool("list-folders", false, "List folders in Dropbox")
	lastChanged := flag.Bool("last-changed", false, "Check the date of the last change in every folder")
	flag.Parse()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	fmt.Println("\nStarting Dropbox Monitor...")

	// 🔹 Step 1: Test Dropbox Authentication Immediately
	err = dropbox.TestConnection()
	if err != nil {
		log.Printf("❌ Dropbox authentication failed: %v", err)
		// Inspect the access token for more details
		dropbox.InspectAccessToken()
		return
	}

	log.Println("✅ Successfully connected to Dropbox!")

	// List folders
	if *listFolders {
		dropbox.ListFolders()
		return
	}

	// Last changed dates
	if *lastChanged {
		dropbox.ListLastChangedDates()
		return
	}

	// Ad-hoc Dropbox check
	if *checkNow {
		log.Println("Running ad-hoc Dropbox check...")
		_, err := dropbox.CheckForChanges()
		if err != nil {
			log.Printf("Error checking Dropbox: %v", err)
		} else {
			log.Println("Ad-hoc Dropbox check completed successfully!")
		}
		return
	}

	// 🔹 Step 2: Start the scheduler
	err = scheduler.Start()
	if err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	select {} // Keeps the program running
}
