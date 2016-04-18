package rss

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
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
	PubDateString string `xml:"pubDate"`
	PubDate time.Time
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
	if err != nil {
		return
	}
	for i, item := range feed.Channel.Items {
		item.PubDate, err = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", item.PubDateString)
		if err != nil {
			return
		}
		feed.Channel.Items[i] = item
	}
	return
}
