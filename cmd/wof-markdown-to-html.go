package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Fatal(err)
		}

		ext := filepath.Ext(abs_path)

		if ext != ".md" {
			log.Printf("%s doesn't look like a Markdown file\n", abs_path)
			continue
		}

		in, err := os.Open(abs_path)

		if err != nil {
			log.Fatal(err)
		}

		html, err := markdown.ToHTML(in)

		if err != nil {
			log.Fatal(err)
		}

		root := filepath.Dir(abs_path)
		index := filepath.Join(root, "index.html")

		out, err := os.Create(index)

		if err != nil {
			log.Fatal(err)
		}

		defer out.Close()

		_, err = io.Copy(out, html)

		if err != nil {
			log.Fatal(err)
		}
	}
}
