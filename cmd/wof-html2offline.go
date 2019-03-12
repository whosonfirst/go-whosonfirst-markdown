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

	to_cache := make([]string, 0)
	
	var f func(node *html.Node, writer io.Writer)

	f = func(n *html.Node, w io.Writer) {

		if n.Type == html.ElementNode {

			switch n.Data {
			case "img":

				for _, attr := range n.Attr {

					if attr.Key == "src" {
						to_cache = append(to_cache, attr.Val)
						break
					}
				}
				
			case "link":

				link := attrs2map(n.Attr...)
				
				rel, rel_ok := link["rel"]
				href, href_ok := link["href"]
				
				if rel_ok && href_ok && rel == "stylesheet" {
					to_cache = append(to_cache, href)
				}
				
			case "script":

				script := attrs2map(n.Attr...)

				script_type, script_type_ok := script["type"]
				src, src_ok := script["src"]

				if script_type_ok && src_ok && script_type == "text/javascript" {
					to_cache = append(to_cache, src)
				}

			default:
				// pass
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, w)
		}
	}

	f(doc, out)

	log.Println(to_cache)
	
	return html.Render(out, doc)
}

func attrs2map(attrs ...html.Attribute) map[string]string {

	attrs_map := make(map[string]string)
				
	for _, a := range attrs {
		attrs_map[a.Key] = a.Val
	}
	
	return attrs_map
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
