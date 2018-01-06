package search

import (
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	"log"
)

type SearchRenderer struct {
	bf  *blackfriday.HTMLRenderer
	doc SearchDocument
}

/*

   Document NodeType = iota
   BlockQuote
   List
   Item
   Paragraph
   Heading
   HorizontalRule
   Emph
   Strong
   Del
   Link
   Image
   Text
   HTMLBlock
   CodeBlock
   Softbreak
   Hardbreak
   Code
   HTMLSpan
   Table
   TableCell
   TableHead
   TableBody
   TableRow

*/

func (r *SearchRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

	log.Println("TYPE", node.Type, node.String())

	switch node.Type {

	case blackfriday.Image:
		return r.bf.RenderNode(w, node, entering)
	default:
		return r.bf.RenderNode(w, node, entering)
	}
}

func (r *SearchRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	return
}

func (r *SearchRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	// write to search index here...
	return
}

type SearchDocument struct {
	Title   string
	Authors []string
	Date    string
	Links   []string
	Body    string
}

func NewSearchDocument(doc *markdown.Document) (*SearchDocument, error) {

	fm := doc.FrontMatter
	body := doc.Body

	search_doc := SearchDocument{
		Title:   fm.Title,
		Authors: fm.Authors,
		Date:    "",
		Links:   []string{},
		Body:    "",
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

type Indexer interface {
	IndexDocument(doc *markdown.Document) (*SearchDocument, error)
}
