package door

import "time"

// Door is an interface for interacting with a door.
type Door interface {
	// Authentication passed. Open door for a short period of time.
	AuthOK() error

	// Authentication failed. Lock door.
	AuthFail() error

	String() string
}

// Latch controls a single locking mechanism in a door.
type Latch interface {
	// Temporarily unlock a door. Resumes default state after duration.
	Unlock(duration time.Duration) error
}
