package content

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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

// scanContent updates the content based on whats found in the directory at
// contentPath.
func scanContent(content *Content, contentPath string) {
	content.Mutex.Lock()
	defer content.Mutex.Unlock()

	articles, err := generateArticles(contentPath)
	if err != nil {
		log.Print(err.Error())
	} else {
		content.Articles = articles
	}
	for _, article := range articles {
		if article.Error == nil {
			log.Printf("✅ %s\n", article.Route)
		} else {
			log.Printf("❌ %s: %s\n", article.Route, article.Error.Error())
		}
	}
}

// IntervalScan scans for content every scanInterval seconds. If scanInterval
// is 0, then a scan is performed every second by default.
func IntervalScan(contentPath string, scanInterval int, done chan bool, wg *sync.WaitGroup) *Content {
	content := &Content{}

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
				scanContent(content, contentPath)
				time.Sleep(time.Duration(scanInterval) * time.Second)
			}
		}
		wg.Done()
	}()

	return content
}
