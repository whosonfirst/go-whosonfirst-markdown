package search

import (
       "errors"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"gopkg.in/russross/blackfriday.v2"
	"io"
)

type BleveIndexer struct {
     Indexer
}

func (i *BleveIndexer) IndexDocument(doc *markdown.Document) (*SearchDocument, error) {
     return nil, errors.New("Please write me")
}

type BleveRenderer struct {
	bf          *blackfriday.HTMLRenderer
	frontmatter *markdown.FrontMatter
}

func (r *BleveRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

	switch node.Type {

	case blackfriday.Image:
		return r.bf.RenderNode(w, node, entering)
	default:
		return r.bf.RenderNode(w, node, entering)
	}
}

func (r *BleveRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	return
}

func (r *BleveRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	// write to search index here...
	return
}

func RenderBleve(d *markdown.Document) error {

	params := blackfriday.HTMLRendererParameters{}
	renderer := blackfriday.NewHTMLRenderer(params)

	r := BleveRenderer{
		bf:          renderer,
		frontmatter: d.FrontMatter,
	}

	blackfriday.Run(d.Body.Bytes(), blackfriday.WithRenderer(&r))
	return nil
}
