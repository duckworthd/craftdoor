package lib

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pakohan/craftdoor/config"
)

// OpenDB opens the database.
func OpenDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", cfg.SQLiteFile)
	if err != nil {
		return nil, err
	}

	err = InitDBSchema(db, cfg.SQLiteSchemaFile)
	if err != nil {
		e := db.Close()
		if e != nil {
			log.Printf("err closing db after initializing schema failed: %s", e.Error())
		}
		return nil, err
	}

	return db, nil
}

// InitDBSchema initializes the DB schema if the sqlite_master table has no entries.
func InitDBSchema(db *sqlx.DB, schemaFile string) error {
	var count int
	err := db.Get(&count, checkTables)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Printf("didn't find any tables, will init db schema")

	f, err := os.Open(schemaFile)
	if err != nil {
		return err
	}
	defer func() {
		e := f.Close()
		if e != nil {
			log.Printf("failed closing schema file: %s", e.Error())
		}
	}()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(b))
	return err
}

const checkTables = `
SELECT COUNT(*)
FROM sqlite_master
WHERE type='table'
ORDER BY name`
