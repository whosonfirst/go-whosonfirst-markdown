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

func RenderDirectory(ctx context.Context, dir string, opts *render.HTMLOptions) error {

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

func GatherPosts(ctx context.Context, root string, opts *render.HTMLOptions) ([]*jekyll.FrontMatter, error) {

	mu := new(sync.Mutex)

	lookup := make(map[string]*jekyll.FrontMatter)
	dates := make([]string, 0)

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:

			if info.IsDir() && path != root {

				i := filepath.Join(path, opts.Input)
				_, err := os.Stat(i)

				if os.IsNotExist(err) {
					RenderDirectory(ctx, path, opts)
				}

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
	c.CrawlDirectories = true

	err := c.Crawl(cb)

	if err != nil {
		return nil, err
	}

	posts := make([]*jekyll.FrontMatter, 0)

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	for _, ymd := range dates {
		posts = append(posts, lookup[ymd])
	}

	return posts, nil
}

func RenderPosts(ctx context.Context, root string, posts []*jekyll.FrontMatter, opts *render.HTMLOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:

		tm := `{{ range $fm := .Posts }}
### [{{ $fm.Title }}]({{ $fm.Permalink }})

> {{ $fm.Excerpt }}

> _{{ if $fm.Date }}<span class="pubdate"><a href="/blog/{{ $fm.Date.Year }}/{{ $fm.Date.Format "01" }}/">{{ $fm.Date.Format "Jan" }}</a> <a href="/blog/{{ $fm.Date.Year }}/{{ $fm.Date.Format "01" }}/{{ $fm.Date.Day }}">{{ $fm.Date.Day}}</a>, <a href="/blog/{{ $fm.Date.Year }}/">{{ $fm.Date.Format "2006" }}</a></span>{{ end }}_
	    {{ end }}`

		t, err := template.New("index").Parse(tm)

		if err != nil {
			return err
		}

		type Data struct {
			Posts []*jekyll.FrontMatter
		}

		d := Data{
			Posts: posts,
		}

		var b bytes.Buffer
		wr := bufio.NewWriter(&b)

		err = t.Execute(wr, d)

		if err != nil {
			return err
		}

		wr.Flush()

		r := bytes.NewReader(b.Bytes())
		fh := nopCloser{r}

		p_opts := parser.DefaultParseOptions()
		fm, buf, err := parser.Parse(fh, p_opts)

		if err != nil {
			log.Printf("FAILED to parse MD document, because %s\n", err)
			return err
		}

		doc, err := markdown.NewDocument(fm, buf)

		if err != nil {
			log.Printf("FAILED to create MD document, because %s\n", err)
			return err
		}

		html, err := render.RenderHTML(doc, opts)

		if err != nil {
			log.Printf("FAILED to render HTML document, because %s\n", err)
			return err
		}

		return utils.WriteHTML(html, root, opts)
	}
}

func RenderPath(ctx context.Context, path string, opts *render.HTMLOptions) (*jekyll.FrontMatter, error) {

	select {

	case <-ctx.Done():
		return nil, nil
	default:

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Printf("FAILED to render path %s, because %s\n", path, err)
			return nil, err
		}

		if filepath.Base(abs_path) != opts.Input {
			return nil, nil
		}

		parse_opts := parser.DefaultParseOptions()
		parse_opts.Body = false

		fm, _, err := parser.ParseFile(abs_path, parse_opts)

		if err != nil {
			log.Printf("FAILED to parse %s, because %s\n", path, err)
			return nil, err
		}

		return fm, nil
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
	var output = flag.String("output", "index.html", "What you expect the output HTML file to be called")
	var header = flag.String("header", "", "The path to a custom (Go) template to use as header for your HTML output")
	var footer = flag.String("footer", "", "The path to a custom (Go) template to use as a footer for your HTML output")

	var templates flags.TemplateFlags
	flag.Var(&templates, "templates", "One or more templates to parse in addition to -header and -footer")

	// var writers flags.WriterFlags
	// flag.Var(&writers, "writer", "One or more writer to output rendered Markdown to. Valid writers are: fs=PATH; null; stdout")

	flag.Parse()

	t, err := templates.Parse()

	if err != nil {
		log.Fatal(err)
	}

	opts := render.DefaultHTMLOptions()
	opts.Input = *input
	opts.Output = *output
	opts.Header = *header
	opts.Footer = *footer
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
