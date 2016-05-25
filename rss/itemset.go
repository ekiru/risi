package rss

import (
	"encoding/json"
)

type ItemSet struct {
	items map[string]Item
}

func NewItemSet() *ItemSet {
	return &ItemSet{items: map[string]Item{}}
}

func NewItemSetFromSlice(items []Item) *ItemSet {
	set := NewItemSet()
	for _, item := range items {
		set.Add(item)
	}
	return set
}

func NewItemSetFromAtomSlice(entries []AtomEntry) *ItemSet {
	set := NewItemSet()
	for _, entry := range entries {
		item := Item{
			Title: entry.Title,
			Link: entry.Link,
			Guid: entry.Id,
			PubDate: entry.PubDate,
		}
		set.Add(item)
	}
	return set
}

func (set *ItemSet) Add(item Item) {
	if !set.Contains(item) {
		set.items[item.Guid] = item
	}
}

func (set ItemSet) Contains(item Item) bool {
	_, found := set.items[item.Guid]
	return found
}

func (set ItemSet) Count() int {
	return len(set.items)
}

func (set ItemSet) Earliest() (earliest Item) {
	first := true
	for _, item := range set.items {
		if first || item.PubDate.Before(earliest.PubDate) {
			earliest = item
			first = false
		}
	}
	return
}

func (set ItemSet) Intersection(other *ItemSet) *ItemSet {
	inter := NewItemSet()
	for _, item := range set.items {
		if other.Contains(item) {
			inter.Add(item)
		}
	}
	return inter
}

func (set ItemSet) Latest() (latest Item) {
	first := true
	for _, item := range set.items {
		if first || latest.PubDate.Before(item.PubDate) {
			latest = item
			first = false
		}
	}
	return
}

func (set *ItemSet) Remove(item Item) {
	delete(set.items, item.Guid)
}

func (set ItemSet) Union(other *ItemSet) *ItemSet {
	union := NewItemSet()
	for _, item := range set.items {
		union.Add(item)
	}
	for _, item := range other.items {
		union.Add(item)
	}
	return union
}

func (set ItemSet) Without(other *ItemSet) *ItemSet {
	diff := NewItemSet()
	for _, item := range set.items {
		if !other.Contains(item) {
			diff.Add(item)
		}
	}
	return diff
}

func (set ItemSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.items)
}

func (set *ItemSet) UnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, &set.items)
}