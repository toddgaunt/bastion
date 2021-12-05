package document

import (
	"errors"
	"html/template"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
)

const headerHTML = `<article>
<hr>
{{if .Title}}<h1 id="article-title">{{.Title}}</h1>
{{end}}{{if .Description}}<p id="article-description">{{.Description}}</p>
{{end}}<hr>
`

const footerHTML = `</article>
`

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
func (doc *Document) GenerateHTML() ([]byte, error) {
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
		textTemplate.Execute(buf, string(doc.Content))
	case "html":
		buf.Write(doc.Content)
		buf.WriteString(footerHTML)
	case "markdown":
		buf.Write(markdown.ToHTML(doc.Content, nil, nil))
	}
	buf.WriteString(footerHTML)

	return []byte(buf.String()), nil
}
