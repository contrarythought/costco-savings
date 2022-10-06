package main

import (
	"costco_savings/app"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var (
	URL = "https://www.costco.com/warehouse-savings.html"
)

func main() {
	var db *sql.DB = nil
	have_file, err := app.SearchForTimeFile()
	if err != nil {
		log.Fatal(err)
	}
	if !have_file {
		// connect to database if program has never been run before
		creds := app.NewCreds()
		err = creds.GetCreds("creds.json")
		if err != nil {
			log.Fatal(err)
		}
		db, err = sql.Open("postgres", "post://"+creds.User+":"+creds.Password+"@127.0.0.1:5432/costco_savings?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
		err = app.SetupTimeFile()
		if err != nil {
			log.Fatal(err)
		}
		app.Run(db)
	} else {
		// run program if time.json was modified at an earlier date
		if result, err := app.RanToday(); !result && err == nil {
			app.Run(db)
		} else if err != nil {
			log.Fatal(err)
		}
	}
}
