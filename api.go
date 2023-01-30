package bastion

import (
	"fmt"
	"html/template"
	"time"
)

type Logger interface {
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Infow(msg string, keyValues ...any)
	Warnw(msg string, keyValues ...any)
	Errorw(msg string, keyValues ...any)
	Fatalw(msg string, keyValues ...any)
	With(keyValues ...any) Logger
}

type Details struct {
	Name        string
	Description string
	Style       string
}

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

type Content interface {
	Get(key string) (Article, bool)
	GetAll(pinned bool) []Article
	Details() Details
}
