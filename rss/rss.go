package rss

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type FeedDoc struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Description string `xml:"description"`
	Title       string `xml:"title"`
	Generator   string `xml:"generator"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	Guid    string `xml:"guid"`
	PubDate string `xml:"pubDate"`
}

func ParseFromUrl(url string) (feed FeedDoc, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = xml.Unmarshal(buf, &feed)
	return
}
