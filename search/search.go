package search

import (
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	"log"
	"strings"
)

type Indexer interface {
	IndexDocument(doc *markdown.Document) (*SearchDocument, error)
}

type SearchDocument struct {
	Title   string
	Authors []string
	Date    string
	Links   map[string]int
	Images  map[string]int
	Body    []string
	Code    []string
}

func NewSearchDocument(doc *markdown.Document) (*SearchDocument, error) {

	fm := doc.FrontMatter
	body := doc.Body

	links := make(map[string]int)
	images := make(map[string]int)

	search_doc := SearchDocument{
		Title:   fm.Title,
		Authors: fm.Authors,
		Date:    "",
		Links:   links,
		Body:    []string{},
		Code:    []string{},
		Images:  images,
	}

	params := blackfriday.HTMLRendererParameters{}
	renderer := blackfriday.NewHTMLRenderer(params)

	r := SearchRenderer{
		bf:  renderer,
		doc: &search_doc,
	}

	blackfriday.Run(body.Bytes(), blackfriday.WithRenderer(&r))

	return &search_doc, nil
}

type SearchRenderer struct {
	bf  *blackfriday.HTMLRenderer
	doc *SearchDocument
}

func (r *SearchRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

	str_value := string(node.Literal)

	switch node.Type {
	case blackfriday.Text:

		str_value = strings.Trim(str_value, " ")

		if str_value != "" {

			if node.Parent.Type == blackfriday.Link {

				url := str_value
				_, ok := r.doc.Links[url]

				if ok {
					r.doc.Links[url] += 1
				} else {
					r.doc.Links[url] = 1
				}

			} else {
				// log.Println("TEXT", str_value)
				r.doc.Body = append(r.doc.Body, str_value)
			}
		}
	case blackfriday.Softbreak:
		// pass
	case blackfriday.Hardbreak:
		// pass
	case blackfriday.Emph:
		// pass
	case blackfriday.Strong:
		// pass
	case blackfriday.Del:
		// pass
	case blackfriday.HTMLSpan:
		// pass
	case blackfriday.Link:

		if entering {
			url := string(node.LinkData.Destination)

			_, ok := r.doc.Links[url]

			if ok {
				r.doc.Links[url] += 1
			} else {
				r.doc.Links[url] = 1
			}
		}
	case blackfriday.Image:

		if entering {
			href := string(node.LinkData.Destination)

			_, ok := r.doc.Links[href]

			if ok {
				r.doc.Links[href] += 1
			} else {
				r.doc.Links[href] = 1
			}

		}
		// pass
	case blackfriday.Code:
		r.doc.Code = append(r.doc.Code, str_value)
	case blackfriday.Document:
		break
	case blackfriday.Paragraph:
		// pass
	case blackfriday.BlockQuote:
		// pass
	case blackfriday.HTMLBlock:
		// pass
	case blackfriday.Heading:
		// pass
	case blackfriday.HorizontalRule:
		// pass
	case blackfriday.List:
		// pass
	case blackfriday.Item:
		// pass
	case blackfriday.CodeBlock:
		r.doc.Code = append(r.doc.Code, str_value)
	case blackfriday.Table:
		// pass
	case blackfriday.TableCell:
		// pass
	case blackfriday.TableHead:
		// pass
	case blackfriday.TableBody:
		// pass
	case blackfriday.TableRow:
		// pass
	default:
		log.Println("Unknown node type " + node.Type.String())
	}
	return blackfriday.GoToNext
}

func (r *SearchRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	return
}

func (r *SearchRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	// write to search index here...
	return
}
