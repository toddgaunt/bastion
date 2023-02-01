package content

import (
	"bufio"
	"bytes"
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

// Document is a structured represtation of the file format for articles.
type Document struct {
	Properties Properties
	Format     string
	Content    []byte
}

// ParseDocument parses bytes and returns a Document, or an error if the bytes did not
// form a valid document representation
func ParseDocument(data []byte) (Document, error) {
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

// Properties is a key value store of document Properties
type Properties map[string][]string

// Add adds a key and value to a property
func (p Properties) Add(key, value string) {
	values, ok := p[strings.ToLower(key)]
	if !ok {
		p[key] = []string{value}
	}

	p[key] = append(values, value)
}

// Value returns the first value associated with a key
func (p Properties) Value(key string) string {
	values, ok := p[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return values[0]
}

// Values returns all of the values associated with a key
func (p Properties) Values(key string) []string {
	values, ok := p[strings.ToLower(key)]
	if !ok {
		return nil
	}
	return values
}

func parseProperties(data []byte) (Properties, error) {
	properties := make(Properties)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		// Ignore blank lines
		if strings.TrimSpace(text) == "" {
			continue
		}

		// KEY : VALUE syntax is expected on non blank lines
		split := strings.SplitN(text, ":", 2)
		if len(split) != 2 {
			return nil, errors.New("expected 'KEY : VALUE' pair")
		}

		key := strings.ToLower(strings.TrimSpace(split[0]))
		value := strings.TrimSpace(split[1])

		properties.Add(key, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return properties, nil
}