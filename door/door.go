package door

// Door is an interface for interacting with a door.
type Door interface {
	// Authentication passed. Open door for a short period of time.
	AuthOK() error

	// Authentication failed. Lock door.
	AuthFail() error

	String() string
}
