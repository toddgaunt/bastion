package router

import (
	_ "embed"
	"html/template"

	"github.com/toddgaunt/bastion"
)

type templateVariables struct {
	Title       string
	Description string
	HTML        template.HTML
	content     bastion.ContentStore
}

func (vars templateVariables) Details() bastion.Details {
	return vars.content.Details()
}

func (vars templateVariables) Pinned() []bastion.Article {
	return vars.content.GetAll(true)
}

func (vars templateVariables) Index() []bastion.Article {
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
