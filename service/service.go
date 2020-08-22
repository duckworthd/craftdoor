package service

import (
	"context"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pakohan/craftdoor/lib"
	"github.com/pakohan/craftdoor/model"
	"github.com/pakohan/craftdoor/rfid"
)

// Service contains the business logic
type Service struct {
	m  model.Model
	r  rfid.Reader
	cl *lib.ChangeListener
}

// New returns a new service instance
func New(m model.Model, r rfid.Reader, cl *lib.ChangeListener) *Service {
	s := &Service{
		m:  m,
		r:  r,
		cl: cl,
	}
	go s.DoorAccessLoop()
	return s
}

// // WaitForChange returns as soon as the state id is different to the one passed
// func (s *Service) WaitForChange(ctx context.Context, id uuid.UUID) (lib.State, error) {
// 	return s.cl.WaitForChange(ctx, id)
// }

// // RegisterKey inserts the key id of the next being presented to the reader into the table
// func (s *Service) RegisterKey(ctx context.Context) (key.Key, error) {
// 	log.Printf("registering key")
// 	state, err := s.cl.ReturnFirstKey(ctx)
// 	if err != nil {
// 		return key.Key{}, err
// 	}

// 	log.Printf("got key %s", state.TagInfo[0])

// 	k := key.Key{
// 		Secret:    state.TagInfo[0],
// 		AccessKey: uuid.New().String(),
// 	}
// 	err = s.m.KeyModel.Create(ctx, &k)
// 	if err != nil {
// 		return key.Key{}, err
// 	}

// 	log.Printf("inserted as key id %d", k.ID)

// 	return k, nil
// }

// func (s *Service) door() {
// 	uid := uuid.Nil
// 	for {
// 		state, err := s.WaitForChange(context.Background(), uid)
// 		if err != nil {
// 			log.Printf("got err waiting for change: %s", err.Error())
// 			continue
// 		} else if !state.IsTagAvailable {
// 			continue
// 		}
// 		uid = state.ID

// 		res, err := s.m.KeyModel.IsAccessAllowed(context.Background(), state.TagInfo[0], 1)
// 		if err != nil {
// 			log.Printf("got err checking whether key is allowed to access: %s", err.Error())
// 		} else {
// 			log.Printf("key %s is allowed to access door %d -> %t", state.TagInfo[0], 1, res)
// 		}
// 	}
// }

// ReadNextTag reads the next available RFID tag before the timeout.
//
// If timeout is reached, a state with an empty TagInfo field is returned.
func (s *Service) ReadNextTag(timeout time.Duration) (*lib.State, error) {
	result := &lib.State{
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
	for {
		// TODO(duckworthd): There is contention for ownership of the tag reader. Find a better way...
		state, err := s.ReadNextTag(5 * time.Second)
		if err != nil {
			log.Printf("Error encountered in OpenDoorLoop: %s", err)
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
		} else {
			log.Printf("Access NOT granted for key=%s.", state.TagInfo.ID)
		}
		time.Sleep(3 * time.Second)
	}
}
