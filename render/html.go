package render

import (
	"bufio"
	"bytes"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"io"
	"log"
	"strings"
)

type Meta struct {
	Title   string
	Excerpt string
	Image   string
	Authors []string
	Tags    []string
}

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
	meta   *Meta
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

func RenderHTML(md io.ReadCloser, opts *HTMLOptions) (io.ReadCloser, error) {

	defer md.Close()

	scanner := bufio.NewScanner(md)

	lineno := 0

	is_jekyll := false

	post := ""

	m := Meta{
		Title:   "",
		Excerpt: "",
		Authors: []string{},
		Tags:    []string{},
	}

	s2l := func(s string) []string {
		s = strings.TrimLeft(s, "[")
		s = strings.TrimRight(s, "]")

		l := make([]string, 0)

		for _, str := range strings.Split(s, ",") {
			str = strings.Trim(str, " ")
			l = append(l, str)
		}

		return l
	}

	for scanner.Scan() {

		lineno += 1

		txt := scanner.Text()
		ln := strings.Trim(txt, " ")

		if lineno == 1 && txt == "---" {
			is_jekyll = true
			continue
		}

		if is_jekyll && txt == "---" {
			is_jekyll = false
			continue
		}

		if is_jekyll {

			kv := strings.Split(ln, ":")
			key := strings.Trim(kv[0], " ")
			value := strings.Trim(kv[1], " ")

			switch key {
			case "title":
				m.Title = value
			case "excerpt":
				m.Excerpt = value
			case "image":
				m.Image = value
			case "authors":
				m.Authors = s2l(value)
			case "tag":
				m.Tags = s2l(value)
			default:
				// pass

			}

			continue
		}

		post += txt + "\n"
	}

	body := []byte(post)

	flags := blackfriday.CommonHTMLFlags
	flags |= blackfriday.CompletePage
	flags |= blackfriday.UseXHTML

	params := blackfriday.HTMLRendererParameters{
		Flags: flags,
	}

	renderer := blackfriday.NewHTMLRenderer(params)

	r := WOFRenderer{
		bf:     renderer,
		meta:   &m,
		header: opts.Header,
		footer: opts.Footer,
	}

	unsafe := blackfriday.Run(body, blackfriday.WithRenderer(&r))

	// safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	html := bytes.NewReader(unsafe)
	return nopCloser{html}, nil

}
