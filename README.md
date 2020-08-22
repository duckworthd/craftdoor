![Go](https://github.com/pakohan/craftdoor/workflows/Go/badge.svg)

# craftdoor

A RFID based access control system written in Go for Raspberry Pi + MFRCC522
tag readers + MIFARE RFID tags.

# Project Overview

Craftdoor is a software suite for an RFID-powered door access system on a
federation of Raspberry Pi devices. With the exception of the "master", each
Raspberry Pi is connected to an [RFID
reader](https://www.nxp.com/docs/en/data-sheet/MFRC522.pdf) and a door.
Registered members may tap the RFID reader with their MIFARE RFID tag to open
the adjacent door.

The system is administered via a WebUI interface and accompanying REST API
served by the master device. Persistent state is stored in a SQLite database on
the master device. See below for valid endpoints.

Instructions below for building, configuring, and launching the webserver.

**Note**: At time of writing, only a single "master" Raspberry Pi is supported.

# Installation

To start the software suite, do the following on your development machine,

1. Connect RC522 to master Raspberry Pi's hardware SPI interface. Follow
   instructions [here](https://github.com/pakohan/craftdoor.git).
1. Download `golang` from https://golang.org. Follow installation instructions
   [here](https://golang.org/doc/install#install). Verify that go is installed
   by running `go version` in a terminal. Expect to see >= 1.14.
1. Install GCC cross-compiler,
  ```
  $ sudo apt install gcc-arm-linux-gnueabi libc6-armel-cross \
    libc6-dev-armel-cross binutils-arm-linux-gnueabi
  ```
1. Run `scripts/build.sh`. This will create a folder, `release/`, containing
   everything needed to run the server on a Raspberry pi
   ```
   $ bash scripts/build.sh
   ```
1. Turn on your Raspberry Pi and ensure it's connected to your local network.
   Follow the instructions
   [here](https://www.raspberrypi.org/documentation/remote-access/ssh/unix.md).
   Ensure that you can SSH into the Raspberry Pi under hostname `raspberrypi`.
   You can setup an SSH config entry like so in `$HOME/.ssh/config`,
   ```
   Host raspberrypi
     Hostname 192.168.0.9  # Your IP address may vary!
     User pi
   ```
1. Copy the contents of `release/` to the Raspberry Pi and run it. This will
   launch a webserver on port :8080.
   ```
   $ bash scripts/deploy.sh
   ```

**Note**: If the RC522 RFID reader is not
[detected](http://pkg.go.dev/periph.io/x/periph/host/rpi#Present), a fake,
dummy interface will be used. This dummy interface cannot interact with RFID
tags.

If you'd like to ensure your code doesn't crash on your development machine,
you can run the following. Note that the only the dummy interface is available
in this context.

1. Run `cmd/master/main.go`. This will launch a webserver listening on port 8080.
  ```
  $ git clone https://github.com/pakohan/craftdoor.git
  $ cd craftdoor/cmd/master
  $ go run main.go develop.json
  ```

# Usage

Once `main.go` is launched, the following endpoints are available via the HTTP
webserver,

- `GET /`: Get the details of the next RFID tag put in front of the reader.

For doors,

- `GET /members`: list doors in database
- `GET /members/<id>`: get detailed information about a single member.
- `POST /members`: Create a new member.
- `PUT /members/<id>`: Update an existing member.
- `DELETE /members/<id>`: Delete an existing member.

Similar to doors, one can query and manage keys via `/keys`.

# Code Organization

```
cmd/
  debug/
    read.go          # debug binary for reading all data on an RFID tag.
  master/
    develop.db       # sqlite database used during development
    develop.json     # JSON config used during development
    main.go          # main binary for this project
    schema.sql       # SQL for initializing develop.db
config/
  config.go          # JSON config file API
controller/
  controller.go      # HTTP request handling logic.
  ...
lib/
  db.go              # initialize database schema
  state.go           # State of the system.
model/               # database definitions, API
  model.go           # interface for interacting with the database.
  ...
rfid/                # wrapper for RFID readers/writers
  dummy.go           # dummy implementation of interface Reader
  fmt.go             # format contents of an RFID tag as a string.
  mfrc522.go         # MFRC522 implementation of interface Reader
  reader.go          # Interface for interacting with RFID readers.
service/             # business logic for adding/removing keys, doors, etc
  service.go         # door-opening loop, access to RFID reader.
vendor/              # third-party code
  ...
```
