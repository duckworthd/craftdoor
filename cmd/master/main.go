// Craftdoor server.
//
// Launches a binary that does the following,
// - Launches a REST API for managing a database of members, keys
// - Launches an infinite loop for authenticating door access.
//
// Example Usage:
// $ export CRAFTDOOR_ROOT_VAR="$(pwd)/assets"
// $ go run cmd/master/main.go --config="${CRAFTDOOR_ROOT}/develop.json"
//
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pakohan/craftdoor/config"
	"github.com/pakohan/craftdoor/controller"
	"github.com/pakohan/craftdoor/door"
	"github.com/pakohan/craftdoor/lib"
	"github.com/pakohan/craftdoor/model"
	"github.com/pakohan/craftdoor/rfid"
	"github.com/pakohan/craftdoor/service"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

func main() {
	log.SetFlags(log.Llongfile)

	// Command line flags.
	configPath := flag.String("config", "./develop.json", "Path to config file.")
	flag.Parse()

	// Read config.
	cfg, err := config.InitializeConfig(*configPath)
	if err != nil {
		log.Panic(err)
	}

	db, err := lib.OpenDB(cfg)
	if err != nil {
		log.Panic(err)
	}

	// TODO(duckworthd): Shut down database gracefully.

	err = start(cfg, db)
	if err != nil {
		// c <- os.Interrupt
		log.Panic(err)
	}

}

func start(cfg *config.Config, db *sqlx.DB) error {
	// Initialize RFID reader, door.
	var r rfid.Reader
	var d door.Door
	var err error
	if rpi.Present() {
		log.Printf("Initializing rpi.")
		_, err = host.Init()
		if err != nil {
			return err
		}

		log.Printf("Initializing rpi reader.")
		r, err = rfid.NewMFRC522Reader()
		if err != nil {
			return err
		}

		err = r.Initialize()
		if err != nil {
			return err
		}

		log.Printf("Initializing rpi door.")
		d, err = door.NewRPiDoor()
		if err != nil {
			return err
		}

	} else {
		log.Printf("Initializing dummy reader")
		r, err = rfid.NewDummyReader()
		if err != nil {
			return err
		}

		log.Printf("Initializing dummy door")
		d, _ = door.NewDummyDoor()
		if err != nil {
			return err
		}
	}

	// Setup backend database, etc.
	m := model.New(db)
	s := service.New(m, r, d)
	c := controller.New(cfg, m, s)

	// Start HTTP server.
	srv := http.Server{
		Addr:    cfg.ListenHTTP,
		Handler: c,
	}

	// TODO(duckworthd): Figure out how to get outbound IP address when the
	// network doesn't reach the internet.
	log.Printf("listening on port: %s", cfg.ListenHTTP)
	err = srv.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}
