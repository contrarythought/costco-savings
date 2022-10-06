package app

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	URL = "https://www.costco.com/warehouse-savings.html"
)

type Creds struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func NewCreds() *Creds {
	return &Creds{
		User:     "",
		Password: "",
	}
}

func (c *Creds) GetCreds(cred_file string) error {
	file, err := os.Open(cred_file)
	if err != nil {
		return err
	}
	defer file.Close()
	buf, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, c)
	if err != nil {
		return err
	}
	return nil
}

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

type Item struct {
	Name     string
	OldPrice float32
	NewPrice float32
}

func NewItem(Name string, OldPrice float32, NewPrice float32) *Item {
	return &Item{
		Name:     Name,
		OldPrice: OldPrice,
		NewPrice: NewPrice,
	}
}

func SetHeaders(req *http.Request) {
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-encoding", "gzip, deflate, br")
	req.Header.Set("accept-language", "en,en-US;q=0.9,zh-TW;q=0.8,zh;q=0.7")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("referer", "https://www.google.com/")
	req.Header.Set("sec-ch-ua", "\"Google Chrome\";v=\"105\", \"Not)A;Brand\";v=\"8\", \"Chromium\";v=\"105\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "Windows")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36")
}

func SearchForTimeFile() (bool, error) {
	working_dir, err := os.Getwd()
	if err != nil {
		return false, err
	}

	wd_entries, err := os.ReadDir(working_dir)
	if err != nil {
		return false, err
	}

	// if time.json is not in wd, create it
	have_file := false
	for _, entry := range wd_entries {
		if strings.Compare(entry.Name(), "time.json") == 0 {
			have_file = true
			break
		}
	}

	return have_file, nil
}

func SetupTimeFile() error {
	time_file, err := os.Create("time.json")
	if err != nil {
		return err
	}
	defer time_file.Close()
	curr_date := NewDate()
	json_fmt, err := json.MarshalIndent(curr_date, "", "	")
	if err != nil {
		return err
	}
	fmt.Fprint(time_file, string(json_fmt))
	return nil
}

func Run(db *sql.DB) {
	// modify time.json
	time_file, err := os.Create("time.json")
	if err != nil {
		log.Fatal(err)
	}
	defer time_file.Close()

	curr_date := NewDate()

	json_fmt, err := json.MarshalIndent(curr_date, "", "	")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(time_file, string(json_fmt))

	// create and send new request
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	SetHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var resp_html []byte

	// determine response encoding and decode
	switch resp.Header.Get("content-encoding") {
	case "gzip":
		resp_html, err = DecodeGzip(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	case "compress":
		if strings.Compare(path.Ext(URL), ".tiff") == 0 ||
			strings.Compare(path.Ext(URL), ".tif") == 0 ||
			strings.Compare(path.Ext(URL), ".pdf") == 0 {
			resp_html, err = DecodeLZW(resp.Body, "msb", 8)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// need to do more research
			resp_html, err = DecodeLZW(resp.Body, "lsb", 8)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "deflate":
		resp_html, err = DecodeDeflate(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	// record response html into a file
	html_file, err := os.Create("savings3.html")
	if err != nil {
		log.Fatal(err)
	}
	defer html_file.Close()

	_, err = fmt.Fprintln(html_file, string(resp_html))
	if err != nil {
		log.Fatal(err)
	}

	// insert items into database
}

// TODO
func GetSaleItems(html_file *os.File) {

}

func RanToday() (bool, error) {
	time_file, err := os.Open("time.json")
	if err != nil {
		return false, err
	}
	defer time_file.Close()

	buf, err := io.ReadAll(time_file)
	if err != nil {
		return false, err
	}

	var old_date Date
	err = json.Unmarshal(buf, &old_date)
	if err != nil {
		return false, err
	}

	curr_date := NewDate()

	res := true
	if curr_date.Day > old_date.Day {
		res = false
	}
	if curr_date.Month > old_date.Month {
		res = false
	}
	if curr_date.Year > old_date.Year {
		res = false
	}
	return res, nil
}

func DecodeGzip(body io.ReadCloser) ([]byte, error) {
	gz_reader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	defer gz_reader.Close()
	buf, err := io.ReadAll(gz_reader)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func DecodeLZW(body io.ReadCloser, order string, width int) ([]byte, error) {
	switch order {
	case "msb":
		lzw_reader := lzw.NewReader(body, lzw.MSB, width)
		defer lzw_reader.Close()
		buf, err := io.ReadAll(lzw_reader)
		if err != nil {
			return nil, err
		}
		return buf, nil
	case "lsb":
		lzw_reader := lzw.NewReader(body, lzw.LSB, width)
		defer lzw_reader.Close()
		buf, err := io.ReadAll(lzw_reader)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	return nil, nil
}

func DecodeDeflate(body io.ReadCloser) ([]byte, error) {
	flate_reader := flate.NewReader(body)
	defer flate_reader.Close()
	buf, err := io.ReadAll(flate_reader)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
