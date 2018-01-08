package parser

import (
	"bufio"
	"bytes"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"github.com/whosonfirst/go-whosonfirst-markdown/jekyll"
	"io"
	"os"
	"strings"
)

func string2string(s string) string {
	s = strings.TrimLeft(s, "\"")
	s = strings.TrimRight(s, "\"")
	return s
}

type ParseOptions struct {
	FrontMatter bool
	Body        bool
}

func DefaultParseOptions() *ParseOptions {

	opts := ParseOptions{
		FrontMatter: true,
		Body:        true,
	}

	return &opts
}

func ParseFile(path string, opts *ParseOptions) (*jekyll.FrontMatter, *markdown.Body, error) {

	fh, err := os.Open(path)

	if err != nil {
		return nil, nil, err
	}

	defer fh.Close()

	return Parse(fh, opts)
}

func Parse(md io.ReadCloser, opts *ParseOptions) (*jekyll.FrontMatter, *markdown.Body, error) {

	fm := jekyll.EmptyFrontMatter()

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	scanner := bufio.NewScanner(md)

	lineno := 0
	is_jekyll := false

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

			if opts.FrontMatter {

				kv := strings.Split(ln, ":")
				key := strings.Trim(kv[0], " ")
				value := strings.Trim(kv[1], " ")

				switch key {
				case "authors":
					fm.Authors = string2list(value)
				case "category":
					fm.Category = value
				case "excerpt":
					fm.Excerpt = value
				case "image":
					fm.Image = value
				case "layout":
					fm.Layout = value
				case "published":
					fm.Published = string2bool(value)
				case "tag":
					fm.Tags = string2list(value)
				case "tags":
					fm.Tags = string2list(value)
				case "title":
					fm.Title = value
				default:
					// pass
				}
			}
			continue
		}

		if opts.Body {
			wr.WriteString(txt + "\n")
		}
	}

	wr.Flush()
	body := markdown.Body{&b}

	return fm, &body, nil
}

func string2list(s string) []string {
	s = strings.TrimLeft(s, "[")
	s = strings.TrimRight(s, "]")

	l := make([]string, 0)

	for _, str := range strings.Split(s, ",") {
		str = strings.Trim(str, " ")
		l = append(l, str)
	}

	return l
}

func string2bool(s string) bool {

	possible := []string{
		"true",
		"y",
		"yes",
	}

	b := false

	for _, p := range possible {

		if strings.ToLower(s) == p {
			b = true
			break
		}
	}

	return b
}
