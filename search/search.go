package search

import (
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	_ "log"
)

type Indexer interface {
	IndexDocument(doc *markdown.Document) (*SearchDocument, error)
}

type SearchDocument struct {
	Title   string
	Authors []string
	Date    string
	Links   []string
	Body    []string
	Code    []string
	Images  []string
}

func NewSearchDocument(doc *markdown.Document) (*SearchDocument, error) {

	fm := doc.FrontMatter
	body := doc.Body

	search_doc := SearchDocument{
		Title:   fm.Title,
		Authors: fm.Authors,
		Date:    "",
		Links:   []string{},
		Body:    []string{},
		Code:    []string{},
		Images:  []string{},
	}

	params := blackfriday.HTMLRendererParameters{}
	renderer := blackfriday.NewHTMLRenderer(params)

	r := SearchRenderer{
		bf:  renderer,
		doc: search_doc,
	}

	blackfriday.Run(body.Bytes(), blackfriday.WithRenderer(&r))

	return &search_doc, nil
}

type SearchRenderer struct {
	bf  *blackfriday.HTMLRenderer
	doc SearchDocument
}

func (r *SearchRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

	str_value := string(node.Literal)

	switch node.Type {
	case blackfriday.Text:

		if node.Parent.Type == blackfriday.Link {
			r.doc.Links = append(r.doc.Links, str_value)
		} else {
			r.doc.Body = append(r.doc.Body, str_value)
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
		r.doc.Body = append(r.doc.Body, str_value)
	case blackfriday.Link:
		str_dest := string(node.LinkData.Destination)
		r.doc.Links = append(r.doc.Links, str_dest)
	case blackfriday.Image:
		// pass
	case blackfriday.Code:
		r.doc.Code = append(r.doc.Code, str_value)
	case blackfriday.Document:
		break
	case blackfriday.Paragraph:
		// WHAT - PLEASE FIX ME...
	case blackfriday.BlockQuote:
		// WHAT - PLEASE FIX ME...
	case blackfriday.HTMLBlock:
		// WHAT - PLEASE FIX ME...
	case blackfriday.Heading:

		if !entering {
			// WHAT
		}
	case blackfriday.HorizontalRule:
		// pass
	case blackfriday.List:
		// WHAT
	case blackfriday.Item:
		// WHAT
	case blackfriday.CodeBlock:
		r.doc.Code = append(r.doc.Code, str_value)
	case blackfriday.Table:
		// pass
	case blackfriday.TableCell:
		if entering {

			if node.Prev == nil {
				// WHAT
			}

		} else {
			// WHAT
		}
	case blackfriday.TableHead:
		// pass
	case blackfriday.TableBody:
		// pass
	case blackfriday.TableRow:
		// pass
	default:
		panic("Unknown node type " + node.Type.String())
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
