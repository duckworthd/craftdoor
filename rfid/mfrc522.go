// An interface for using an MFRC522 RFID reader/writer.
//
// Uses a Raspberry Pi's hardware SPI interface to interact with RFID
// reader/writer. Assumes the following additional PIN configuration: RESET=22
// and IRQ=18.

package rfid

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host/rpi"
)

// Key used for all operations.
//
// TODO(duckworthd): Add support for user-defined keys.
var key = mfrc522.DefaultKey

// MFRC522Reader wraps an MFRC522 MFRC522Reader on SPI.
type MFRC522Reader struct {
	port   spi.PortCloser
	device *mfrc522.Dev
}

// AuthBlock wraps the trailer block containing the authentication keys and permissions bits for a sector.
type AuthBlock struct {
	KeyA        mfrc522.Key
	KeyB        mfrc522.Key
	Permissions mfrc522.BlocksAccess
}

// NewMFRC522Reader creates a new Reader object.
func NewMFRC522Reader() (*MFRC522Reader, error) {
	return &MFRC522Reader{}, nil
}

// Initialize prepares a Reader for reading.
func (r *MFRC522Reader) Initialize() error {
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

	err = r.device.SetAntennaGain(7)
	if err != nil {
		log.Println("Failed to set antenna gain.")
		return err
	}

	log.Println("Successfully initialized Reader.")
	return nil
}

// Halt stops the Reader.
func (r *MFRC522Reader) Halt() error {
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
func (r *MFRC522Reader) ReadUID(timeout time.Duration) ([]byte, error) {
	return r.device.ReadUID(timeout)
}

// ReadDataBlocks reads all data blocks from a given sector.
//
// Returns a total of 16 bytes/block * 3 blocks = 48 bytes.
func (r *MFRC522Reader) ReadDataBlocks(timeout time.Duration, sector int) (data []byte, err error) {
	b0, err := r.ReadDataBlock(timeout, sector, 0)
	if err != nil {
		log.Printf("Failed to read sector=%d block=%d: %s", sector, 0, err)
		return nil, err
	}

	b1, err := r.ReadDataBlock(timeout, sector, 1)
	if err != nil {
		log.Printf("Failed to read sector=%d block=%d: %s", sector, 1, err)
		return nil, err
	}

	b2, err := r.ReadDataBlock(timeout, sector, 2)
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

// ReadDataBlock reads a single block from a single sector.
//
// Returns a total of 16 bytes/block * 3 blocks = 48 bytes.
func (r *MFRC522Reader) ReadDataBlock(timeout time.Duration, sector int, block int) (data []byte, err error) {
	if sector < 0 || sector >= NumSectors {
		return nil, fmt.Errorf("invalid sector: %d", sector)
	}
	if block < 0 || block >= NumDataBlocksPerSector {
		return nil, fmt.Errorf("invalid block: %d", block)
	}

	var auth byte = commands.PICC_AUTHENT1B
	return r.device.ReadCard(timeout, auth, sector, block, key)
}

// ReadAuthBlock reads the keys and permissions bits for a given sector (aka the "sector trailer").
func (r *MFRC522Reader) ReadAuthBlock(timeout time.Duration, sector int) (authBlock *AuthBlock, err error) {
	var auth byte = commands.PICC_AUTHENT1B
	data, err := r.device.ReadAuth(timeout, auth, sector, key)
	if err != nil {
		log.Printf("Failed to read authentication block.")
		return nil, err
	}

	var keyA [6]byte
	copy(keyA[0:6], data[0:6])

	var keyB [6]byte
	copy(keyB[0:6], data[10:16])

	var permissions mfrc522.BlocksAccess = mfrc522.BlocksAccess{}
	permissions.Init(data)

	return &AuthBlock{KeyA: keyA, KeyB: keyB, Permissions: permissions}, nil
}

// String returns a human-readable string describing this Reader.
func (r *MFRC522Reader) String() string {
	return r.device.String()
}
