package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// CRAFTDOOR_ROOT_VAR is an environment variable containing root directory for craftdoor.
const CRAFTDOOR_ROOT_VAR = "CRAFTDOOR_ROOT"

// Config represents the config file's contents.
//
// Filepaths may reference environment variable CRAFTDOOR_ROOT when resolving paths.
type Config struct {
	// Path to SQLite database.
	SQLiteFile string `json:"sqlite_file"`

	// Path to schema.sql used to initialize SQLite database.
	SQLiteSchemaFile string `json:"sqlite_schema_file"`

	// Port for REST API.
	ListenHTTP string `json:"listen_http"`
}

// InitializeConfig reads a JSON config file and decodes it as type Config.
//
// If unspecified, sets $CRAFTDOOR_ROOT to the directory of this binary.
func InitializeConfig(configPath string) (*Config, error) {
	config, err := ReadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	_, isDefined := os.LookupEnv(CRAFTDOOR_ROOT_VAR)
	if !isDefined {
		log.Printf("%s not defined. Setting value to path of binary.", CRAFTDOOR_ROOT_VAR)
		cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return nil, err
		}
		os.Setenv(CRAFTDOOR_ROOT_VAR, cwd)
	}

	// Expand environment variables.
	config.SQLiteFile = os.ExpandEnv(config.SQLiteFile)
	config.SQLiteSchemaFile = os.ExpandEnv(config.SQLiteSchemaFile)

	// Print out final config.
	marshalledText, _ := json.MarshalIndent(config, "", " ")
	fmt.Println(string(marshalledText))

	return &config, nil
}

// ReadConfigFile reads the contents of a JSON file  and decodes it as a Config object.
func ReadConfigFile(filename string) (Config, error) {
	log.Printf("reading config from '%s'", filename)
	// #nosec G304
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer func() {
		e := f.Close()
		if e != nil {
			log.Printf("failed closing config file: %s", e.Error())
		}
	}()

	cfg := Config{}
	return cfg, json.NewDecoder(f).Decode(&cfg)
}
