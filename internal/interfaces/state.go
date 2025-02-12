package interfaces

// StateManager defines the interface for state management
type StateManager interface {
	GetString(key string) string
	SetString(key, value string) error
}
