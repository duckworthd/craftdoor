package main

import (
	"encoding/hex"
	"log"
	"time"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

// Reader wraps an MFRC522 Reader on SPI.
type Reader struct {
	port   spi.PortCloser
	device *mfrc522.Dev
}

// Initialize prepares a Reader for reading.
func (r *Reader) Initialize() error {
	r.Halt()

	log.Println("Initializing Reader.")
	var err error
	r.port, err = spireg.Open("")
	if err != nil {
		log.Println("Failed to open SPI port.")
		return err
	}

	r.device, err = mfrc522.NewSPI(r.port, rpi.P1_22, rpi.P1_18, mfrc522.WithSync())
	if err != nil {
		log.Println("Failed to start mfrc522.Dev.")
		return err
	}

	err = r.device.SetAntennaGain(5)
	if err != nil {
		log.Println("Failed to set antenna gain.")
		return err
	}

	log.Println("Successfully initialized Reader.")
	return nil
}

// Halt stops the Reader.
func (r *Reader) Halt() error {
	log.Println("Halting Reader.")
	var err error
	if r.device != nil {
		err := r.device.Halt()
		if err != nil {
			log.Println("Failed to Halt mfrc522.Dev.")
			return err
		}
	}

	if r.port != nil {
		err := r.port.Close()
		if err != nil {
			log.Println("Failed to close spi.PortCloser.")
			return err
		}
	}

	log.Println("Successfully halted Reader.")
	return err
}

// ReadUID reads the UID of the RFID tag.
func (r *Reader) ReadUID(timeout time.Duration) ([]byte, error) {
	return r.device.ReadUID(timeout)
}

// ReadSectorData reads all three data blocks from a given sector.
func (r *Reader) ReadSectorData(timeout time.Duration, sector int) ([]byte, error) {
	var auth byte = commands.PICC_AUTHENT1B
	key := mfrc522.DefaultKey
	// key := mfrc522.Key{6, 5, 4, 3, 2, 1}

	b0, err := r.device.ReadCard(timeout, auth, sector, 0, key)
	if err != nil {
		log.Printf("Failed to read sector=%d block=%d: %s", sector, 0, err)
		return nil, err
	}

	b1, err := r.device.ReadCard(timeout, auth, sector, 1, key)
	if err != nil {
		log.Printf("Failed to read sector=%d block=%d: %s", sector, 1, err)
		return nil, err
	}

	b2, err := r.device.ReadCard(timeout, auth, sector, 2, key)
	if err != nil {
		log.Printf("Failed to read sector=%d block=%d: %s", sector, 2, err)
		return nil, err
	}

	result := []byte{}
	result = append(result, b0...)
	result = append(result, b1...)
	result = append(result, b2...)
	return result, nil
}

func main() {
	// Make sure periph is initialized.
	log.Println("Initializing host.")
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Initialize RFID Reader.
	log.Println("Creating MFRC522 SPI device.")
	rfid := Reader{}
	err := rfid.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer rfid.Halt()

	timedOut := false
	cb := make(chan []byte)
	timer := time.NewTimer(30 * time.Second)

	// Stopping timer, flagging reader thread as timed out
	defer func() {
		timer.Stop()
		timedOut = true
		close(cb)
	}()

	go func() {
		log.Printf("Starting %s", rfid.device.String())

		for {
			timeout := 5 * time.Second
			// sector := 1
			// uid, err := rfid.ReadSectorData(timeout, sector)
			uid, err := rfid.ReadUID(timeout)
			if err != nil {
				log.Printf("Error in ReadSectorData: %s", err)
				if err.Error() == "mfrc522 lowlevel: IRQ error" {
					rfid.Initialize()
				}
				continue
			}

			if timedOut {
				log.Println("Timed out. Exiting thread.")
				return
			}

			log.Println("UID successfully read. Passing UID downstream.")
			cb <- uid
			return
		}
	}()

	for {
		select {
		case <-timer.C:
			log.Fatal("Didn't receive device data. Exiting.")
			return
		case data := <-cb:
			log.Printf("Received UID: %s. Exiting.", hex.EncodeToString(data))
			return
		}
	}
}
