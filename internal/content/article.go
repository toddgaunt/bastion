package content

import (
	"fmt"
	"html/template"
	"time"
)

// Article represents an article served by the bastion webservice.
type Article struct {
	// Filepath is the canonical filepath leading the document source.
	FilePath string

	// Route is the route to the article on the website.
	Route string

	// Content of the article
	Title       string
	Description string
	Author      string
	Pinned      bool
	Created     time.Time
	Updated     time.Time

	// Original text content of the article
	Text []byte

	// HTML generated from the original article source
	HTML template.HTML

	Err error
}

// FormattedDate returns a formatted date string from an article's Created
// field.
func (article Article) FormattedDate() string {
	return article.Created.Format("2006-01-02")
}

// SetTimestamps parses string timestamps and converts them to timestamps to
// set in the article.
func (a *Article) SetTimestamps(created string, updated string) {
	if created != "" {
		t, err := time.Parse("2006-01-02", created)
		if err != nil {
			a.Err = fmt.Errorf("could not parse 'created': %w", err)
		}
		a.Created = t
	}
	if updated != "" {
		t, err := time.Parse("2006-01-02", updated)
		if err != nil {
			a.Err = fmt.Errorf("couldn't parse 'updated': %w", err)
		}
		a.Updated = t
	}
}
