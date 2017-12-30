package markdown

import (
	"bufio"
	"bytes"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	"strings"
)

type nopCloser struct {
	io.Reader
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

	unsafe := blackfriday.Run(body)
	safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	html := bytes.NewReader(safe)
	return nopCloser{html}, nil

}
