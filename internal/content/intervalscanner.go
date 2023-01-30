package content

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

	"github.com/toddgaunt/bastion"
)

type IntervalScanner struct {
	ScanInterval int
	Logger       bastion.Logger
	WithDetails  bastion.Details

	mutex      sync.RWMutex
	articleMap map[string]bastion.Article
}

func (m *IntervalScanner) Get(key string) (bastion.Article, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	article, ok := m.articleMap[key]
	return article, ok
}

func (m *IntervalScanner) GetAll(pinned bool) []bastion.Article {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	list := []bastion.Article{}
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

func (m *IntervalScanner) Details() bastion.Details {
	return m.WithDetails
}

// generateArticles walks a directory, and generates articles from
// subdirectories and markdown files found.
func generateArticles(dirpath string) (map[string]bastion.Article, error) {
	articles := make(map[string]bastion.Article)

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

		article := bastion.Article{Route: route}

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

		doc, err := parseDocument(bytes)
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
func (s *IntervalScanner) ScanArticles(articlesPath string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	articles, err := generateArticles(articlesPath)
	if err != nil {
		s.Logger.Error(err.Error())
	} else {
		s.articleMap = articles
	}
	for _, article := range articles {
		if article.Err == nil {
			s.Logger.Infow("scan",
				"status", "ok",
				"route", article.Route,
			)
		} else {
			s.Logger.Infow("scan",
				"status", "fail",
				"route", article.Route,
				"err", article.Err.Error(),
			)
		}
	}
}

// Start starts a goroutine to scan for articles every s.ScanInterval seconds.
// If s.ScanInterval is 0, then a scan is only performed once at startup.
func (s *IntervalScanner) Start(articlesPath string, done chan bool, wg *sync.WaitGroup) {
	if s.ScanInterval == 0 {
		s.Logger.Warn("scan_interval is 0, articles will only be scanned once")
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
				if s.ScanInterval == 0 {
					break loop
				}
				time.Sleep(time.Duration(s.ScanInterval) * time.Second)
			}
		}
		wg.Done()
	}()
}
