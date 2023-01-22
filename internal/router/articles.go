package router

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/articles"
	"github.com/toddgaunt/bastion/internal/httpjson"
)

const articlesCtxKey = contextKey("articleID")

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
func GetArticle(tmpl *template.Template, config Config, articleMap *articles.ArticleMap) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *httpjson.Problem {
		articleID := r.Context().Value(articlesCtxKey).(string)

		// The critical section is wrapped within a closure so defer can be
		// used for the mutex operations.
		var markdown string
		var vars templateVariables
		var getArticle = func(articleKey string) *httpjson.Problem {
			articleMap.Mutex.RLock()
			defer articleMap.Mutex.RUnlock()

			article, ok := articleMap.Articles[articleKey]

			if !ok {
				return &httpjson.Problem{
					Title:  "No Such Article",
					Status: http.StatusNotFound,
					Detail: fmt.Sprintf("Article %s does not exist", articleKey),
				}
			}

			if article.Error != nil {
				return &httpjson.Problem{
					Title:  "Article Generation Error",
					Status: http.StatusNotImplemented,
					Detail: article.Error.Error(),
				}
			}

			markdown = article.Markdown
			vars = templateVariables{
				Title:       article.Title,
				Description: article.Description,
				Site:        config,
				ArticleMap:  articleMap,
				HTML:        article.HTML,
			}

			return nil
		}

		if strings.HasSuffix(articleID, ".md") {
			problem := getArticle(strings.TrimSuffix(articleID, ".md"))
			if problem != nil {
				return problem
			}
			w.Header().Add("Content-Type", "text")
			w.Write([]byte(markdown))
		} else {
			problem := getArticle(articleID)
			if problem != nil {
				return problem
			}
			buf := &bytes.Buffer{}
			tmpl.Execute(buf, vars)

			w.Header().Add("Content-Type", "text/html")
			w.Write(buf.Bytes())
		}

		return nil
	}

	return ProblemHandler(f)
}
