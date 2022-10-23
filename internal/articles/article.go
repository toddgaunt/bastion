package articles

import (
	"fmt"
	"html/template"
	"sync"
	"time"
)

type ArticleMap struct {
	Mutex    sync.RWMutex
	Articles map[string]*Article
}

type Article struct {
	Route       string
	Title       string
	Description string
	Author      string
	Created     time.Time
	Updated     time.Time

	HTML     template.HTML
	Markdown string
	Error    error
}

func (article Article) FormattedDate() string {
	return article.Created.Format("2006-01-02")
}

// SetTimestamps will set the provided timestamp strings, if not empty, in the
// article. If they fail to be parsed as timestamps, the article's problem
// field is set with a relevant ProblemJSON object.
func (a *Article) SetTimestamps(created string, updated string) {
	if created != "" {
		t, err := time.Parse("2006-01-02", created)
		if err != nil {
			a.Error = fmt.Errorf("could not parse 'created': %w", err)
		}
		a.Created = t
	}
	if updated != "" {
		t, err := time.Parse("2006-01-02", updated)
		if err != nil {
			a.Error = fmt.Errorf("couldn't parse 'updated': %w", err)
		}
		a.Updated = t
	}
}
