package main

import (
	"costco_savings/app"
	"fmt"
	"log"
)

var (
	URL = "https://www.costco.com/warehouse-savings.html"
)

func main() {
	have_file, err := app.SearchForTimeFile()
	if err != nil {
		log.Fatal(err)
	}
	if !have_file {
		err = app.SetupTimeFile()
		if err != nil {
			log.Fatal(err)
		}
		app.Run()
	} else {
		// run program if time.json was modified at an earlier date
		if result, err := app.RanToday(); !result && err == nil {
			fmt.Println("here")
			app.Run()
		} else if err != nil {
			fmt.Println("err here")
			log.Fatal(err)
		}
	}
}
