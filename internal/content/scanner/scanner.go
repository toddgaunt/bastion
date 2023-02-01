package scanner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/log"
)

type Scanner struct {
	Interval int
	Logger   log.Logger
	Details  content.Details

	mutex      sync.RWMutex
	articleMap map[string]content.Article
}

func (m *Scanner) Get(key string) (content.Article, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	article, ok := m.articleMap[key]
	return article, ok
}

func (m *Scanner) GetAll(pinned bool) []content.Article {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	list := []content.Article{}
	for _, v := range m.articleMap {
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

func (m *Scanner) GetDetails() content.Details {
	return m.Details
}

// generateArticles walks a directory, and generates articles from
// subdirectories and markdown files found.
func generateArticles(dirpath string) (map[string]content.Article, error) {
	articles := make(map[string]content.Article)

	err := filepath.Walk(dirpath, func(articlePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := path.Base(articlePath)
		route := strings.TrimPrefix(strings.TrimSuffix(articlePath, path.Ext(articlePath)), path.Clean(dirpath))

		if strings.HasPrefix(name, ".") {
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

		article := content.Article{Route: route}

		// Past this point the article should always be added, even if only partially
		// made, since if there is an error a ProblemJSON will be generated.
		defer func() {
			if article.Err != nil || article.HTML != "" {
				articles[route] = article
			}
		}()

		bytes, err := ioutil.ReadFile(articlePath)
		if err != nil {
			article.Err = fmt.Errorf("article '%s' could not be read from the filesystem", route)
			return nil
		}
		article.Markdown = string(bytes)

		doc, err := content.ParseDocument(bytes)
		if err != nil {
			article.Err = err
			return nil
		}

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

		html, err := doc.GenerateHTML()
		if err != nil {
			article.Err = err
			return nil
		}
		article.HTML = html

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate articles: %w", err)
	}

	return articles, nil
}

// Scan updates the articleMap based on whats found in the directory at
// articlesPath.
func (s *Scanner) ScanArticles(articlesPath string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	articles, err := generateArticles(articlesPath)
	if err != nil {
		s.Logger.Print(log.Error, err.Error())
	} else {
		s.articleMap = articles
	}
	for _, article := range articles {
		if article.Err == nil {
			s.Logger.With(
				"status", "ok",
				"route", article.Route,
			).Print(log.Info, "scan")
		} else {
			s.Logger.With(
				"status", "fail",
				"route", "err",
			).Print(log.Info, "scan")
		}
	}
}

// Start starts a goroutine to scan for articles every s.ScanInterval seconds.
// If s.ScanInterval is 0, then a scan is only performed once at startup.
func (s *Scanner) Start(articlesPath string, done chan bool, wg *sync.WaitGroup) {
	if s.Interval == 0 {
		s.Logger.Print(log.Warn, "scan_interval is 0, articles will only be scanned once")
	}

	wg.Add(1)
	go func() {
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
				s.ScanArticles(articlesPath)
				if s.Interval == 0 {
					break loop
				}
				time.Sleep(time.Duration(s.Interval) * time.Second)
			}
		}
		wg.Done()
	}()
}
