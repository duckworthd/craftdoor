package service

import (
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
	// go s.door()
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

// ReadNextTag reads the next available RFID tag before the next timeout.
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
				log.Printf("Timed out waiting for next RFID tag.")
				result.IsTagAvailable = false
				result.TagInfo = nil
				break
			}
		}

		// Successful read.
		result.IsTagAvailable = true
		result.TagInfo = &lib.TagInfo{
			ID:   hex.EncodeToString(uid),
			Data: "",
		}
		break
	}
	return result, nil
}
