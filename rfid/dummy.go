package rfid

import (
	"time"

	"periph.io/x/periph/experimental/devices/mfrc522"
)

// DummyReader does nothing.
type DummyReader struct{}

// NewDummyReader returns a new DummyReader instance.
func NewDummyReader() (*DummyReader, error) {
	return &DummyReader{}, nil
}

// Initialize does nothing.
func (r *DummyReader) Initialize() error {
	return nil
}

// Halt does nothing.
func (r *DummyReader) Halt() error {
	return nil
}

// ReadUID returns an all-zero UID.
func (r *DummyReader) ReadUID(timeout time.Duration) ([]byte, error) {
	return make([]byte, 16), nil
}

// ReadDataBlocks returns all-zero data blocks.
func (r *DummyReader) ReadDataBlocks(timeout time.Duration, sector int) (data []byte, err error) {
	return make([]byte, NumBytesPerBlock*NumDataBlocksPerSector), nil

}

// ReadDataBlock returns an all-zero data block.
func (r *DummyReader) ReadDataBlock(timeout time.Duration, sector int, block int) (data []byte, err error) {
	return make([]byte, NumBytesPerBlock), nil
}

// ReadAuthBlock returns an all-zero AuthBlock.
func (r *DummyReader) ReadAuthBlock(timeout time.Duration, sector int) (authBlock *AuthBlock, err error) {
	result := &AuthBlock{
		KeyA:        mfrc522.Key{},
		KeyB:        mfrc522.DefaultKey,
		Permissions: mfrc522.BlocksAccess{},
	}
	return result, nil
}

// String returns a string representation of a DummyReader.
func (r *DummyReader) String() string {
	return "DummyReader"
}
