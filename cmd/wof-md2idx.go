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
	"github.com/whosonfirst/go-whosonfirst-markdown/uri"
	"github.com/whosonfirst/go-whosonfirst-markdown/writer"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
	_ "unicode"
)

var re_ymd *regexp.Regexp

var default_index_list string
var default_index_rollup string

func init() {

	re_ymd = regexp.MustCompile(".*(\\d{4})(?:/(\\d{2}))?(?:/(\\d{2}))?$")

	default_index_rollup = `{{ range $w := .Rollup}}
* [ {{ $w }} ]( {{ prune_string $w }} )
{{ end }}`

	default_index_list = `{{ range $fm := .Posts }}
### [{{ $fm.Title }}]({{ $fm.Permalink }}) 

> {{ $fm.Excerpt }}

{{$lena := len $fm.Authors }}
{{$lent := len $fm.Tags }}
<small class="this-is">This is a blog post by
    {{ range $ia, $a := $fm.Authors }}{{ if gt $lena 1 }}{{if eq $ia 0}}{{else if eq (plus1 $ia) $lena}} and {{else}}, {{end}}{{ end }}[{{ $a }}](/blog/authors/{{ prune_string $a  }}){{ end }}.
    {{ if $fm.Date }}It was published on <span class="pubdate"><a href="/blog/{{ $fm.Date.Year }}/{{ $fm.Date.Format "01" }}/">{{ $fm.Date.Format "January" }}</a> <a href="/blog/{{ $fm.Date.Year }}/{{ $fm.Date.Format "01" }}/{{ $fm.Date.Format "02" }}/">{{ $fm.Date.Format "02"}}</a>, <a href="/blog/{{ $fm.Date.Year }}/">{{ $fm.Date.Format "2006" }}</a></span>{{ if gt $lent 0 }} and tagged {{ range $it, $t := $fm.Tags }}{{ if gt $lent 1 }}{{if eq $it 0}}{{else if eq (plus1 $it) $lent}} and {{else}}, {{end}}{{ end }}[{{ $t }}](/blog/tags/{{ prune_string $t  }}){{ end }}{{ end}}.
    {{ else }}
    It was tagged {{ range $it, $t := $fm.Tags }}{{ if gt $lent 1 }}{{if eq $it 0}}{{else if eq (plus1 $it) $lent}} and {{else}}, {{end}}{{ end }}[{{ $t }}](/blog/tags/{{ prune_string $t  }}){{ end }}.
    {{ end }}
</small>
{{ end }}
{{ end }}`
}

type MarkdownOptions struct {
	MarkdownTemplates *template.Template
	List              string
	Rollup            string
	Mode              string
}

