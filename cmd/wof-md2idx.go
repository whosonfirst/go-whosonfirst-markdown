package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"github.com/whosonfirst/go-whosonfirst-markdown/flags"
	"github.com/whosonfirst/go-whosonfirst-markdown/jekyll"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"github.com/whosonfirst/go-whosonfirst-markdown/render"
	"github.com/whosonfirst/go-whosonfirst-markdown/writer"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"text/template"
)

var default_index_list string

func init() {

	default_index_list = `{{ range $fm := .Posts }}
### [{{ $fm.Title }}]({{ $fm.Permalink }}) WHAT

> {{ $fm.Excerpt }}

{{$lena := len $fm.Authors }}
{{$lent := len $fm.Tags }}
<small style="font-style:italic;display:block;margin-bottom:2em;">This is a blog post by
{{ range $ia, $a := $fm.Authors }}{{ if gt $lena 1 }}{{if eq $ia 0}}{{else if eq (plus1 $ia) $lena}} and {{else}}, {{end}}{{ end }}<a href="#" class="hey-look">{{ $a }}</a>{{ end }}
{{ if gt $lent 0 }} that is tagged {{ range $it, $t := .Tags }}{{ if gt $lent 1 }}{{if eq $it 0}}{{else if eq (plus1 $it) $lent}} and {{else}}, {{end}}{{ end }}<a href="#" class="hey-look">{{ $t }}</a>{{ end }}.{{end}}
{{ if $fm.Date }} It was published on <span class="pubdate"><a href="/blog/{{ $fm.Date.Year }}/{{ $fm.Date.Format "01" }}/">{{ $fm.Date.Format "January" }}</a> <a href="/blog/{{ $fm.Date.Year }}/{{ $fm.Date.Format "01" }}/{{ $fm.Date.Format "02" }}">{{ $fm.Date.Format "02"}}</a>, <a href="/blog/{{ $fm.Date.Year }}/">{{ $fm.Date.Format "2006" }}</a></span>.{{ end }}
</small>

	    {{ end }}`
}

// please get rid of this - it's built in to ioutil (20180709/thisisaaronland)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type MarkdownOptions struct {
	MarkdownTemplates *template.Template
	List              string
}

func RenderDirectory(ctx context.Context, dir string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	posts, err := GatherPosts(ctx, dir, html_opts, md_opts)

	if err != nil {
		return err
	}

	if len(posts) == 0 {
		return nil
	}

	return RenderPosts(ctx, dir, posts, html_opts, md_opts)
}

func GatherPosts(ctx context.Context, root string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) ([]*jekyll.FrontMatter, error) {

	mu := new(sync.Mutex)

	lookup := make(map[string]*jekyll.FrontMatter)
	dates := make([]string, 0)

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:

			if info.IsDir() && path != root {

				i := filepath.Join(path, html_opts.Input)
				_, err := os.Stat(i)

				if os.IsNotExist(err) {
					RenderDirectory(ctx, path, html_opts, md_opts)
				}

				return nil
			}

			abs_path, err := filepath.Abs(path)

			if err != nil {
				return err
			}

			if filepath.Base(abs_path) != html_opts.Input {
				return nil
			}

			fm, err := RenderPath(ctx, path, html_opts)

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

func RenderPosts(ctx context.Context, root string, posts []*jekyll.FrontMatter, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:

		t := md_opts.MarkdownTemplates.Lookup(md_opts.List)

		if t == nil {

			tm, err := template.New("list").Parse(default_index_list)

			if err != nil {
				return err
			}

			t = tm
		}

		type Data struct {
			Posts []*jekyll.FrontMatter
		}

		d := Data{
			Posts: posts,
		}

		var b bytes.Buffer
		wr := bufio.NewWriter(&b)

		err := t.Execute(wr, d)

		if err != nil {
			return err
		}

		wr.Flush()

		r := bytes.NewReader(b.Bytes())
		fh := nopCloser{r}

		parse_opts := parser.DefaultParseOptions()
		fm, buf, err := parser.Parse(fh, parse_opts)

		if err != nil {
			log.Printf("FAILED to parse MD document, because %s\n", err)
			return err
		}

		doc, err := markdown.NewDocument(fm, buf)

		if err != nil {
			log.Printf("FAILED to create MD document, because %s\n", err)
			return err
		}

		html, err := render.RenderHTML(doc, html_opts)

		if err != nil {
			log.Printf("FAILED to render HTML document, because %s\n", err)
			return err
		}

		w := ctx.Value("writer").(writer.Writer)

		if w == nil {
			return errors.New("Can't load writer from context")
		}

		out_path := filepath.Join(root, html_opts.Output)
		return w.Write(out_path, html)
	}
}

// THIS IS A BAD NAME - ALSO SHOULD BE SHARED CODE...
// (20180130/thisisaaronland)

func RenderPath(ctx context.Context, path string, html_opts *render.HTMLOptions) (*jekyll.FrontMatter, error) {

	select {

	case <-ctx.Done():
		return nil, nil
	default:

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Printf("FAILED to render path %s, because %s\n", path, err)
			return nil, err
		}

		if filepath.Base(abs_path) != html_opts.Input {
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

func Render(ctx context.Context, path string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		return RenderDirectory(ctx, path, html_opts, md_opts)
	}

}

func main() {

	var input = flag.String("input", "index.md", "What you expect the input Markdown file to be called")
	var output = flag.String("output", "index.html", "What you expect the output HTML file to be called")
	var header = flag.String("header", "", "The name of the (Go) template to use as a custom header")
	var footer = flag.String("footer", "", "The name of the (Go) template to use as a custom footer")
	var list = flag.String("list", "", "The name of the (Go) template to use as a custom list view")

	var templates flags.HTMLTemplateFlags
	flag.Var(&templates, "templates", "One or more directories containing (Go) templates to parse")

	var md_templates flags.MarkdownTemplateFlags
	flag.Var(&md_templates, "markdown-templates", "One or more directories containing (Go) Markdown templates to parse")

	var writers flags.WriterFlags
	flag.Var(&writers, "writer", "One or more writer to output rendered Markdown to. Valid writers are: fs=PATH; null; stdout")

	flag.Parse()

	wr, err := writers.ToWriter()

	if err != nil {
		log.Fatal(err)
	}

	t, err := templates.Parse()

	if err != nil {
		log.Fatal(err)
	}

	html_opts := render.DefaultHTMLOptions()
	html_opts.Input = *input
	html_opts.Output = *output
	html_opts.Header = *header
	html_opts.Footer = *footer
	html_opts.Templates = t

	markdown_t, err := md_templates.Parse()

	if err != nil {
		log.Fatal(err)
	}

	md_opts := &MarkdownOptions{
		MarkdownTemplates: markdown_t,
		List:              *list,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "writer", wr)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, path := range flag.Args() {

		err := Render(ctx, path, html_opts, md_opts)

		if err != nil {
			log.Println(err)
			cancel()
			break
		}
	}
}
