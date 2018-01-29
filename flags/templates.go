package flags

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"html/template"
	_ "log"
	"os"
	"path/filepath"
	"sync"
)

type TemplateFlags []string

func (t *TemplateFlags) String() string {
	return fmt.Sprintf("%v", *t)
}

func (t *TemplateFlags) Set(root string) error {

	mu := new(sync.Mutex)

	cb := func(path string, info os.FileInfo) error {

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)

		if ext != ".html" {
			return nil
		}

		mu.Lock()
		*t = append(*t, path)
		mu.Unlock()

		return nil
	}

	c := crawl.NewCrawler(root)
	return c.Crawl(cb)
}

func (t *TemplateFlags) Parse() error {

	if len(*t) == 0 {
		return nil
	}

	_, err := template.ParseFiles(*t...)
	return err
}
