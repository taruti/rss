package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

func parseAtom(data []byte, seen Seen) (*Feed, error) {
	feed := atomFeed{}
	p := xml.NewDecoder(bytes.NewReader(data))
	p.CharsetReader = charsetReader
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}

	out := new(Feed)
	out.Title = feed.Title
	out.Description = feed.Description
	out.Link = feed.Link.Href
	out.Image = feed.Image.Image()
	out.Refresh = time.Now().Add(10 * time.Minute)
	defAuthors := aa2i(feed.Authors)

	if feed.Items == nil {
		return nil, fmt.Errorf("Error: no feeds found in %q.", string(data))
	}

	out.Items = make([]*Item, 0, len(feed.Items))

	// Process items.
	for _, item := range feed.Items {

		id := item.ID
		if id == "" {
			id = item.Link.Href
		}
		if _, found := seen[id]; found || id == "" {
			continue
		}
		seen[id] = struct{}{}

		next := new(Item)
		next.Title = item.Title
		next.Content = item.Content
		next.Link = item.Link.Href
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err != nil {
				return nil, err
			}
		}
		next.ID = id
		next.Read = false
		next.Authors = aa2i(item.Authors)
		if len(next.Authors) == 0 && len(defAuthors) > 0 {
			next.Authors = defAuthors
		}

		out.Items = append(out.Items, next)
		out.Unread++
	}

	return out, nil
}

type atomFeed struct {
	XMLName     xml.Name     `xml:"feed"`
	Title       string       `xml:"title"`
	Description string       `xml:"subtitle"`
	Link        atomLink     `xml:"link"`
	Image       atomImage    `xml:"image"`
	Items       []atomItem   `xml:"entry"`
	Updated     string       `xml:"updated"`
	Authors     []atomAuthor `xml:"author"`
}
type atomLink struct {
	Href string `xml:"href,attr"`
}
type atomItem struct {
	XMLName xml.Name     `xml:"entry"`
	Title   string       `xml:"title"`
	Content string       `xml:"summary"`
	Link    atomLink     `xml:"link"`
	Date    string       `xml:"updated"`
	ID      string       `xml:"id"`
	Authors []atomAuthor `xml:"author"`
}

type atomImage struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Url     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

type atomAuthor struct {
	Name  string `xml:"name"`
	Uri   string `xml:"uri"`
	Email string `xml:"email"`
}

func aa2i(as []atomAuthor) []Author {
	var rs []Author
	for _, a := range as {
		if a != (atomAuthor{}) {
			rs = append(rs, Author{Name: a.Name, Uri: a.Uri, Email: a.Email})
		}
	}
	return rs
}

func (a *atomImage) Image() *Image {
	out := new(Image)
	out.Title = a.Title
	out.Url = a.Url
	out.Height = uint32(a.Height)
	out.Width = uint32(a.Width)
	return out
}
