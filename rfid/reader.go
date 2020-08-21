package rfid

import (
	"time"
)

// NumSectors is the number of sectors on a MIFARE Classic 1K card.
var NumSectors = 16

// NumDataBlocksPerSector is the number of data blocks per sector on a MIFARE Classic 1K card.
var NumDataBlocksPerSector = 3

// NumBytesPerBlock is the number of bytes per block for all blocks in a sector.
var NumBytesPerBlock = 16

// Reader accesses a RFID reader
type Reader interface {
	Initialize() error
	Halt() error
	ReadUID(timeout time.Duration) ([]byte, error)
	ReadDataBlocks(timeout time.Duration, sector int) (data []byte, err error)
	ReadDataBlock(timeout time.Duration, sector int, block int) (data []byte, err error)
	ReadAuthBlock(timeout time.Duration, sector int) (authBlock *AuthBlock, err error)
	String() string
}
