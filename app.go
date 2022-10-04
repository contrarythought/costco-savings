package main

import (
	"costco_savings/net_help"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	URL = "https://www.costco.com/warehouse-savings.html"
)

type Date struct {
	Day   int        `json:"day"`
	Month time.Month `json:"month"`
	Year  int        `json:"year"`
}

func NewDate() *Date {
	return &Date{
		Day:   time.Now().Day(),
		Month: time.Now().Month(),
		Year:  time.Now().Year(),
	}
}

func main() {
	working_dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	wd_entries, err := os.ReadDir(working_dir)
	if err != nil {
		log.Fatal(err)
	}

	// if time.json is not in wd, create it
	var have_file bool = false
	for _, entry := range wd_entries {
		if strings.Compare(entry.Name(), "time.json") == 0 {
			have_file = true
			break
		}
	}

	var time_file *os.File = nil
	if !have_file {
		time_file, err = os.Create("time.json")
		if err != nil {
			log.Fatal(err)
		}
		defer time_file.Close()

		time_map := make(map[string]Date)
		time_map["date"] = *NewDate()

		// convert map to json
		json_buf, err := json.MarshalIndent(time_map, "", "	")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(time_file, string(json_buf))
	} else {
		// run program if time.json was recorded yesterday
	}

	/* put all of this into a separate run() function inside of the else {} above */
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	net_help.SetHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var resp_html []byte

	switch resp.Header.Get("content-encoding") {
	case "gzip":
		resp_html, err = net_help.DecodeGzip(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	case "compress":
		if strings.Compare(path.Ext(URL), ".tiff") == 0 ||
			strings.Compare(path.Ext(URL), ".tif") == 0 ||
			strings.Compare(path.Ext(URL), ".pdf") == 0 {
			resp_html, err = net_help.DecodeLZW(resp.Body, "msb", 8)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// need to do more research
			resp_html, err = net_help.DecodeLZW(resp.Body, "lsb", 8)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "deflate":
		resp_html, err = net_help.DecodeDeflate(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	html_file, err := os.Create("savings3.html")
	if err != nil {
		log.Fatal(err)
	}
	defer html_file.Close()

	_, err = fmt.Fprintln(html_file, string(resp_html))
	if err != nil {
		log.Fatal(err)
	}
}
