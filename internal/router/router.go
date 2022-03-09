package router

import (
	"html/template"
	"net/http"
	"sync"

	"bastionburrow.com/bastion/internal/content"
	"github.com/go-chi/chi"
)

// Config contains all configuration for a Monastery website's index, article,
// and problem pages. Some options are: website title, css style, and which
// articles are pinned to the navigation bar rather than indexed.
type Config struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Style        string            `json:"style"`
	Pinned       map[string]string `json:"pinned"`
	ScanInterval int               `json:"scan_interval"`
}

const indexTemplateString = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.Title}}</title>
		<meta name="description" content="{{.Description}}">
		<link href="/.static/styles/{{.Site.Style}}.css" type="text/css" rel="stylesheet">
	</head>
	<body>
		<div id="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Site.Pinned}}
			<a href="/{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div id="content">
			<article>
				<hr>
				<h1 id="article-title">{{.Site.Name}}</h1>
				<p id="article-description">{{.Site.Description}}</p>
				<hr>
				<ul>
					{{range $k, $v := .SortedIndex}}
					<li><a href="{{$v.Route}}">{{$v.FormattedDate}} - {{$v.Title}}</a></li>
					{{end}}
				</ul>
			</article>
		</div>
	</body>
</html>`

const articleTemplateString = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.Title}}</title>
		<meta name="description" content="{{.Description}}">
		<link href="/.static/styles/{{.Site.Style}}.css" type="text/css" rel="stylesheet">
	</head>
	<body>
		<div id="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Site.Pinned}}
			<a href="/{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div id="content">
			{{.HTML}}
		</div>
	</body>
</html>`

const problemTemplateString = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.Title}}</title>
		<meta name="description" content="{{.Description}}">
		<link href="/.static/styles/{{.Site.Style}}.css" type="text/css" rel="stylesheet">
	</head>
	<body>
		<div id="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Site.Pinned}}
			<a href="/{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div id="content">
			<article>
				<hr>
				<h1 id="problem-title">{{.Title}}</h1>
				<hr>
				<p>{{.Description}}</p>
			</article>
		</div>
	</body>
</html>`

// New creates a new router for a bastion website.
func New(prefixDir string, config Config) (chi.Router, error) {
	r := chi.NewRouter()

	indexTemplate := template.Must(template.New("index").Parse(indexTemplateString))
	articleTemplate := template.Must(template.New("article").Parse(articleTemplateString))
	problemTemplate := template.Must(template.New("problem").Parse(problemTemplateString))

	var done chan bool
	var wg sync.WaitGroup

	staticFileServer := http.FileServer(http.Dir(prefixDir + "/static"))
	content := content.IntervalScan(prefixDir+"/content", config.ScanInterval, done, wg)

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetIndex(indexTemplate, config, content))
		r.With(ArticlesCtx).Get("/*", GetArticle(articleTemplate, config, content))
	})

	r.Route("/"+ProblemPath, func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(ProblemsCtx)
			r.Get("/", GetProblem(problemTemplate, config))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	// Closing this channel signals all worker threads to stop and cleanup.
	//close(done)
	//wg.Wait()

	return r, nil
}
