package main

import (
	"encoding/xml"
	"net/http"

	"golang.org/x/net/html/charset"
)

// RSS feed structures
type RSS struct {
	Channel Channel `xml:"channel"`
	Items   []Item  `xml:"item"`
}

type Channel struct {
	Title string `xml:"title"`
}

type Item struct {
	Title       string `xml:"title"`
	Date        string `xml:"date"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

func readFeed(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rss := &RSS{}
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	return rss, decoder.Decode(rss)
}
