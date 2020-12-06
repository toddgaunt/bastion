// Copyright 2020, Todd Gaunt <toddgaunt@protonmail.com>
//
// This file is part of Monastery.
//
// Monastery is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Monastery is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Monastery.  If not, see <https://www.gnu.org/licenses/>.

package document

import (
	"html/template"
	"strings"

	"github.com/gomarkdown/markdown"
)

type variables struct {
	Title       string
	Author      string
	Description string
	Created     string
}

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

// GenerateHTML generates HTML from a given document
func GenerateHTML(doc Document) ([]byte, error) {
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
