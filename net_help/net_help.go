package net_help

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"
	"net/http"
)

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
