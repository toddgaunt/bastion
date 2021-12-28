package router

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"bastionburrow.com/bastion/internal/content"
	"github.com/go-chi/chi"
)

type articleVariables struct {
	Title       string
	Description string
	Site        Config
	HTML        template.HTML
}

const articlesCtxKey = "articleID"

// ArticlesCtx is middleware for a router to provide a clean path to an article
// for an HTTPHandler.
func ArticlesCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := "/" + filepath.Clean(chi.URLParam(r, "*"))
		ctx := context.WithValue(r.Context(), articlesCtxKey, articleID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetArticle returns an HTTP handler function to respond to HTTP requests for
// an article. The handler will write an HTML representation of an article as
// a response, or a problemjson response if the article does not exist or there
// was a problem generating it.
func GetArticle(tmpl *template.Template, config Config, content *content.Content) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		articleID := r.Context().Value(articlesCtxKey).(string)
		log.Print(articleID)

		// The critical section is wrapped within a closure so defer can be
		// used for the mutex operations.
		var vars articleVariables
		var problem = func() *ProblemJSON {
			content.Mutex.RLock()
			defer content.Mutex.RUnlock()

			article, ok := content.Articles[articleID]

			if !ok {
				return &ProblemJSON{Title: "No Such Article", Status: http.StatusNotFound, Detail: fmt.Sprintf("Article %s does not exist", articleID)}
			}

			if article.Error != nil {
				return &ProblemJSON{Title: "Article Generation Error", Status: http.StatusNotImplemented, Detail: article.Error.Error()}
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
