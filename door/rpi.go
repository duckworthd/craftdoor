package door

import (
	"errors"
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/rpi"
)

// RPiDoor is a door connected to a Raspberry Pi's GPIO pins.
type RPiDoor struct {
	authOkCh      chan struct{}
	authOkLatch   Latch
	authFailCh    chan struct{}
	authFailLatch Latch
	timeout       time.Duration
}

// NewRPiDoor returns a new RPiDoor instance.
func NewRPiDoor() (*RPiDoor, error) {
	// TODO(dukckworthd): Don't hardcode timezone.
	timezone, _ := time.LoadLocation("Europe/Berlin")

	// TODO(duckworthd): Don't hardcode times.
	openFrom := time.Date(0, 0, 0, 5, 0, 0, 0, timezone)  // 5AM
	openTill := time.Date(0, 0, 0, 23, 0, 0, 0, timezone) // 11PM

	// TODO(duckworthd): Make pins configurable.
	mainLatch, err := NewBasicLatch(rpi.P1_15)
	if err != nil {
		return nil, err
	}
	boltLatch, err := NewTimedEntryLatch(openFrom, openTill, rpi.P1_13)
	if err != nil {
		return nil, err
	}
	authOkLatch, err := NewMultiLatch([]Latch{mainLatch, boltLatch})
	if err != nil {
		return nil, err
	}

	authFailLatch, err := NewBasicLatch(rpi.P1_16)
	if err != nil {
		return nil, err
	}

	result := &RPiDoor{
		authOkCh:      make(chan struct{}),
		authOkLatch:   authOkLatch,
		authFailCh:    make(chan struct{}),
		authFailLatch: authFailLatch,
		// TODO(duckworthd): Make timeout configurable.
		timeout: 3 * time.Second,
	}
	go result.DoorLoop()
	return result, nil
}

// AuthOK enqueues a successful authentication message.
func (r *RPiDoor) AuthOK() error {
	message := struct{}{}
	select {
	case r.authOkCh <- message:
		log.Println("Enqueued AuthOK message.")
	default:
		log.Println("Failed to enqueue AuthOK message.")
	}
	return nil
}

// AuthFail enqueues a faild authentication message.
func (r *RPiDoor) AuthFail() error {
	message := struct{}{}
	select {
	case r.authFailCh <- message:
		log.Println("Enqueued AuthFail message.")
	default:
		log.Println("Failed to enqueue AuthOK message.")
	}
	return nil
}

// String returns a string representation of a RPiDoor.
func (r *RPiDoor) String() string {
	return "RPiDoor"
}

// DoorLoop is an infinite loop monitoring door access.
func (r *RPiDoor) DoorLoop() {
	for {
		select {
		case <-r.authOkCh:
			log.Println("AuthOK message received.")
			r.authOkLatch.Unlock(r.timeout)
		case <-r.authFailCh:
			log.Println("AuthFail message received.")
			r.authFailLatch.Unlock(r.timeout)
		}
	}
}

// BasicLatch is a latch that opens when asked.
type BasicLatch struct {
	pin      gpio.PinOut
	unlockCh chan time.Duration
}

// NewBasicLatch creates a new basic latch.
func NewBasicLatch(pin gpio.PinOut) (*BasicLatch, error) {
	result := &BasicLatch{
		pin:      pin,
		unlockCh: make(chan time.Duration),
	}
	go result.LatchLoop()
	return result, nil
}

// Unlock unlocks this latch for a limited time.
func (r *BasicLatch) Unlock(duration time.Duration) error {
	select {
	case r.unlockCh <- duration:
		log.Printf("Enqueuing Unlock duration: %s", duration)
	default:
		log.Println("Failed to enqueue Unlock duration.")
	}
	return nil
}

// LatchLoop is an infinite loop monitoring the latch.
func (r *BasicLatch) LatchLoop() {
	log.Println("Starting BasicLatch.LatchLoop().")
	r.pin.Out(gpio.High)

	for {
		duration := <-r.unlockCh
		log.Printf("Unlock event received. Holding latch open for: %s", duration)
		r.pin.Out(gpio.Low)
		time.Sleep(duration)
		r.pin.Out(gpio.High)
	}
}

// TimedEntryLatch is a latch that remains unlocked between two periods of time.
type TimedEntryLatch struct {
	// TODO(duckworthd): Use a custom TimeOfDay type that ignores date.
	openFrom time.Time
	openTill time.Time
	pin      gpio.PinOut
	unlockCh chan time.Duration
}

