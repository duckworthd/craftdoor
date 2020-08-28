package door

import (
	"log"
	"time"
)

// DummyDoor does nothing.
type DummyDoor struct {
	authOkCh   chan struct{}
	authFailCh chan struct{}
}

// NewDummyDoor returns a new DummyDoor instance.
func NewDummyDoor() (*DummyDoor, error) {
	result := &DummyDoor{
		authOkCh:   make(chan struct{}),
		authFailCh: make(chan struct{}),
	}
	go result.DoorLoop()
	return result, nil
}

// AuthOK returns all-zero data blocks.
func (r *DummyDoor) AuthOK() error {
	message := struct{}{}
	select {
	case r.authOkCh <- message:
		log.Println("Enqueued AuthOK message.")
	}
	return nil
}

// AuthFail returns all-zero data blocks.
func (r *DummyDoor) AuthFail() error {
	message := struct{}{}
	select {
	case r.authFailCh <- message:
		log.Println("Enqueued AuthFail message.")
	}
	return nil
}

// String returns a string representation of a DummyDoor.
func (r *DummyDoor) String() string {
	return "DummyDoor"
}

// DoorLoop is an infinite loop monitoring door access.
func (r *DummyDoor) DoorLoop() {
	timeout := 3 * time.Second
	for {
		select {
		case <-r.authOkCh:
			log.Println("Opening AuthOK.")
			time.Sleep(timeout)
			log.Println("Closing AuthOK.")
		case <-r.authFailCh:
			log.Println("Opening AuthFail")
			time.Sleep(timeout)
			log.Println("Closing AuthFail.")
		}
	}
}