func RenderDirectory(ctx context.Context, dir string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	lookup, err := GatherPosts(ctx, dir, html_opts, md_opts)

	if err != nil {
		return err
	}

	keys := make([]string, 0)

	for k, _ := range lookup {
		keys = append(keys, k)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	if md_opts.Mode == "date" {

		posts := make([]*jekyll.FrontMatter, 0)

		for _, k := range keys {

			for _, p := range lookup[k] {
				posts = append(posts, p)
			}
		}

		if len(posts) == 0 {
			return nil
		}

		title := "" // where is date...

		return RenderPosts(ctx, dir, title, posts, html_opts, md_opts)
	}

	switch md_opts.Mode {
	case "authors":
		// pass
	case "tags":
		// pass
	default:
		return errors.New("Invalid or unsupported mode")
	}

	root := filepath.Join(dir, md_opts.Mode)

	for _, raw := range keys {

		clean, err := uri.PruneString(raw)

		if err != nil {
			return err
		}

		if clean == "" {
			continue
		}

		// html_opts.Title = raw

		k_dir := filepath.Join(root, clean)

		title := raw
		posts := lookup[raw]

		err = RenderPosts(ctx, k_dir, title, posts, html_opts, md_opts)

		if err != nil {
			return err
		}
	}

	return RenderRollup(ctx, root, keys, html_opts, md_opts)
}

func GatherPosts(ctx context.Context, root string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) (map[string][]*jekyll.FrontMatter, error) {

	mu := new(sync.Mutex)

	lookup := make(map[string][]*jekyll.FrontMatter)

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		if info.IsDir() && path != root {

			// SEE THIS? WE'RE "RECURSING"

			if md_opts.Mode == "date" {

				i := filepath.Join(path, html_opts.Input)
				_, err := os.Stat(i)

				if os.IsNotExist(err) {
					RenderDirectory(ctx, path, html_opts, md_opts)
				}
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

		fm, err := FrontMatterForPath(ctx, path, html_opts)

		if err != nil {
			return err
		}

		if fm == nil {
			return nil
		}

		var keys []string

		switch md_opts.Mode {
		case "authors":
			keys = fm.Authors
		case "date":
			ymd := fm.Date.Format("20060102")
			keys = []string{ymd}
		case "tags":
			keys = fm.Tags
		default:
			return errors.New("Invalid or unsupported mode")
		}

		mu.Lock()

		for _, k := range keys {

			posts, ok := lookup[k]

			if ok {
				posts = append(posts, fm)
				lookup[k] = posts
			} else {
				posts = []*jekyll.FrontMatter{fm}
				lookup[k] = posts
			}
		}

		mu.Unlock()
		return nil
	}

	c := crawl.NewCrawler(root)
	c.CrawlDirectories = true

	err := c.Crawl(cb)

	if err != nil {
		return nil, err
	}

	// ensure that everything is sorted by date (reverse chronological)

	for k, unsorted := range lookup {

		count := len(unsorted)

		by_date := make(map[string]*jekyll.FrontMatter)
		dates := make([]string, count)

		for idx, post := range unsorted {

			dt := post.Date.Format(time.RFC3339)

			by_date[dt] = post
			dates[idx] = dt
		}

		sort.Sort(sort.Reverse(sort.StringSlice(dates)))

		sorted := make([]*jekyll.FrontMatter, count)

		for idx, dt := range dates {
			sorted[idx] = by_date[dt]
		}

		lookup[k] = sorted
	}

	return lookup, nil
}

// see notes below about passing a struct for post details

func RenderPosts(ctx context.Context, root string, title string, posts []*jekyll.FrontMatter, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	t := md_opts.MarkdownTemplates.Lookup(md_opts.List)

	if t == nil {

		func_map := template.FuncMap{
			"prune_string": uri.PruneString,
		}

		tm, err := template.New("list").Funcs(func_map).Parse(default_index_list)

		if err != nil {
			return err
		}

		t = tm
	}

	// maybe just pass this to RenderPosts?
	// (20190409/thisisaaronland)

	type Data struct {
		Mode  string
		Title string
		Posts []*jekyll.FrontMatter
	}

	d := Data{
		Mode:  md_opts.Mode,
		Title: title,
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
	fh := ioutil.NopCloser(r)

	parse_opts := parser.DefaultParseOptions()
	fm, buf, err := parser.Parse(fh, parse_opts)

	if err != nil {
		log.Printf("FAILED to parse MD document, because %s\n", err)
		return err
	}

	if re_ymd.MatchString(root) {

		matches := re_ymd.FindStringSubmatch(root)

		str_yyyy := matches[1]
		str_mm := matches[2]
		str_dd := matches[3]

		parse_string := make([]string, 0)
		ymd_string := make([]string, 0)

		if str_yyyy != "" {
			parse_string = append(parse_string, "2006")
			ymd_string = append(ymd_string, str_yyyy)
		}

		if str_mm != "" {
			parse_string = append(parse_string, "01")
			ymd_string = append(ymd_string, str_mm)
		}

		if str_dd != "" {
			parse_string = append(parse_string, "02")
			ymd_string = append(ymd_string, str_dd)
		}

		// Y U SO WEIRD GO...

		dt, err := time.Parse(strings.Join(parse_string, "-"), strings.Join(ymd_string, "-"))

		if err == nil {
			fm.Date = &dt
		}
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

func RenderRollup(ctx context.Context, root string, rollup []string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	t := md_opts.MarkdownTemplates.Lookup(md_opts.Rollup)

	if t == nil {

		func_map := template.FuncMap{
			"prune_string": uri.PruneString,
		}

		tm, err := template.New("rollup").Funcs(func_map).Parse(default_index_rollup)

		if err != nil {
			return err
		}

		t = tm
	}

	sort.Sort(sort.StringSlice(rollup))

	type Data struct {
		Mode   string
		Rollup []string
	}

	d := Data{
		Mode:   md_opts.Mode,
		Rollup: rollup,
	}

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	err := t.Execute(wr, d)

	if err != nil {
		return err
	}

	wr.Flush()

	r := bytes.NewReader(b.Bytes())
	fh := ioutil.NopCloser(r)

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

func FrontMatterForPath(ctx context.Context, path string, html_opts *render.HTMLOptions) (*jekyll.FrontMatter, error) {

	select {

	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

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

func Render(ctx context.Context, path string, html_opts *render.HTMLOptions, md_opts *MarkdownOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	return RenderDirectory(ctx, path, html_opts, md_opts)
}

func main() {

	var input = flag.String("input", "index.md", "What you expect the input Markdown file to be called")
	var output = flag.String("output", "index.html", "What you expect the output HTML file to be called")
	var header = flag.String("header", "", "The name of the (Go) template to use as a custom header")
	var footer = flag.String("footer", "", "The name of the (Go) template to use as a custom footer")
	var list = flag.String("list", "", "The name of the (Go) template to use as a custom list view")
	var rollup = flag.String("rollup", "", "The name of the (Go) template to use as a custom rollup view (for things like tags and authors)")
	var mode = flag.String("mode", "date", "...")

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
		Rollup:            *rollup,
		Mode:              *mode,
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
