package handlers

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/errors"
)

type contextKey string

const articlesCtxKey = contextKey("articleID")

// ArticlePath is middleware for a router to provide a clean path to an article
// for an HTTPHandler.
func ArticlePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := "/" + filepath.Clean(chi.URLParam(r, "*"))
		ctx := context.WithValue(r.Context(), articlesCtxKey, articleID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Article returns an HTTP handler function to respond to HTTP requests for
// an article. The handler will write an HTML representation of an article as
// a response, or a problemjson response if the article does not exist or there
// was a problem generating it.
func Articles(store content.Store) func(w http.ResponseWriter, r *http.Request) {
	const op = "GetArticle"
	fn := func(w http.ResponseWriter, r *http.Request) error {
		articleID := r.Context().Value(articlesCtxKey).(string)

		// The critical section is wrapped within a closure so defer can be
		// used for the mutex operations.
		var markdown string
		var vars templateVariables
		var getArticle = func(articleKey string) error {
			article, ok := store.Get(articleKey)

			if !ok {
				return errors.Annotation{
					WithOp:     op,
					WithTitle:  "Article Not Found",
					WithStatus: http.StatusNotFound,
					WithDetail: fmt.Sprintf("No article located at %s", articleKey),
				}.Wrap(errors.New("article not in map"))
			}

			if article.Err != nil {
				return errors.Annotation{
					WithOp:     op,
					WithTitle:  "Article Generation Error",
					WithStatus: http.StatusInternalServerError,
				}.Wrap(article.Err)
			}

			markdown = article.Markdown
			vars = templateVariables{
				Title:       article.Title,
				Description: article.Description,
				HTML:        article.HTML,
				content:     store,
			}

			return nil
		}

		if strings.HasSuffix(articleID, ".md") {
			err := getArticle(strings.TrimSuffix(articleID, ".md"))
			if err != nil {
				return err
			}
			w.Header().Add("Content-Type", "text")
			w.Write([]byte(markdown))
		} else {
			err := getArticle(articleID)
			if err != nil {
				return err
			}
			buf := &bytes.Buffer{}
			articleTemplate.Execute(buf, vars)

			w.Header().Add("Content-Type", "text/html")
			w.Write(buf.Bytes())
		}

		return nil
	}

	return wrapper(fn)
}