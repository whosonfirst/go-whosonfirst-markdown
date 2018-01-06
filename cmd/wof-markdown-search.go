package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"github.com/whosonfirst/go-whosonfirst-markdown/search"
	"log"
)

func main() {

	flag.Parse()

	opts := parser.DefaultParseOptions()

	for _, path := range flag.Args() {

		fm, b, err := parser.ParseFile(path, opts)

		if err != nil {
			log.Fatal(err)
		}

		doc, err := markdown.NewDocument(fm, b)

		if err != nil {
			log.Fatal(err)
		}

		s, err := search.NewSearchDocument(doc)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(s)
	}
}
