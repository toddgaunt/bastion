package content

import (
	"fmt"
	"html/template"
	"time"
)

type Article struct {
	Route       string
	Title       string
	Description string
	Author      string
	Pinned      bool
	Created     time.Time
	Updated     time.Time

	HTML     template.HTML
	Markdown string
	Err      error
}

func (article Article) FormattedDate() string {
	return article.Created.Format("2006-01-02")
}

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
