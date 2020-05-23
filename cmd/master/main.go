package main

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pakohan/craftdoor/config"
	"github.com/pakohan/craftdoor/controller"
	"github.com/pakohan/craftdoor/lib"
	"github.com/pakohan/craftdoor/model"
	"github.com/pakohan/craftdoor/service"
)

func main() {
	log.SetFlags(log.Llongfile | log.Lmicroseconds)

	cfg, err := config.ReadConfig()
	if err != nil {
		log.Panic(err)
	}

	db, err := sqlx.Connect("sqlite3", cfg.SQLiteFile)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = start(cfg, db)
	if err != nil {
		log.Panic(err)
	}
}

func start(cfg config.Config, db *sqlx.DB) error {
	cl := lib.NewChangeListener()

	r, err := lib.NewReader(cfg, cl)
	if err != nil {
		return err
	}

	m := model.New(db)
	s := service.New(m, r, cl)
	c := controller.New(m, s)

	return http.ListenAndServe(cfg.ListenHTTP, c)
}
