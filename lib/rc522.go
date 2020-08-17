package lib

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pakohan/craftdoor/config"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host"
)

type reader struct {
	subscriber     Subscriber
	deviceFile     string
	rstPin, irqPin gpio.PinIO
	lock           *sync.Mutex
	rlock          *sync.Mutex
	rfid           *mfrc522.Dev
	portCloser     spi.PortCloser
}

// NewRC522Reader returns a new Reader implementation accessing the RC522 reader
// via SPI on a RPI GPIO board.
//
//  cfg         Configuration object
//  s           Callback used when new data is read from the RC522.
func NewRC522Reader(cfg config.Config, s Subscriber) (Reader, error) {
	_, err := host.Init()
	if err != nil {
		return nil, err
	}

	rstPinReg := gpioreg.ByName(cfg.RSTPin)
	if rstPinReg == nil {
		return nil, fmt.Errorf("Reset pin %s can not be found", cfg.RSTPin)
	}

	irqPinReg := gpioreg.ByName(cfg.IRQPin)
	if irqPinReg == nil {
		return nil, fmt.Errorf("IRQ pin %s can not be found", cfg.IRQPin)
	}

	r := &reader{
		subscriber: s,
		deviceFile: cfg.Device,
		rstPin:     rstPinReg,
		irqPin:     irqPinReg,
		lock:       &sync.Mutex{},
		rlock:      &sync.Mutex{},
	}

	err = r.initreader()
	if err != nil {
		return nil, err
	}
	go r.runloop()
	log.Printf("initialized reader")
	return r, nil
}

// initReader initializes the RC522 device.
func (r *reader) initreader() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.rfid != nil {
		err := r.rfid.Halt()
		if err != nil {
			return err
		}
	}

	if r.portCloser != nil {
		err := r.portCloser.Close()
		if err != nil {
			return err
		}
	}

	var err error
	r.portCloser, err = spireg.Open(r.deviceFile)
	if err != nil {
		return err
	}
	log.Printf("Opened SPI Port: %s", r.portCloser.String())

	// Creates an MFRC522 device. All operations on the device are synchronous.
	r.rfid, err = mfrc522.NewSPI(r.portCloser, r.rstPin, r.irqPin, mfrc522.WithSync())
	if err != nil {
		return err
	}
	log.Printf("Successfully created mfrc522 SPI device: %s", r.rfid.String())

	return r.rfid.SetAntennaGain(5)
}

// runLoop is an infinite loop reading data from the device.
//
// Calls r.subscriber.Notify() whenever data changes.
func (r *reader) runloop() {
	var old [3]string
	for range time.Tick(1 * time.Second) {
		timeout := 10 * time.Second
		if old[0] != "" {
			timeout = 0
		}

		data, err := r.read(timeout)
		if err != nil {
			log.Printf("err reading data: %s", err)
			continue
		}

		if data != old {
			old = data
			r.subscriber.Notify(data)
		}
	}
}

// Reads all 16-byte blocks off of sector 0.
func (r *reader) read(timeout time.Duration) ([3]string, error) {
	r.rlock.Lock()
	defer r.rlock.Unlock()

	// Package-level defaults.
	var auth byte = commands.PICC_AUTHENT1B

	// Sectors range from 0...15. Each is 64 bytes.
	sector := 0

	// Each sector has 4 blocks, ranging from 0...3. Each is 16 bytes.
	// The last block ("Sector Trailer") contains the access keys for this sector.
	// The first block of the first sector contains the UID of the key.
	block := 0

	// Each sector is protected by two keys, "A" and "B". It is up to us to say
	// what the keys are and what privileges are granted by posessing them.
	key := mfrc522.Key{6, 5, 4, 3, 2, 1}

	b0, err := r.rfid.ReadCard(timeout, auth, sector, block, key)
	if err != nil {
		switch {
		case err.Error() == "mfrc522 lowlevel: IRQ error": // card needs to be reinitialized, see https://github.com/google/periph/issues/425
			return [3]string{}, r.initreader()
		case strings.HasPrefix(err.Error(), "mfrc522 lowlevel: timeout waiting for IRQ edge: "): // there's no card
			return [3]string{}, nil
		default: // any other error
			return [3]string{}, err
		}
	}

	b1, err := r.rfid.ReadCard(timeout, auth, sector, block+1, key)
	if err != nil {
		return [3]string{}, err
	}

	b2, err := r.rfid.ReadCard(timeout, auth, sector, block+2, key)
	if err != nil {
		return [3]string{}, err
	}

	return [3]string{
		string(b0),
		string(b1),
		string(b2),
	}, nil
}

// Writes keys, access bits, and data blocks for a singel sector
//
//  keyID       Identifier used to identify this card.
//  keySecret   Extra bits to write to card's data block
//  oldKey      Key that gives access t owrite over entire sector
//  keyA        New Key "A". Has authority to overwrite everything in the sector.
//  keyB        New Key "B". Has authority to overwrite data blocks.
func (r *reader) InitKey(keyID, keySecret [16]byte, oldKey, keyA, keyB mfrc522.Key) error {
	r.rlock.Lock()
	defer r.rlock.Unlock()

	defer log.Print("exit")

	// TODO(duckworthd): Synchronize which sector to use with reader.read().
	sector := 2

	timeout := 10 * time.Second
	//
	// Block |        Key A      |      Key B
	// ------+-------------------+---------------
	//   0   |        R/W        |      R/W
	//   1   |        R/W        |      R/W
	//   2   |        R          |      R/W
	//   3   | A:W, B:R/W, B:R/W |      ---
	//
	err := r.rfid.WriteSectorTrail(timeout, commands.PICC_AUTHENT1A, sector, keyA, keyB, &mfrc522.BlocksAccess{
		B0: mfrc522.AnyKeyRWID,
		B1: mfrc522.AnyKeyRWID,
		B2: mfrc522.RAB_WB_IN_DN,
		B3: mfrc522.KeyA_RN_WA_BITS_RA_WA_KeyB_RA_WA,
	}, oldKey)
	if err != nil {
		return err
	}
	log.Printf("successfully changed key of card % x to % x / % x", keyID, keyA, keyB)

	// Writes the following to the sector's three data blocks
	//
	//  Block | Contents
	// -------+----------
	//    0   | "craftwerk"
	//    1   | identifier for this card
	//    2   | Extra bits to write to card's data block
	for i, data := range [][16]byte{craftwerk, keyID, keySecret} {
		err = r.rfid.WriteCard(timeout, commands.PICC_AUTHENT1B, sector, i, data, keyB)
		if err != nil {
			return err
		}
	}

	return nil
}
