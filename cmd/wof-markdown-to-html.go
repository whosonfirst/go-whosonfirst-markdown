package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-markdown/render"
	"github.com/whosonfirst/go-whosonfirst-markdown/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RenderDirectory(ctx context.Context, path string, opts *render.HTMLOptions) error {

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:

			if info.IsDir() {
				return nil
			}

			return RenderPath(ctx, path, opts)
		}
	}

	c := crawl.NewCrawler(path)
	return c.Crawl(cb)
}

func RenderPath(ctx context.Context, path string, opts *render.HTMLOptions) error {

	select {

	case <-ctx.Done():
		return nil
	default:

		abs_path, err := filepath.Abs(path)

		if err != nil {
			return err
		}

		fname := filepath.Base(abs_path)

		if fname != opts.Input {
			return nil
		}

		in, err := os.Open(abs_path)

		if err != nil {
			return err
		}

		defer in.Close()

		root := filepath.Dir(abs_path)
		
		parts := strings.Split(root, "/")
		count := len(parts)

		yyyy := parts[ (count - 1) - 3 ]
		mm := parts[ (count - 1) - 2 ]
		dd := parts[ (count - 1) - 1 ]				
		post := parts[ (count - 1) ]

		t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", yyyy, mm, dd))

		if err != nil {
			return err
		}

		uri := fmt.Sprintf("/blog/%s/%s/%s/%s/", yyyy, mm, dd, post)
		
		hints := render.DefaultHTMLHints()
		hints.Date = t.Format("January 02, 2006")
		hints.URI = uri

		// log.Println(hints)
		
		html, err := render.RenderHTML(in, opts, hints)

		if err != nil {
			return err
		}

		if opts.Output == "STDOUT" {
			_, err = io.Copy(os.Stdout, html)
			return err

		}

		index := filepath.Join(root, opts.Output)

		out, err := atomicfile.New(index, os.FileMode(0644))

		if err != nil {
			return err
		}

		defer out.Close()

		_, err = io.Copy(out, html)

		if err != nil {
			out.Abort()
			return err
		}

		log.Printf("wrote %s", index)
		return nil
	}
}

func Render(ctx context.Context, path string, opts *render.HTMLOptions) error {

	select {
	case <-ctx.Done():
		return nil
	default:

		switch opts.Mode {

		case "files":
			return RenderPath(ctx, path, opts)
		case "directory":
			return RenderDirectory(ctx, path, opts)
		default:
			return errors.New("Unknown or invalid mode")
		}
	}

}

func main() {

	var mode = flag.String("mode", "files", "Valid modes are: files, directory")
	var input = flag.String("input", "index.md", "What you expect the input Markdown file to be called")
	var output = flag.String("output", "index.html", "What you expect the output HTML file to be called")
	var header = flag.String("header", "", "The path to a custom (Go) template to use as header for your HTML output")
	var footer = flag.String("footer", "", "The path to a custom (Go) template to use as a footer for your HTML output")

	flag.Parse()

	opts := render.DefaultHTMLOptions()
	opts.Mode = *mode
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
