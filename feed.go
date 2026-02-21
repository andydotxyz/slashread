package main

import (
	"encoding/xml"
	"fmt"
	"io"

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
	Section     string `xml:"section"`
	Subject     string `xml:"subject"`
	Date        string `xml:"date"`
	Creator     string `xml:"creator"`
	Department  string `xml:"department"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

func (i Item) ImageURL() string {
	src := i.Subject

	return fmt.Sprintf("https://a.fsdn.com/sd/topics/%s_64.png", src)
}

func readFeed(r io.Reader) (*RSS, error) {
	rss := &RSS{}
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel

	return rss, decoder.Decode(rss)
}
