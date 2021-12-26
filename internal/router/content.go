package router

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bastionburrow.com/bastion/internal/document"
	"github.com/go-chi/chi"
)

type articleVariables struct {
	Title       string
	Description string
	Site        Config
	HTML        template.HTML
}

type Article struct {
	Route       string
	Title       string
	Description string
	Author      string
	Created     time.Time
	Updated     time.Time

	HTML    template.HTML
	Problem *ProblemJSON
}

type Content struct {
	mutex    sync.RWMutex
	Articles map[string]*Article
}

func (article Article) FormattedDate() string {
	return article.Created.Format("2006-01-02")
}

const articlesCtxKey = "articleID"
const articleGenerationProblemTitle = "article generation error"

// SetTimestamps will set the provided timestamp strings, if not empty, in the
// article. If they fail to be parsed as timestamps, the article's problem
// field is set with a relevant ProblemJSON object.
func (a *Article) SetTimestamps(created string, updated string) {
	if created != "" {
		t, err := time.Parse("2006-01-02", created)
		if err != nil {
			a.Problem = &ProblemJSON{Title: articleGenerationProblemTitle, Status: http.StatusInternalServerError, Detail: err.Error()}
		}
		a.Created = t
	}
	if updated != "" {
		t, err := time.Parse("2006-01-02", updated)
		if err != nil {
			a.Problem = &ProblemJSON{Title: articleGenerationProblemTitle, Status: http.StatusInternalServerError, Detail: err.Error()}
		}
		a.Updated = t
	}
}

// ArticlesCtx is middleware for a router to provide a clean path to an article
// for an HTTPHandler
func ArticlesCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := "/"+filepath.Clean(chi.URLParam(r, "*"))
		ctx := context.WithValue(r.Context(), articlesCtxKey, articleID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetArticle returns an HTTP handler function to respond to HTTP requests for
// an article. The handler will write an HTML representation of an article as
// a response, or a problemjson response if the article does not exist or there
// was a problem generating it.
func GetArticle(tmpl *template.Template, config Config, content *Content) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		articleID := r.Context().Value(articlesCtxKey).(string)
		log.Print(articleID)

		// The critical section is wrapped within a closure so defer can be
		// used for the mutex operations.
		var vars articleVariables
		var problem = func() *ProblemJSON {
			content.mutex.RLock()
			defer content.mutex.RUnlock()

			article, ok := content.Articles[articleID]

			if !ok {
				return &ProblemJSON{Title: "No Such Article", Status: http.StatusNotFound, Detail: fmt.Sprintf("Article %s does not exist", articleID)}
			}

			if article.Problem != nil {
				return article.Problem
			}

			vars = articleVariables{
				Title:       article.Title,
				Description: article.Description,
				Site:        config,
				HTML:        article.HTML,
			}

			return nil
		}()

		if problem != nil {
			return problem
		}

		buf := &bytes.Buffer{}
		tmpl.Execute(buf, vars)

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}

// GenerateArticles walks a directory, and generates articles from
// subdirectories and markdown files found
func GenerateArticles(contentPath string) (map[string]*Article, error) {
	articles := make(map[string]*Article)

	filepath.Walk(contentPath, func(articlePath string, info os.FileInfo, err error) error {
		articleID := articlePath[len(contentPath):len(articlePath)-len(path.Ext(articlePath))]
		articleName := path.Base(articlePath)

		if strings.HasPrefix(articleName, ".") {
			// Skip "hidden" files and directories, since '.' is reserved for built-in routes
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			// Don't do anything for directories
			return nil
		}

		article := &Article{Route: articleID}

		// Past this point the article should always be added, even if only partially
		// made, since if there is an error a ProblemJSON will be generated.
		defer func() {
			if article.Problem != nil || article.HTML != "" {
				articles[articleID] = article
			}
		}()

		data, err := ioutil.ReadFile(articlePath)
		if err != nil {
			article.Problem = &ProblemJSON{Title: articleGenerationProblemTitle, Status: http.StatusNotFound, Detail: fmt.Sprintf("Article '%s' could not be read from the filesystem", articleID)}
			return nil
		}

		doc, err := document.Parse(data)
		if err != nil {
			article.Problem = &ProblemJSON{Title: articleGenerationProblemTitle, Status: http.StatusInternalServerError, Detail: err.Error()}
			return nil
		}

		article.Title = doc.Properties.Value("Title")
		article.Description = doc.Properties.Value("Description")
		article.Author = doc.Properties.Value("Author")
		article.SetTimestamps(doc.Properties.Value("Created"), doc.Properties.Value("Updated"))

		bytes, err := doc.GenerateHTML()
		if err != nil {
			article.Problem = &ProblemJSON{Title: articleGenerationProblemTitle, Status: http.StatusInternalServerError, Detail: err.Error()}
			return nil
		}

		article.HTML = template.HTML(bytes)

		return nil
	})

	return articles, nil
}

// ScanContent scans for articles for a given configuration
func ScanContent(contentPath string, scanInterval int) *Content {
	content := &Content{}

	go func() {
		for {
			log.Print("🔍 scanning content")
			//NOTE: critical section begin
			content.mutex.Lock()
			articles, err := GenerateArticles(contentPath)
			if err != nil {
				log.Print(err.Error())
			} else {
				content.Articles = articles
			}
			for _, article := range articles {
				if article.Problem == nil {
					log.Printf("✅ %s\n", article.Route)
				} else {
					log.Printf("❌ %s: %s\n", article.Route, article.Problem.Detail)
				}
			}
			content.mutex.Unlock()
			//NOTE: critical section end
			time.Sleep(time.Duration(scanInterval) * time.Second)
		}
	}()

	return content
}
