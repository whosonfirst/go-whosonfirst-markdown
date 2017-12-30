package markdown

import (
	"bufio"
	"bytes"
	// "github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	_ "log"
	"strings"
)

type nopCloser struct {
	io.Reader
}

type WOFRenderer struct {
	bf   *blackfriday.HTMLRenderer
	meta map[string]string
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

	// please replace me with proper templates

	r.bf.RenderHeader(w, ast)

	w.Write([]byte(`<style type="text/css">
		body { font-family: serif; margin:2em; margin-left: 10%; margin-right: 10%; font-size: 1.3em; line-height: 1.5em; }
		img { max-width: 640px !important; max-height: 480px !important; border: 1px dotted #ccc; padding: .5em; margin: 0 auto; margin-bottom:1em; margin-top: 1em; }
	</style> `))

	_, ok := r.meta["title"]

	if ok {
		w.Write([]byte("<h2>" + r.meta["title"] + "</h2>"))
	}
}

func (r *WOFRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {

	// please replace me with proper templates
	r.bf.RenderFooter(w, ast)
}

func (nopCloser) Close() error { return nil }

func ToHTML(md io.ReadCloser) (io.ReadCloser, error) {

	defer md.Close()

	scanner := bufio.NewScanner(md)

	lineno := 0
	meta := make(map[string]string)

	is_jekyll := false

	post := ""

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
			meta[key] = value
			continue
		}

		post += txt + "\n"
	}

	body := []byte(post)

	flags := blackfriday.CommonHTMLFlags
	flags |= blackfriday.CompletePage
	flags |= blackfriday.UseXHTML

	title := ""

	_, ok := meta["title"]

	if ok {
		title = meta["title"]
	}

	params := blackfriday.HTMLRendererParameters{
		Flags: flags,
		Title: title,
	}

	renderer := blackfriday.NewHTMLRenderer(params)

	r := WOFRenderer{
		bf:   renderer,
		meta: meta,
	}

	unsafe := blackfriday.Run(body, blackfriday.WithRenderer(&r))

	// safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	html := bytes.NewReader(unsafe)
	return nopCloser{html}, nil

}
