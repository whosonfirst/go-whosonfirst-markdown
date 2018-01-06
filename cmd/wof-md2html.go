package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"github.com/whosonfirst/go-whosonfirst-markdown/render"
	"io"
	"log"
	"os"
)

func main() {

	var output = flag.String("output", "", "A path to write rendered Markdown. Default is STDOUT.")

	flag.Parse()

	parse_opts := parser.DefaultParseOptions()
	html_opts := render.DefaultHTMLOptions()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("Missing markdown file")
	}

	path := args[0]

	fm, body, err := parser.ParseFile(path, parse_opts)

	if err != nil {
		log.Fatal(err)
	}

	doc, err := markdown.NewDocument(fm, body)

	if err != nil {
		log.Fatal(err)
	}

	html, err := render.RenderHTML(doc, html_opts)

	if err != nil {
		log.Fatal(err)
	}

	var wr io.Writer

	if *output == "" {
		wr = os.Stdout
	} else {
		fh, err := os.Create(*output)

		if err != nil {
			log.Fatal(err)
		}

		wr = fh
	}

	_, err = io.Copy(wr, html)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