// NewTimedEntryLatch returns a new TimedEntryLatch.
func NewTimedEntryLatch(openFrom time.Time, openTill time.Time, pin gpio.PinOut) (*TimedEntryLatch, error) {
	// TODO(duckworthd): This test will do the wrong thing if openFrom and openTill are on
	// different dates. Fix it.
	if !openFrom.Before(openTill) {
		return nil, errors.New("openFrom < openTill required")
	}
	result := &TimedEntryLatch{
		openFrom: openFrom,
		openTill: openTill,
		pin:      pin,
		unlockCh: make(chan time.Duration),
	}
	go result.LatchLoop()
	return result, nil
}

// Unlock unlocks this latch for a limited time.
func (r *TimedEntryLatch) Unlock(duration time.Duration) error {
	select {
	case r.unlockCh <- duration:
		log.Printf("Enqueuing Unlock duration: %s", duration)
	default:
		log.Println("Failed to enqueue Unlock duration.")
	}
	return nil
}

// LatchLoop is an infinite loop monitoring the latch.
func (r *TimedEntryLatch) LatchLoop() {
	log.Println("Starting TimedEntryLatch.LatchLoop().")
	r.pin.Out(r.BaselineLevel(time.Now()))
	timer := time.NewTimer(1 * time.Nanosecond)

	for {
		select {
		case <-timer.C:
			log.Println("Timer fired. Setting latch to baseline level.")
			now := time.Now()
			level := r.BaselineLevel(now)
			log.Printf("Current time is now %s. Setting level to: %s.", now, level)
			r.pin.Out(level)
			next := r.NextTimedEntryEvent(now)
			log.Printf("Resetting timer. Next timer event: %s.", next)
			timer = time.NewTimer(next.Sub(now))
		case duration := <-r.unlockCh:
			log.Printf("Unlock event received. Holding latch open for: %s", duration)
			r.pin.Out(gpio.Low)
			time.Sleep(duration)
			r.pin.Out(r.BaselineLevel(time.Now()))
		}

	}
}

// BaselineLevel returns the default GPIO level right now.
func (r *TimedEntryLatch) BaselineLevel(t time.Time) gpio.Level {
	// TODO(duckworthd): Figure how to reduce complexity around timezones.
	t = t.In(r.openFrom.Location())
	s := r.openFrom
	openFrom := time.Date(t.Year(), t.Month(), t.Day(), s.Hour(), s.Minute(), s.Second(), s.Nanosecond(), t.Location())

	t = t.In(r.openTill.Location())
	e := r.openTill
	openTill := time.Date(t.Year(), t.Month(), t.Day(), e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), t.Location())

	if openFrom.Before(t) && t.Before(openTill) {
		return gpio.Low
	}
	return gpio.High
}

// NextTimedEntryEvent returns the time of the next timed entry event.
func (r *TimedEntryLatch) NextTimedEntryEvent(t time.Time) time.Time {
	// TODO(duckworthd): Figure how to reduce complexity around timezones.
	t = t.In(r.openFrom.Location())
	s := r.openFrom
	var openFrom time.Time = time.Date(t.Year(), t.Month(), t.Day(), s.Hour(), s.Minute(), s.Second(), s.Nanosecond(), t.Location())
	for openFrom.Before(t) {
		openFrom = openFrom.Add(24 * time.Hour)
	}

	t = t.In(r.openTill.Location())
	e := r.openTill
	openTill := time.Date(t.Year(), t.Month(), t.Day(), e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), t.Location())
	for openTill.Before(t) {
		openTill = openTill.Add(24 * time.Hour)
	}

	// Return the first of the two events.
	if openFrom.Before(openTill) {
		return openFrom
	}
	return openTill
}

// A MultiLatch represents multiple latches.
type MultiLatch struct {
	latches []Latch
}

// NewMultiLatch creates a new MultiLatch instance.
func NewMultiLatch(latches []Latch) (*MultiLatch, error) {
	result := &MultiLatch{
		latches: latches,
	}
	return result, nil
}

// Unlock unlocks all latches.
func (l *MultiLatch) Unlock(duration time.Duration) error {
	log.Printf("Enqueuing Unlock in all %d latches. Duration: %s", len(l.latches), duration)
	for _, latch := range l.latches {
		err := latch.Unlock(duration)
		if err != nil {
			return err
		}
	}
	return nil
}
