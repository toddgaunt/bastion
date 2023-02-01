package handlers

import (
	_ "embed"
	"html/template"

	"github.com/toddgaunt/bastion/internal/content"
)

type templateVariables struct {
	Title       string
	Description string
	HTML        template.HTML
	content     content.Store
}

func (vars templateVariables) Details() content.Details {
	return vars.content.GetDetails()
}

func (vars templateVariables) Pinned() []content.Article {
	return vars.content.GetAll(true)
}

func (vars templateVariables) Index() []content.Article {
	return vars.content.GetAll(false)
}

var (
	//go:embed problems.html
	problemTemplateString string
	//go:embed index.html
	indexTemplateString string
	//go:embed articles.html
	articleTemplateString string
)

var (
	indexTemplate   = template.Must(template.New("index").Parse(indexTemplateString))
	articleTemplate = template.Must(template.New("article").Parse(articleTemplateString))
	problemTemplate = template.Must(template.New("problem").Parse(problemTemplateString))
)
