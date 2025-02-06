package state

import (
	"encoding/json"
	"log"
	"os"
)

const stateFile = "config/state.json"

// State represents the last known Dropbox state
type State struct {
	Cursor string `json:"cursor"`
}

// Save writes the latest state to a file
func Save(cursor string) error {
	data, err := json.Marshal(State{Cursor: cursor})
	if err != nil {
		return err
	}

	err = os.WriteFile(stateFile, data, 0644)
	if err != nil {
		return err
	}

	log.Println("✅ State saved successfully.")
	return nil
}

// Load reads the last saved state from a file
func Load() (string, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No previous state exists
		}
		return "", err
	}

	var s State
	err = json.Unmarshal(data, &s)
	if err != nil {
		return "", err
	}

	return s.Cursor, nil
}

