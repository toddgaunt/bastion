package content

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
)

// Article represents an article served by the bastion webservice.
type Article struct {
	// Filepath is the canonical filepath leading the document source.
	FilePath string

	// Path is the relative path to the article.
	Path string

	// Route is the route to the article on the website.
	Route string

	// Content of the article
	Title       string
	Description string
	Author      string
	Pinned      bool
	Created     time.Time
	Updated     time.Time

	// Does the article require authentication to view?
	Authenticator auth.Authenticator

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

// ArticleRoute creates the route to an article from the root filepath and the
// path to the document the article was generated from. The route does not
// include the file extension.
func ArticleRoute(root, filepath string) string {
	return strings.TrimPrefix(strings.TrimSuffix(filepath, path.Ext(filepath)), path.Clean(root))
}

// ArticlePath returns the key used to find an article in the articleMap. This
// is similar to the article's route, however it includes the file extension.
func ArticlePath(root, filepath string) string {
	return strings.TrimPrefix(filepath, path.Clean(root))
}

// GenerateArticle reads a document from the filesystem and generates an
// in-memory article for use by the web-server.
func GenerateArticle(root, filepath string) Article {
	key := ArticlePath(root, filepath)
	route := ArticleRoute(root, filepath)

	article := Article{Path: key, Route: route}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		article.Err = errors.Errorf("failed to load document: %v", err)
		return article
	}

	var doc Document
	doc, article.Err = UnmarshalDocument(bytes)
	if article.Err != nil {
		return article
	}

	// Marshal here rather than use the bytes directly
	article.Text, article.Err = MarshalDocument(doc)
	if article.Err != nil {
		return article
	}

	article.FilePath = filepath
	article.Title = doc.Properties.Value("Title")
	article.Description = doc.Properties.Value("Description")
	article.Author = doc.Properties.Value("Author")

	// Setup authentication for an article
	username := doc.Properties.Value("Username")
	password := doc.Properties.Value("Password")

	if username != "" || password != "" {
		article.Authenticator, article.Err = auth.NewSimple(username, password)
		if article.Err != nil {
			return article
		}
	}

	pin := strings.ToLower(doc.Properties.Value("Pinned"))
	if pin != "" {
		if pin == "true" || pin == "false" {
			article.Pinned, _ = strconv.ParseBool(pin)
		} else {
			article.Err = errors.New("article property 'Pinned' must be true or false")
		}
	}

	article.SetTimestamps(doc.Properties.Value("Created"), doc.Properties.Value("Updated"))

	article.HTML, article.Err = doc.GenerateHTML()

	return article
}
