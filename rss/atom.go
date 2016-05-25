package rss

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
)

type AtomFeedDoc struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomEntry struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Id            string `xml:"id"`
	PubDateString string `xml:"published"`
	PubDate       time.Time
}

func ParseFromAtomUrl(url string) (feed AtomFeedDoc, err error) {
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
	for i, entry := range feed.Entries {
		entry.PubDate, err = time.Parse("2006-01-02T15:04:05-07:00", entry.PubDateString)
		if err != nil {
			return
		}
		feed.Entries[i] = entry
	}
	return
}
