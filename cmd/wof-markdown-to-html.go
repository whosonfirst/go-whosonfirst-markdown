package main

import (
	"context"
	"errors"
	"flag"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"io"
	"log"
	"os"
	"path/filepath"
)

func RenderDirectory(ctx context.Context, path string, input string, output string) error {

	cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:

			if info.IsDir() {
				return nil
			}

			return RenderPath(ctx, path, input, output)
		}
	}

	c := crawl.NewCrawler(path)
	return c.Crawl(cb)
}

func RenderPath(ctx context.Context, path string, input string, output string) error {

	select {

	case <-ctx.Done():
		return nil
	default:
		abs_path, err := filepath.Abs(path)

		if err != nil {
			return err
		}

		fname := filepath.Base(abs_path)

		if fname != input {
			// log.Printf("%s doesn't look like a Markdown file\n", abs_path)
			return nil
		}

		in, err := os.Open(abs_path)

		if err != nil {
			return err
		}

		defer in.Close()

		html, err := markdown.ToHTML(in)

		if err != nil {
			return err
		}

		root := filepath.Dir(abs_path)
		index := filepath.Join(root, output)

		log.Println("render", index)
		return nil

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

		return nil
	}
}

func Render(ctx context.Context, mode string, path string, input string, output string) error {

	select {
	case <-ctx.Done():
		return nil
	default:

		switch mode {

		case "files":
			return RenderPath(ctx, path, input, output)
		case "directory":
			return RenderDirectory(ctx, path, input, output)
		default:
			return errors.New("Unknown or invalid mode")
		}
	}

}

func main() {

	var input = flag.String("input", "index.md", "...")
	var output = flag.String("output", "index.html", "...")
	var mode = flag.String("mode", "files", "...")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, path := range flag.Args() {

		err := Render(ctx, *mode, path, *input, *output)

		if err != nil {
			log.Println(err)
			cancel()
			break
		}
	}
}
