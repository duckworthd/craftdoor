package rfid

import (
	"encoding/hex"
	"fmt"
	"strings"

	"periph.io/x/periph/experimental/devices/mfrc522"
)

// FmtSector creates a string representation of a single sector.
func FmtSector(sector int, data []byte, auth *AuthBlock) string {
	var builder strings.Builder
	for i := 0; i < NumDataBlocksPerSector; i++ {
		start := i * NumBytesPerBlock
		end := (i + 1) * NumBytesPerBlock
		line := fmt.Sprintf("%02d.%d | %s\n", sector, i, FmtBlock(data[start:end]))
		builder.WriteString(line)
	}
	keyA := FmtKey(auth.KeyA)
	keyB := FmtKey(auth.KeyB)
	line := fmt.Sprintf("     | %s\n", FmtAuthBlock(auth))
	builder.WriteString(line)
	return builder.String()
}

// FmtBlock creates a string representation for a single 16-byte block of data.
func FmtBlock(data []byte) string {
	var builder strings.Builder
	for i := 0; i < NumBytesPerBlock; i++ {
		builder.WriteString(fmt.Sprintf("%s", hex.EncodeToString(data[i:i+1])))
		if i != NumBytesPerBlock-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

// FmtAuthBlock creates a string representation for an AuthBlock.
func FmtAuthBlock(auth *AuthBlock) string {
	return fmt.Sprintf("KeyA=%s KeyB=%s BlocksAccess=(%s)\n", FmtKey(auth.KeyA), FmtKey(auth.KeyB), FmtBlocksAccess(auth.Permissions))
}

// FmtKey converts a Key to a string.
func FmtKey(key mfrc522.Key) string {
	var bytes []byte
	bytes = make([]byte, 6)
	for i := 0; i < 6; i++ {
		bytes[i] = key[i]
	}
	return hex.EncodeToString(bytes)
}

// FmtBlocksAccess creates a string representation for a BlocksAccess.
func FmtBlocksAccess(blocksAccess mfrc522.BlocksAccess) string {
	return blocksAccess.String()
}
