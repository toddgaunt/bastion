package watcher

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/errors"
	"github.com/toddgaunt/bastion/internal/log"
)

// Watcher reads documents from the filesystem on whenever it detects they
// have been updated interval and creates articles from them. Each document is
// related to an article with a shared key.
type Watcher struct {
	Logger  log.Logger
	Details content.Details

	mutex      sync.RWMutex
	articleMap map[string]content.Article
}

// Get returns a single article associated with the given key.
func (w *Watcher) Get(key string) (content.Article, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	article, ok := w.articleMap[key]
	if !ok {
		return content.Article{}, errors.New("article does not exist")
	}

	return article, nil
}

// Update modifies the underlying document associated with the given key.
func (w *Watcher) Update(key string, doc content.Document) error {
	article, err := w.Get(key)
	if err != nil {
		return err
	}

	bytes, err := content.MarshalDocument(doc)
	if err != nil {
		return err
	}

	err = os.WriteFile(article.FilePath, bytes, 0755)
	if err != nil {
		return err
	}

	return nil
}

// GetAll returns all articles generated from documents
func (w *Watcher) GetAll(pinned bool) []content.Article {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	list := []content.Article{}
	for _, v := range w.articleMap {
		// Only add pinned articles to the list
		if v.Pinned == pinned {
			list = append(list, v)
		}
	}

	sort.Slice(list, func(i int, j int) bool {
		return list[i].Title < list[j].Title
	})

	sort.Slice(list, func(i int, j int) bool {
		return list[i].Created.After(list[j].Created)
	})

	return list
}

// GetDetails returns the details of the content.
func (w *Watcher) GetDetails() content.Details {
	return w.Details
}

// articleRoute generates the route to be used by the webserver handler from
// the root filepath and the path to the document the article was generated
// from.
func articleRoute(root, filePath string) string {
	return strings.TrimPrefix(strings.TrimSuffix(filePath, path.Ext(filePath)), path.Clean(root))
}

// generateArticle reads a document from the filesystem and generates an
// in-memory article for use by the web-server.
func generateArticle(root, filePath string) content.Article {
	route := articleRoute(root, filePath)

	article := content.Article{Route: route}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		article.Err = errors.Errorf("failed to load document: %v", err)
		return article
	}

	var doc content.Document
	doc, article.Err = content.UnmarshalDocument(bytes)
	if article.Err != nil {
		return article
	}

	// Marshal here rather than use the bytes directly
	article.Text, article.Err = content.MarshalDocument(doc)
	if article.Err != nil {
		return article
	}

	article.FilePath = filePath
	article.Title = doc.Properties.Value("Title")
	article.Description = doc.Properties.Value("Description")
	article.Author = doc.Properties.Value("Author")

	pin := strings.ToLower(doc.Properties.Value("Pinned"))
	if pin != "" {
		if pin == "true" || pin == "false" {
			article.Pinned, _ = strconv.ParseBool(pin)
		} else {
			article.Err = errors.New("article property 'Pinned' must be true or false")
		}
	}

	article.SetTimestamps(doc.Properties.Value("Created"), doc.Properties.Value("Updated"))

	article.HTML, article.Err = doc.GenerateHTML()

	return article
}

// watchArticles walks a directory, and find all directories and adds them to the
// watcher. Each document found is used to generate and article.
func watchArticles(watcher *fsnotify.Watcher, root string) (map[string]content.Article, error) {
	articles := make(map[string]content.Article)

	err := filepath.Walk(root, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := path.Base(filePath)

		if strings.HasPrefix(name, ".") {
			// Skip "hidden" files and directories, since '.' is reserved for built-in routes
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only directories should be watched. Files shouldn't be watched directly.
		if info.IsDir() {
			watcher.Add(filePath)
			return nil
		}

		article := generateArticle(root, filePath)
		articles[article.Route] = article

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate articles: %w", err)
	}

	return articles, nil
}

func logArticleCreateOrUpdate(article content.Article, op string, logger log.Logger) {
	if article.Err == nil {
		logger.With(
			"op", op,
			"status", "ok",
			"route", article.Route,
		).Print(log.Info, "watch")
	} else {
		logger.With(
			"op", op,
			"status", "fail",
			"route", article.Route,
			"err", article.Err,
		).Print(log.Info, "watch")
	}
}

// Watch starts a goroutine to watch for file updates in the given articlesPath. When an update is detected the
func (w *Watcher) Start(articlesPath string, done chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.Logger.Print(log.Fatal, "failed to create new watcher: %v", err)
		return
	}

	articles, err := watchArticles(watcher, articlesPath)
	if err != nil {
		w.Logger.Print(log.Fatal, "failed to watch articles: %v", err)
	}

	for _, article := range articles {
		logArticleCreateOrUpdate(article, "init", w.Logger)
	}

	w.mutex.Lock()
	w.articleMap = articles
	w.mutex.Unlock()

	go func() {
	loop:
		for {
			select {
			case event := <-watcher.Events:
				op := event.Op

				if op.Has(fsnotify.Create) || op.Has(fsnotify.Write) {
					info, err := os.Stat(event.Name)
					if err != nil {
						w.Logger.Printf(log.Error, err.Error())
						break
					}

					// Only directories should be watched. Files shouldn't be
					// watched directly.
					if info.IsDir() {
						watcher.Add(event.Name)
						break
					}

					article := generateArticle(articlesPath, event.Name)
					logArticleCreateOrUpdate(article, op.String(), w.Logger)

					w.mutex.Lock()
					w.articleMap[article.Route] = article
					w.mutex.Unlock()
				}

				if op.Has(fsnotify.Rename) || op.Has(fsnotify.Remove) {
					route := articleRoute(articlesPath, event.Name)

					w.Logger.With(
						"op", "delete",
						"route", route,
					).Print(log.Info, "watch")

					w.mutex.Lock()
					delete(w.articleMap, route)
					w.mutex.Unlock()
				}
			case <-done:
				break loop
			}
		}
		wg.Done()
	}()
}
