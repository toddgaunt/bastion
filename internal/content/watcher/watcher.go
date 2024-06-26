package watcher

import (
	"os"
	"path"
	"path/filepath"
	"sort"
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
	// Path specifies where the watcher should watch for articles
	Path string
	// Logger is where logs will be sent.
	Logger log.Logger
	// Details
	Details content.Details

	// Internal state
	mutex      sync.RWMutex
	articleMap map[string]content.Article
}

// Get returns a single article associated with the given key.
func (w *Watcher) Get(key string) (content.Article, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	article, ok := w.articleMap[key+".md"]
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

// GetAll returns all articles generated from documents that are not unlisted.
// Only articles with the given value of pinned are returned.
// TODO: Maybe pass in a string rather than a bool for pinned? That way we can just GetAll("category")?
func (w *Watcher) GetAll(pinned bool) []content.Article {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	list := []content.Article{}
	for _, v := range w.articleMap {
		// Only add pinned articles to the list
		// Don't add unlisted articles
		if v.Pinned == pinned  && !v.Unlisted {
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

// watchArticles walks a directory to find all subdirectories and add them to
// the watcher. Each document found within a subdirectory is used to generate
// an article.
func watchArticles(root string, watcher *fsnotify.Watcher) (map[string]content.Article, error) {
	articles := make(map[string]content.Article)

	err := filepath.Walk(root, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filePath == "." {
			return nil
		}

		filename := path.Base(filePath)

		if strings.HasPrefix(filename, ".") {
			// Skip "hidden" files and directories, since '.' is reserved for built-in routes
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only directories should be watched. Files shouldn't be watched
		// directly. This allows us to detect when files are added or removed.
		if info.IsDir() {
			watcher.Add(filePath)
		} else {
			article := content.GenerateArticle(root, filePath)
			articles[article.Path] = article
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return articles, nil
}

func logStatus(b bool) string {
	if b {
		return "ok"
	}
	return "fail"
}

// Watch starts a goroutine to watch for file updates in the given
// articlesPath. When an update is detected from the filesystem, an event is
// sent to the goroutine to handle any article generation.
func (w *Watcher) Start(done chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.Logger.Printf(log.Fatal, "failed to create new watcher: %v", err)
		return
	}

	articles, err := watchArticles(w.Path, watcher)
	if err != nil {
		w.Logger.Printf(log.Fatal, "failed to watch articles: %v", err)
	}

	for _, article := range articles {
		w.Logger.With(
			"op", "init",
			"route", article.Route,
			"err", article.Err,
		).Print(log.Info, "init")
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

				path := content.ArticlePath(w.Path, event.Name)
				route := content.ArticleRoute(w.Path, event.Name)

				logger := w.Logger.With(
					"op", op.String(),
					"path", path,
					"route", route,
				)

				if op.Has(fsnotify.Remove) || op.Has(fsnotify.Rename) {
					logger.Print(log.Info, "watch")

					w.mutex.Lock()
					delete(w.articleMap, path)
					w.mutex.Unlock()
				}

				if op.Has(fsnotify.Create) || op.Has(fsnotify.Write) {
					info, err := os.Stat(event.Name)
					if err != nil {
						logger.With("err", err.Error()).Print(log.Error, "falied to stat")
						break
					}

					// Only directories should be watched. Files shouldn't be
					isDir := info.IsDir()
					logger = logger.With("dir", isDir)

					// watched directly.
					if isDir {
						logger.Print(log.Info, "watch")
						watcher.Add(event.Name)
						break
					}

					article := content.GenerateArticle(w.Path, event.Name)

					logger.With(
						"err", article.Err,
					).Print(log.Info, "watch")

					w.mutex.Lock()
					w.articleMap[article.Path] = article
					w.mutex.Unlock()
				}
			case err, ok := <-watcher.Errors:
				logger := w.Logger
				if ok {
					logger = logger.With("err", err.Error())
				}
				logger.Print(log.Error, "watcher error")
			case <-done:
				break loop
			}
		}
		wg.Done()
	}()
}
