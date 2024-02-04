package handlers

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/errors"
	"github.com/toddgaunt/bastion/internal/log"
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

// GetArticle returns an HTTP handler function to respond to HTTP requests for
// an article. The handler will write an HTML representation of an article as
// a response, or a problemjson response if the article does not exist or there
// was a problem generating it.
func (env Env) GetArticle(w http.ResponseWriter, r *http.Request) {
	const op = "Get"
	fn := func(w http.ResponseWriter, r *http.Request) errors.Problem {
		articleID := r.Context().Value(articlesCtxKey).(string)

		var markdown []byte
		var vars templateVariables
		var getArticle = func(articleKey string) errors.Problem {
			article, err := env.Store.Get(articleKey)

			if err != nil {
				return errors.Note{
					Op:         op,
					Title:      "Article Not Found",
					StatusCode: http.StatusNotFound,
					Detail:     fmt.Sprintf("No article located at %s", articleKey),
				}.Wrap(errors.New("article not in map"))
			}

			if article.Err != nil {
				return errors.Note{
					Op:         op,
					Title:      "Article Generation Error",
					StatusCode: http.StatusInternalServerError,
				}.Wrap(article.Err)
			}

			markdown = article.Text
			vars = templateVariables{
				Title:       article.Title,
				Description: article.Description,
				HTML:        article.HTML,
				content:     env.Store,
			}

			return nil
		}

		if strings.HasSuffix(articleID, ".md") {
			err := getArticle(strings.TrimSuffix(articleID, ".md"))
			if err != nil {
				return err
			}
			w.Header().Add("Content-Type", "text")
			w.Write(markdown)
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

	err := fn(w, r)
	handleError(w, err, env.Logger)
}

// UpdateDocument returns an HTTP handler function to respond to HTTP requests
// to update an article. The handler will update the underlying representation
// of an article and reply with a 200 OK, or problemjson response if the
// article does not exist or there was a problem updating it.
func (env Env) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	const op = "Update"
	fn := func(w http.ResponseWriter, r *http.Request) errors.Problem {
		articleID := r.Context().Value(articlesCtxKey).(string)
		articleID = strings.TrimSuffix(articleID, ".md")

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Note{
				StatusCode: http.StatusInternalServerError,
				Detail:     "failed to read request",
			}.Wrap(err)
		}

		doc, err := content.UnmarshalDocument(bytes)
		if err != nil {
			return errors.Note{
				StatusCode: http.StatusBadRequest,
				Detail:     "failed to parse document",
			}.Wrap(err)
		}

		err = env.Store.Update(articleID, doc)
		if err != nil {
			errors.Note{
				StatusCode: http.StatusInternalServerError,
				Detail:     "failed to update document",
			}.Wrap(err)
		}

		env.Logger.With("articleID", articleID).Print(log.Info, "Updated Document")

		w.WriteHeader(http.StatusOK)

		return nil
	}

	err := fn(w, r)
	handleError(w, err, env.Logger)
}
