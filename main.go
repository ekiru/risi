package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	ReadItems *rss.ItemSet
	UnreadItems *rss.ItemSet
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
		if flag.NArg() != 1 {
			usage(os.Stderr, "check")
			os.Exit(1)
		}
		for i, feed := range data.Feeds {
			feed.LastChecked = time.Now().Local()
			doc, err := rss.ParseFromUrl(feed.Url)
			dieIfErr(err, "Unable to check feed %s", feed.Url)
			allItems := rss.NewItemSetFromSlice(doc.Channel.Items)
			oldUnreadCount := feed.UnreadItems.Count()
			feed.ReadItems = allItems.Intersection(feed.ReadItems)
			feed.UnreadItems = feed.UnreadItems.Union(allItems.Without(feed.ReadItems))
			fmt.Printf("%s: %d unread items, %d new\n", feed.Url, feed.UnreadItems.Count(), feed.UnreadItems.Count()-oldUnreadCount)
			data.Feeds[i] = feed
			data.Dirty = true

		}
	case "feeds":
		for i, feed := range data.Feeds {
			fmt.Printf("%d\t%s\t%d unread\tlast checked at %s\n",
				i, feed.Url, feed.UnreadItems.Count(), feed.LastChecked.Format(time.UnixDate))
		}
	case "next":
		if flag.NArg() > 2 {
			usage(os.Stderr, "next [index]")
			os.Exit(1)
		}
		if flag.NArg() == 2 {
			i, feed := getFeed(data, flag.Arg(1))
			nextInFeed(&data, i, feed)
		} else {
			found := false
			for i, feed := range data.Feeds {
				if feed.UnreadItems.Count() != 0 {
					nextInFeed(&data, i, feed)
					found = true
					break
				}
			}
			if !found {
				fmt.Println("no unread")
			}
		}
	case "read":
		if flag.NArg() != 2 {
			usage(os.Stderr, "next index")
			os.Exit(1)
		}
		i, feed := getFeed(data, flag.Arg(1))
		howMany := feed.UnreadItems.Count()
		feed.ReadItems = feed.ReadItems.Union(feed.UnreadItems)
		feed.UnreadItems = rss.NewItemSet()
		data.Feeds[i] = feed
		data.Dirty = true
		fmt.Printf("%d marked read\n", howMany)
	case "subscribe":
		if flag.NArg() != 2 {
			usage(os.Stderr, "subscribe feed")
			os.Exit(1)
		}
		data.Dirty = true
		data.Feeds = append(data.Feeds, Feed{
			Url:         flag.Arg(1),
			LastChecked: time.Unix(0, 0).Local(),
			ReadItems:   rss.NewItemSet(),
			UnreadItems: rss.NewItemSet(),
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

func nextInFeed(data *Data, i int, feed Feed) {
	if feed.UnreadItems.Count() == 0 {
		fmt.Println("no unread")
	} else {
		item := feed.UnreadItems.Earliest()
		feed.UnreadItems.Remove(item)
		feed.ReadItems.Add(item)
		fmt.Println(item.Link)
		data.Feeds[i] = feed
		data.Dirty = true

	}
}

func getFeed(data Data, is string) (i int, feed Feed) {
	i, err := strconv.Atoi(flag.Arg(1))
	dieIfErr(err, "Feed indices must be integers")
	if i < 0 || i >= len(data.Feeds) {
		die("Feed indices out of range")
	}
	feed = data.Feeds[i]
	return
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
