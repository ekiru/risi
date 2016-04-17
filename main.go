package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"

	"risi/rss"

	"github.com/mitchellh/go-homedir"
)

type Data struct {
	Feeds []Feed
	Dirty bool `json:"-"`
}

type Feed struct {
	Url         string
	LastChecked time.Time
	ReadItems   []string // Guids of seen items, sorted.
	UnreadItems []string
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usage(os.Stderr, "")
		os.Exit(1)
	}
	data, err := loadData()
	dieIfErr(err, "Unable to load database")
	switch cmd := flag.Arg(0); cmd {
	case "check":
		if flag.NArg() != 2 {
			usage(os.Stderr, "check index")
			os.Exit(1)
		}
		i, err := strconv.Atoi(flag.Arg(1))
		dieIfErr(err, "Feed indices must be integers")
		if i < 0 || i >= len(data.Feeds) {
			die("Feed indices out of range")
		}
		feed := data.Feeds[i]
		doc, err := rss.ParseFromUrl(feed.Url)
		dieIfErr(err, "Unable to check feed %s", feed.Url)
		alreadyRead := feed.ReadItems
		feed.ReadItems = nil
		oldUnread := feed.UnreadItems
		for _, item := range doc.Channel.Items {
			if setContains(alreadyRead, item.Guid) {
				feed.ReadItems = append(feed.ReadItems, item.Guid)
			} else if !setContains(oldUnread, item.Guid) {
				feed.UnreadItems = append(feed.UnreadItems, item.Guid)
			}
		}
		fmt.Printf("%d unread items, %d new\n", len(feed.UnreadItems), len(feed.UnreadItems)-len(oldUnread))
		sort.Strings(feed.UnreadItems)
		sort.Strings(feed.ReadItems)
		data.Feeds[i] = feed
		data.Dirty = true
	case "feeds":
		for i, feed := range data.Feeds {
			fmt.Printf("%d\t%s\tlast checked at %s\n", i, feed.Url, feed.LastChecked.Format(time.UnixDate))
		}
	case "subscribe":
		if flag.NArg() != 2 {
			usage(os.Stderr, "subscribe feed")
			os.Exit(1)
		}
		data.Dirty = true
		data.Feeds = append(data.Feeds, Feed{
			Url:         flag.Arg(1),
			LastChecked: time.Unix(0, 0).Local(),
			ReadItems:   []string{},
		})
	default:
		die("Unrecognized command")
	}
	if data.Dirty {
		dieIfErr(saveData(data), "Unable to save database")
	}
}

func setContains(set []string, elem string) bool {
	for {
		if len(set) == 0 {
			return false
		}
		mid := len(set) / 2
		if set[mid] == elem {
			return true
		} else if set[mid] > elem {
			set = set[:mid]
		} else {
			set = set[mid:]
		}

	}
}

var datafileName string

func loadData() (data Data, err error) {
	datafileName, err = homedir.Expand("~/.risi")
	if err != nil {
		return
	}
	_, err = os.Stat(datafileName)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil // return an empty Data since no database file exists yet.
		}
		return
	}
	buf, err := ioutil.ReadFile(datafileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &data)
	return
}

func saveData(data Data) error {
	buf, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(datafileName, buf, 0600)
}

func die(s string) {
	fmt.Fprintln(os.Stderr, s)
	os.Exit(3)
}

func dieIfErr(err error, format string, vs ...interface{}) {
	if err != nil {
		if format != "" {
			fmt.Fprintf(os.Stderr, format, vs...)
		} else {
			fmt.Print("An error occurred")
		}
		fmt.Fprintf(os.Stderr, ": %s\n", err.Error())
		os.Exit(2)
	}
}

func usage(w io.Writer, args string) {
	if args == "" {
		args = "subcommand arg ..."
	}
	fmt.Fprintf(w, "Usage: %s %s\n", os.Args[0], args)
}
