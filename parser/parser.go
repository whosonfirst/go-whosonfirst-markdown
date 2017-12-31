package parser

import (
	"bufio"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"io"
	"strings"
)

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

func ParseMarkdown(md io.ReadCloser) (*markdown.Document, error) {

	scanner := bufio.NewScanner(md)

	lineno := 0
	is_jekyll := false

	post := ""

	fm := markdown.FrontMatter{
		Title:   "",
		Excerpt: "",
		Authors: []string{},
		Tags:    []string{},
		Date:    "",
		URI:     "",
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
				fm.Title = value
			case "excerpt":
				fm.Excerpt = value
			case "image":
				fm.Image = value
			case "authors":
				fm.Authors = string2list(value)
			case "tag":
				fm.Tags = string2list(value)
			case "tags":
				fm.Tags = string2list(value)
			default:
				// pass
			}

			continue
		}

		post += txt + "\n"
	}

	body := []byte(post)

	d := markdown.Document{
		FrontMatter: &fm,
		Body:        body,
	}

	return &d, nil
}
