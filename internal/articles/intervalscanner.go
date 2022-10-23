package articles

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// generateArticles walks a directory, and generates articles from
// subdirectories and markdown files found.
func generateArticles(contentPath string) (map[string]*Article, error) {
	articles := make(map[string]*Article)

	filepath.Walk(contentPath, func(articlePath string, info os.FileInfo, err error) error {
		articleName := path.Base(articlePath)
		articleRoute := strings.TrimPrefix(strings.TrimSuffix(articlePath, path.Ext(articlePath)), path.Clean(contentPath))

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

		article := &Article{Route: articleRoute}

		// Past this point the article should always be added, even if only partially
		// made, since if there is an error a ProblemJSON will be generated.
		defer func() {
			if article.Error != nil || article.HTML != "" {
				articles[articleRoute] = article
			}
		}()

		bytes, err := ioutil.ReadFile(articlePath)
		if err != nil {
			article.Error = fmt.Errorf("Article '%s' could not be read from the filesystem", articleRoute)
			return nil
		}
		article.Markdown = string(bytes)

		doc, err := parseDocument(bytes)
		if err != nil {
			article.Error = err
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
				article.Error = errors.New("Pinned must be true or false")
			}
		}

		article.SetTimestamps(doc.Properties.Value("Created"), doc.Properties.Value("Updated"))

		html, err := doc.GenerateHTML()
		if err != nil {
			article.Error = err
			return nil
		}
		article.HTML = html

		return nil
	})

	return articles, nil
}

// scanArticles updates the articleMap based on whats found in the directory at
// articlesPath.
func scanArticles(articleMap *ArticleMap, articlesPath string) {
	articleMap.Mutex.Lock()
	defer articleMap.Mutex.Unlock()

	articles, err := generateArticles(articlesPath)
	if err != nil {
		log.Print(err.Error())
	} else {
		articleMap.Articles = articles
	}
	for _, article := range articles {
		if article.Error == nil {
			log.Printf("✅ %s\n", article.Route)
		} else {
			log.Printf("❌ %s: %s\n", article.Route, article.Error.Error())
		}
	}
}

// IntervalScan scans for articles every scanInterval seconds. If scanInterval
// is 0, then a scan is performed every second by default.
func IntervalScan(articlesPath string, scanInterval int, done chan bool, wg *sync.WaitGroup) *ArticleMap {
	articleMap := &ArticleMap{}

	if scanInterval == 0 {
		scanInterval = 1
	}

	wg.Add(1)
	go func() {
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
				log.Print("🔍 scanning content")
				scanArticles(articleMap, articlesPath)
				time.Sleep(time.Duration(scanInterval) * time.Second)
			}
		}
		wg.Done()
	}()

	return articleMap
}
