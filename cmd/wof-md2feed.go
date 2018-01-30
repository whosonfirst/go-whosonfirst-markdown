package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"github.com/whosonfirst/go-whosonfirst-markdown/flags"
	"github.com/whosonfirst/go-whosonfirst-markdown/jekyll"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"github.com/whosonfirst/go-whosonfirst-markdown/render"
	"github.com/whosonfirst/go-whosonfirst-markdown/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"text/template"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func RenderDirectory(ctx context.Context, dir string, opts *render.FeedOptions) error {

	// log.Println("RENDER", dir)

	posts, err := GatherPosts(ctx, dir, opts)

	if err != nil {
		return err
	}

	if len(posts) == 0 {
		return nil
	}

	return RenderPosts(ctx, dir, posts, opts)
}

func GatherPosts(ctx context.Context, root string, opts *render.FeedOptions) ([]*jekyll.FrontMatter, error) {

	mu := new(sync.Mutex)

	lookup := make(map[string]*jekyll.FrontMatter)
	dates := make([]string, 0)

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:

			if info.IsDir()  {
				return nil
			}

			abs_path, err := filepath.Abs(path)

			if err != nil {
				return err
			}

			if filepath.Base(abs_path) != opts.Input {
				return nil
			}

			fm, err := RenderPath(ctx, path, opts)

			if err != nil {
				return err
			}

			if fm == nil {
				return nil
			}

			mu.Lock()
			ymd := fm.Date.Format("20060102")
			dates = append(dates, ymd)
			lookup[ymd] = fm
			mu.Unlock()
		}

		return nil
	}

	c := crawl.NewCrawler(root)
	err := c.Crawl(cb)

	if err != nil {
		return nil, err
	}

	posts := make([]*jekyll.FrontMatter, 0)

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	for _, ymd := range dates {
		posts = append(posts, lookup[ymd])

		if len(posts) == 10 {	// please make me a variable
			break
		}
	}

	return posts, nil
}

func RenderPosts(ctx context.Context, root string, posts []*jekyll.FrontMatter, opts *render.FeedOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:

		type Data struct {
			Posts []*jekyll.FrontMatter
		}

		d := Data{
			Posts: posts,
		}

		var b bytes.Buffer
		wr := bufio.NewWriter(&b)

		// RENDER TEMPLATE HERE
		// err = t.Execute(wr, d)

		if err != nil {
			return err
		}

		wr.Flush()

		fh := nopCloser{ b }

		return utils.WriteFeed(fh, root, opts)
	}
}

func Render(ctx context.Context, path string, opts *render.HTMLOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		return RenderDirectory(ctx, path, opts)
	}

}

func main() {

	var input = flag.String("input", "index.md", "What you expect the input Markdown file to be called")
	// var output = flag.String("output", "index.html", "What you expect the output HTML file to be called")

	var format = flag.String("format", "rss", "...")
	var items = flag.Int("items", 10, "...")

	var templates flags.TemplateFlags
	flag.Var(&templates, "templates", "One or more templates to parse in addition to -header and -footer")

	// var writers flags.WriterFlags
	// flag.Var(&writers, "writer", "One or more writer to output rendered Markdown to. Valid writers are: fs=PATH; null; stdout")

	flag.Parse()

	t, err := templates.Parse()

	if err != nil {
		log.Fatal(err)
	}

	opts := render.DefaultFeedOptions()
	opts.Input = *input
	opts.Format = *format
	opts.Items = *items
	opts.Templates = t

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, path := range flag.Args() {

		err := Render(ctx, path, opts)

		if err != nil {
			log.Println(err)
			cancel()
			break
		}
	}
}
