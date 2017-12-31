package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	_ "github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"github.com/whosonfirst/go-whosonfirst-markdown/render"
	"github.com/whosonfirst/go-whosonfirst-markdown/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func RenderDirectory(ctx context.Context, path string, opts *render.HTMLOptions) error {

	lookup := make(map[string]*parser.FrontMatter)
	dates := make([]string, 0)

	mu := new(sync.Mutex)

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:

			if info.IsDir() {

				f := func(p string, i os.FileInfo, e error) error {

					if e != nil {
						return e
					}

					if !i.IsDir() {
						return nil
					}

					if p == path {
						return nil
					}

					idx := filepath.Join(p, "index.md")
					info, err := os.Stat(idx)

					if err != nil && !os.IsNotExist(err) {
						return err
					}

					if info != nil {
						return nil
					}

					return RenderDirectory(ctx, p, opts)
				}

				return filepath.Walk(path, f)
			}

			fm, err := RenderPath(ctx, path, opts)

			if err != nil {
				return err
			}

			if fm == nil {
				return nil
			}

			dt := fm.Date
			t, err := time.Parse("January 02, 2006", dt)

			if err != nil {
				return err
			}

			ymd := t.Format("20060102")

			mu.Lock()
			dates = append(dates, ymd)
			lookup[ymd] = fm
			mu.Unlock()

			return nil
		}
	}

	c := crawl.NewCrawler(path)
	c.CrawlDirectories = true

	err := c.Crawl(cb)

	if err != nil {
		return nil
	}

	posts := make([]*parser.FrontMatter, 0)
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	for _, ymd := range dates {
		posts = append(posts, lookup[ymd])
	}

	if len(posts) == 0 {
		return nil
	}

	return RenderPosts(ctx, path, posts, opts)
}

func RenderPosts(ctx context.Context, path string, posts []*parser.FrontMatter, opts *render.HTMLOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:

		tm := `{{ range $fm := .Posts }}
	    * **[{{ $fm.Title }}]({{ $fm.URI }})** {{ $fm.Date }}

	    _{{ $fm.Excerpt }}_
	    
	    {{ end }}
	    `

		t, err := template.New("index").Parse(tm)

		if err != nil {
			return err
		}

		type Data struct {
			Posts []*parser.FrontMatter
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

		r := bytes.NewReader(b.Bytes())
		fh := nopCloser{r}

		p, err := parser.ParseMarkdown(fh)

		if err != nil {
			return err
		}

		o, err := render.RenderHTML(p, opts)

		if err != nil {
			return err
		}

		io.Copy(os.Stdout, o)
		return nil
	}
}

func RenderPath(ctx context.Context, path string, opts *render.HTMLOptions) (*parser.FrontMatter, error) {

	select {

	case <-ctx.Done():
		return nil, nil
	default:

		abs_path, err := filepath.Abs(path)

		if err != nil {
			return nil, err
		}

		fname := filepath.Base(abs_path)

		if fname != opts.Input {
			return nil, nil
		}

		in, err := os.Open(abs_path)

		if err != nil {
			return nil, err
		}

		defer in.Close()

		root := filepath.Dir(abs_path)

		parts := strings.Split(root, "/")
		count := len(parts)

		yyyy := parts[(count-1)-3]
		mm := parts[(count-1)-2]
		dd := parts[(count-1)-1]
		post := parts[(count - 1)]

		t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", yyyy, mm, dd))

		if err != nil {
			return nil, err
		}

		dt := t.Format("January 02, 2006")
		uri := fmt.Sprintf("/blog/%s/%s/%s/%s/", yyyy, mm, dd, post)

		parsed, err := parser.ParseMarkdown(in)

		if err != nil {
			return nil, err
		}

		parsed.FrontMatter.Date = dt
		parsed.FrontMatter.URI = uri

		return parsed.FrontMatter, nil
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

	flag.Parse()

	opts := render.DefaultHTMLOptions()
	opts.Input = *input
	opts.Output = *output

	if *header != "" {

		t, err := utils.LoadTemplate(*header, "header")

		if err != nil {
			log.Fatal(err)
		}

		opts.Header = t
	}

	if *footer != "" {

		t, err := utils.LoadTemplate(*footer, "footer")

		if err != nil {
			log.Fatal(err)
		}

		opts.Footer = t
	}

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
