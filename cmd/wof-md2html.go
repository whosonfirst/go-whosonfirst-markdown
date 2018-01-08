package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"github.com/whosonfirst/go-whosonfirst-markdown/render"
	"github.com/whosonfirst/go-whosonfirst-markdown/writer"
	"log"
	"os"
)

func main() {

	// var output = flag.String("output", "", "A path to write rendered Markdown. Default is STDOUT.")

	flag.Parse()

	parse_opts := parser.DefaultParseOptions()
	html_opts := render.DefaultHTMLOptions()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("Missing markdown file")
	}

	writers := make([]writer.Writer, 0)

	stdout, err := writer.NewStdoutWriter()

	if err != nil {
		log.Fatal(err)
	}

	writers = append(writers, stdout)

	wr, err := writer.NewMultiWriter(writers...)

	if err != nil {
		log.Fatal(err)
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

	wr.Write(fm.Permalink, html)

	os.Exit(0)
}
