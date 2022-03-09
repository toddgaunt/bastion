package document

import (
	"errors"
	"html/template"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
)

const headerHTML = `<article>
<div class="article-header">
{{if .Title}}<h1 class="article-title">{{.Title}}</h1>
{{end}}{{if .Description}}<p class="article-description">{{.Description}}</p>
{{end}}</div>
<div class="article-body">`

const footerHTML = `</div>
</article>`

var headerTemplate = template.Must(template.New("header").Parse(headerHTML))
var textTemplate = template.Must(template.New("text").Parse(`<pre>{{.}}</pre>`))

type Document struct {
	Properties Properties
	Format     string
	Content    []byte
}

// Parse parses bytes and returns a Document, or an error if the bytes did not
// form a valid document representation
func Parse(data []byte) (Document, error) {
	re := regexp.MustCompile(`===.*===`)
	index := re.FindIndex(data)
	if index == nil {
		return Document{}, errors.New("document does not have article delimiter")
	}

	properties, err := parseProperties(data[:index[0]])
	if err != nil {
		return Document{}, err
	}
	format := strings.TrimSpace(string(data[index[0]+3 : index[1]-3]))
	content := data[index[1]:]

	return Document{
		Properties: properties,
		Format:     format,
		Content:    content,
	}, nil
}

// GenerateHTML generates HTML from a given document
func (doc *Document) GenerateHTML() (template.HTML, error) {
	type variables struct {
		Title       string
		Author      string
		Description string
		Created     string
	}

	vars := variables{
		Title:       doc.Properties.Value("Title"),
		Description: doc.Properties.Value("Description"),
		Author:      doc.Properties.Value("Author"),
		Created:     doc.Properties.Value("Created"),
	}

	buf := &strings.Builder{}

	headerTemplate.Execute(buf, vars)
	switch strings.ToLower(doc.Format) {
	case "text":
		// An HTML template ensures the text content doesn't escape <pre> tags
		if err := textTemplate.Execute(buf, string(doc.Content)); err != nil {
			return "", err
		}
	case "html":
		_, _ = buf.Write(doc.Content)
	case "markdown":
		_, _ = buf.Write(markdown.ToHTML(doc.Content, nil, nil))
	}
	buf.WriteString(footerHTML)

	return template.HTML(buf.String()), nil
}
