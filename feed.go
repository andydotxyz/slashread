package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"

	"fyne.io/fyne/v2"
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

var (
	resources = make(map[string]fyne.Resource)
	resLock   = sync.RWMutex{}
)

func (i Item) ImageResource() fyne.Resource {
	resLock.RLock()
	res, ok := resources[i.Subject]
	resLock.RUnlock()
	if ok {
		return res
	}

	res, err := fyne.LoadResourceFromURLString(i.ImageURL())
	if err != nil {
		fyne.LogError("Failed to read section image", err)
		return nil
	}

	resLock.Lock()
	resources[i.Subject] = res
	resLock.Unlock()
	return res
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
