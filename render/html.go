package render

import (
	"bytes"
	"github.com/whosonfirst/go-whosonfirst-markdown/parser"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"io"
	"log"
)

type HTMLOptions struct {
	Mode   string
	Input  string
	Output string
	Header *template.Template
	Footer *template.Template
}

func DefaultHTMLOptions() *HTMLOptions {

	opts := HTMLOptions{
		Mode:   "files",
		Input:  "index.md",
		Output: "output.md",
		Header: nil,
		Footer: nil,
	}

	return &opts
}

type nopCloser struct {
	io.Reader
}

type WOFRenderer struct {
	bf     *blackfriday.HTMLRenderer
	meta   *parser.FrontMatter
	header *template.Template
	footer *template.Template
}

func (r *WOFRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

	switch node.Type {

	case blackfriday.Image:
		return r.bf.RenderNode(w, node, entering)
	default:
		return r.bf.RenderNode(w, node, entering)
	}
}

func (r *WOFRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {

	if r.header == nil {
		r.bf.RenderHeader(w, ast)
		return
	}

	err := r.header.Execute(w, r.meta)

	if err != nil {
		log.Println(err)
	}
}

func (r *WOFRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {

	if r.header == nil {
		r.bf.RenderFooter(w, ast)
		return
	}

	err := r.footer.Execute(w, r.meta)

	if err != nil {
		log.Println(err)
	}

}

func (nopCloser) Close() error { return nil }

func RenderHTML(p *parser.Parsed, opts *HTMLOptions) (io.ReadCloser, error) {

	flags := blackfriday.CommonHTMLFlags
	flags |= blackfriday.CompletePage
	flags |= blackfriday.UseXHTML

	params := blackfriday.HTMLRendererParameters{
		Flags: flags,
	}

	renderer := blackfriday.NewHTMLRenderer(params)

	r := WOFRenderer{
		bf:     renderer,
		meta:   p.FrontMatter,
		header: opts.Header,
		footer: opts.Footer,
	}

	unsafe := blackfriday.Run(p.Body, blackfriday.WithRenderer(&r))

	// safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	html := bytes.NewReader(unsafe)
	return nopCloser{html}, nil

}
