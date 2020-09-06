package service

import (
	"context"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pakohan/craftdoor/door"
	"github.com/pakohan/craftdoor/lib"
	"github.com/pakohan/craftdoor/model"
	"github.com/pakohan/craftdoor/rfid"
)

// Service contains the business logic
type Service struct {
	m model.Model
	r rfid.Reader
	d door.Door
}

// New returns a new service instance
func New(m model.Model, r rfid.Reader, d door.Door) *Service {
	s := &Service{
		m: m,
		r: r,
		d: d,
	}

	// Start infinite loop that unlocks the door.
	go s.DoorAccessLoop()

	return s
}

// ReadNextTag reads the next available RFID tag before the timeout.
//
// If timeout is reached, a state with an empty TagInfo field is returned.
func (s *Service) ReadNextTag(timeout time.Duration) (*lib.State, error) {
	result := &lib.State{
		// TODO(duckworthd): Replace with a new UUID. Use UUID for state tracking.
		UUID: uuid.UUID{},
	}

	for {
		start := time.Now()
		uid, err := s.r.ReadUID(timeout)
		if err != nil {
			// Internal error worthy of a retry.
			if strings.Contains(err.Error(), "lowlevel: IRQ error") {
				log.Printf("IRQ error encountered. Re-initializing reader and trying again.")
				timeout -= time.Since(start)
				s.r.Initialize()
				continue
			}

			// Timeout.
			if strings.Contains(err.Error(), "lowlevel: timeout waiting for IRQ edge") {
				result.IsTagAvailable = false
				result.TagInfo = nil
				break
			}
		}

		if len(uid) == 0 {
			log.Printf("Encountered empty UID. Retrying.")
			timeout -= time.Since(start)
			continue
		}

		// Successful read.
		log.Printf("Successfully read tag: %s", hex.EncodeToString(uid))
		result.IsTagAvailable = true
		result.TagInfo = &lib.TagInfo{
			ID:   hex.EncodeToString(uid),
			Data: "",
		}
		break
	}
	return result, nil
}

// DoorAccessLoop is an infinite loop monitoring RFID tags put in front of the door.
//
// When a new RFID tag is put in front of the door and the tag is approved for entry, the door is unlocked.
func (s *Service) DoorAccessLoop() {
	log.Println("Starting DoorAccessLoop()...")
	timeout := 3 * time.Second
	for {
		// TODO(duckworthd): There is contention for ownership of the tag reader. Find a better way...
		state, err := s.ReadNextTag(timeout)
		if err != nil {
			log.Printf("Error encountered in DoorAccessLoop: %s", err)
			continue
		}

		if !state.IsTagAvailable {
			continue
		}

		// TODO(duckworthd): Add support for >1 doors.
		accessAllowed, err := s.m.KeyModel.IsAccessAllowed(context.Background(), state.TagInfo.ID)
		if err != nil {
			log.Printf("Error determining in IsAccessAllowed() for key=%s.", state.TagInfo.ID)
			continue
		}

		if accessAllowed {
			log.Printf("Access granted for key=%s.", state.TagInfo.ID)
			s.d.AuthOK()
		} else {
			log.Printf("Access NOT granted for key=%s.", state.TagInfo.ID)
			s.d.AuthFail()
		}

		// TODO(duckworthd): Figure out how to remove the need for this timeout.
		// AuthOK and AuthFail both enqueue messages and return instantaneously, which
		// will cause this loop to cycle very quickly when a tag is available.
		time.Sleep(timeout)
	}
}
