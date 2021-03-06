// Reads the full contents of an RFID tag.
//
// This binary reads all data blocks of all sectors on a MIFARE Classic 1K RFID
// tag, in order. The same key is used to read all blocks and all sectors,
// including the "sector trailer" containing keys and permissions bits.
//
// Prints contents in the following format,
//
// SECTOR.BLOCK | BLOCK DATA
// SECTOR.BLOCK | BLOCK DATA
// SECTOR.BLOCK | BLOCK DATA
//              | KeyA=... KeyB=... BlocksAccess=...
//
// For example, the contents of a new RFID tag might look like this,
//
// 00.0 | 35 c1 70 53 d7 08 04 00 02 71 fa 6b df ae 55 1d
// 00.1 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
// 00.2 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
//      | KeyA=000000000000 KeyB=ffffffffffff BlocksAccess=(B0: 0, B1: 0, B2: 0, B3: 0)
//

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pakohan/craftdoor/rfid"
	"periph.io/x/periph/host"
)

func main() {
	// Make sure periph is initialized.
	log.Println("Initializing host.")
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Initialize RFID Reader.
	log.Println("Creating MFRC522 SPI device.")
	reader, _ := rfid.NewMFRC522Reader()
	err := reader.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Halt()

	done := make(chan struct{})
	defer func() {
		close(done)
	}()

	go func() {
		log.Printf("Starting %s", reader.String())
		sector := 0
		for {
			if sector >= rfid.NumSectors {
				done <- struct{}{}
				return
			}
			timeout := 5 * time.Second

			data, err := reader.ReadDataBlocks(timeout, sector)
			if err != nil {
				log.Printf("Error in ReadSectorData: %s", err)
				if strings.Contains(err.Error(), "mfrc522 lowlevel: IRQ error") {
					// See https://github.com/google/periph/issues/425
					reader.Initialize()
				}
				continue
			}

			auth, err := reader.ReadAuthBlock(timeout, sector)
			if err != nil {
				log.Printf("Error in ReadSectorData: %s", err)
				continue
			}

			fmt.Printf("%s\n", rfid.FmtSector(sector, data, auth))
			sector++
		}
	}()

	for {
		select {
		case <-done:
			return
		}
	}
}
