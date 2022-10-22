package router

import (
	"html/template"
	"net/http"
	"path"
	"sync"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/content"
)

// Config contains all configuration for a bastion server's router.
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
		<div class="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Site.Pinned}}
			<a href="/{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div class="content">
			<article>
				<div class="article-header">
					<h1 class="article-title">{{.Site.Name}}</h1>
					<p class="article-description">{{.Site.Description}}</p>
				</div>
				<div class="article-body">
					<ul>
						{{range $k, $v := .SortedIndex}}
						<li><a href="{{$v.Route}}">{{$v.FormattedDate}} - {{$v.Title}}</a></li>
						{{end}}
					</ul>
				</div>
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
		<div class="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Site.Pinned}}
			<a href="/{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div class="content">
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
		<div class="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Site.Pinned}}
			<a href="/{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div class="content">
			<article>
				<div class="problem-header">
					<hr>
					<h1 class="problem-title">{{.Title}}</h1>
					<hr>
				</div>
				<div class="problem-body">
					<p>{{.Description}}</p>
				</div>
			</article>
		</div>
	</body>
</html>`

var (
	indexTemplate   = template.Must(template.New("index").Parse(indexTemplateString))
	articleTemplate = template.Must(template.New("article").Parse(articleTemplateString))
	problemTemplate = template.Must(template.New("problem").Parse(problemTemplateString))
)

// New creates a new router for a bastion website.
func New(prefixDir string, config Config) (chi.Router, error) {
	dir := path.Clean(prefixDir)
	r := chi.NewRouter()

	var done chan bool
	var wg = &sync.WaitGroup{}

	staticFileServer := http.FileServer(http.Dir(dir + "/static"))
	content := content.IntervalScan(dir+"/content", config.ScanInterval, done, wg)

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
