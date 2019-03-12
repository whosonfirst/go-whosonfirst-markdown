package main

import (
	"flag"
	"golang.org/x/net/html"
	_ "golang.org/x/net/html/atom"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Parse(in io.Reader, out io.Writer) error {

	doc, err := html.Parse(in)

	if err != nil {
		return err
	}

	var f func(node *html.Node, writer io.Writer)

	f = func(n *html.Node, w io.Writer) {

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, w)
		}
	}

	f(doc, out)

	return html.Render(out, doc)
}

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Fatal(err)
		}

		in, err := os.Open(abs_path)

		if err != nil {
			log.Fatal(err)
		}

		out := os.Stdout

		err = Parse(in, out)

		if err != nil {
			log.Fatal(err)
		}
	}

}
