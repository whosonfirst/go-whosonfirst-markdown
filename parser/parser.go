package parser

import (
	"bufio"
	"bytes"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"io"
	"strings"
)

func ParseFrontMatter(md io.ReadCloser) (*markdown.FrontMatter, error) {

	fm := markdown.EmptyFrontMatter()
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
			break
		}

		if is_jekyll {

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

			continue
		}
	}

	return fm, nil
}

func ParseMarkdown(md io.ReadCloser) (*markdown.Document, error) {

	fm, err := ParseFrontMatter(md)

	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	_, err = io.Copy(wr, md)

	if err != nil {
		return nil, err
	}

	d := markdown.Document{
		FrontMatter: fm,
		Body:        b.Bytes(),
	}

	return &d, nil
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
