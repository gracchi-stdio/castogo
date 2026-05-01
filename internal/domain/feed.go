package domain

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Length  int64    `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
}

type PubDate struct {
	time.Time
}

func (p PubDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	e.EncodeToken(xml.CharData(p.Format(time.RFC1123Z)))
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	totalSeconds := int64(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	durStr := ""
	if hours > 0 {
		durStr += fmt.Sprintf("%d:", hours)
	}
	durStr += fmt.Sprintf("%02d:%02d", minutes, seconds)

	e.EncodeToken(start)
	e.EncodeToken(xml.CharData(durStr))
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

type ITunesOwner struct {
	XMLName xml.Name `xml:"itunes:owner"`
	Name    string   `xml:"itunes:name"`
	Email   string   `xml:"itunes:email"`
}

type ITunesImage struct {
	XMLName xml.Name `xml:"itunes:image"`
	Href    string   `xml:"href,attr"`
}

type ITunesCategory struct {
	XMLName  xml.Name        `xml:"itunes:category"`
	Text     string          `xml:"text,attr"`
	SubItems *ITunesCategory `xml:"itunes:category,omitempty"`
}

type GUID struct {
	XMLName     xml.Name `xml:"guid"`
	IsPermaLink string   `xml:"isPermaLink,attr"`
	Value       string   `xml:",chardata"`
}

type Item struct {
	XMLName        xml.Name     `xml:"item"`
	Title          string       `xml:"title"`
	Description    string       `xml:"description"`
	ContentEncoded string       `xml:"content:encoded,omitempty"`
	Link           string       `xml:"link,omitempty"`
	GUID           GUID         `xml:"guid"`
	PubDate        *PubDate     `xml:"pubDate,omitempty"`
	Enclosure      Enclosure    `xml:"enclosure"`
	ITunesDuration Duration     `xml:"itunes:duration"`
	ITunesEpisode  *int         `xml:"itunes:episode,omitempty"`
	ITunesExplicit string       `xml:"itunes:explicit,omitempty"`
	ITunesAuthor   string       `xml:"itunes:author,omitempty"`
	ITunesSummary  string       `xml:"itunes:summary,omitempty"`
	ITunesImage    *ITunesImage `xml:"itunes:image,omitempty"`
}

type Channel struct {
	XMLName        xml.Name        `xml:"channel"`
	Title          string          `xml:"title"`
	Description    string          `xml:"description"`
	Link           string          `xml:"link,omitempty"`
	Language       string          `xml:"language,omitempty"`
	Copyright      string          `xml:"copyright,omitempty"`
	ITunesAuthor   string          `xml:"itunes:author,omitempty"`
	ITunesType     string          `xml:"itunes:type,omitempty"`
	ITunesExplicit string          `xml:"itunes:explicit,omitempty"`
	Owner          *ITunesOwner    `xml:"itunes:owner,omitempty"`
	Image          *ITunesImage    `xml:"itunes:image,omitempty"`
	Category       *ITunesCategory `xml:"itunes:category,omitempty"`
	PodcastLocked  string          `xml:"podcast:locked,omitempty"`
	Generator      string          `xml:"generator,omitempty"`
	Items          []Item          `xml:"item"`
}

type RSS struct {
	XMLName   xml.Name `xml:"rss"`
	Version   string   `xml:"version,attr"`
	NSItunes  string   `xml:"xmlns:itunes,attr"`
	NSPodcast string   `xml:"xmlns:podcast,attr"`
	NSContent string   `xml:"xmlns:content,attr"`
	Channel   Channel  `xml:"channel"`
}

// Write writes the RSS feed as XML to w.
func (r *RSS) Write(w io.Writer) error {
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(r)
}
