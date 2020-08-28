package door

import (
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/rpi"
)

// RPiDoor does nothing.
type RPiDoor struct {
	authOkCh    chan struct{}
	authOkPin   gpio.PinOut
	authFailCh  chan struct{}
	authFailPin gpio.PinOut
}

// NewRPiDoor returns a new RPiDoor instance.
func NewRPiDoor() (*RPiDoor, error) {
	// TODO(duckworthd): Make pins configurable.
	result := &RPiDoor{
		authOkCh:    make(chan struct{}),
		authOkPin:   rpi.P1_15,
		authFailCh:  make(chan struct{}),
		authFailPin: rpi.P1_16,
	}
	go result.DoorLoop()
	return result, nil
}

// AuthOK returns all-zero data blocks.
func (r *RPiDoor) AuthOK() error {
	message := struct{}{}
	select {
	case r.authOkCh <- message:
		log.Println("Enqueued AuthOK message.")
	}
	return nil
}

// AuthFail returns all-zero data blocks.
func (r *RPiDoor) AuthFail() error {
	message := struct{}{}
	select {
	case r.authFailCh <- message:
		log.Println("Enqueued AuthFail message.")
	}
	return nil
}

// String returns a string representation of a RPiDoor.
func (r *RPiDoor) String() string {
	return "RPiDoor"
}

// DoorLoop is an infinite loop monitoring door access.
func (r *RPiDoor) DoorLoop() {
	// TODO(duckworthd): Make timeout configurable.
	timeout := 3 * time.Second
	for {
		select {
		case <-r.authOkCh:
			log.Println("AuthOK message received. Turning on AuthOk pin.")
			r.authOkPin.Out(gpio.High)

			time.Sleep(timeout)

			log.Println("Turning off AuthOk pin.")
			r.authOkPin.Out(gpio.Low)
		case <-r.authFailCh:
			log.Println("AuthFail message received. Turning on AuthFail pin.")
			r.authFailPin.Out(gpio.High)

			time.Sleep(timeout)

			log.Println("Turning off AuthFail pin.")
			r.authFailPin.Out(gpio.Low)
		}
	}
}
