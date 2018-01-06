package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"text/template"
)

var fm_template *template.Template

func init() {

	tm := `---
layout: {{ .Layout }}
title: {{ .Title }}
category: {{ .Category}}
excerpt: {{ .Excerpt }}
authors: {{ .Authors }}
image: {{ .Image }}
tags: {{ .Tags }}
---`

	t, err := template.New("frontmatter").Parse(tm)

	if err != nil {
		log.Fatal(err)
	}

	fm_template = t
}

type Document struct {
	FrontMatter *FrontMatter
	Body        *Body
}

func NewDocument(fm *FrontMatter, body *Body) (*Document, error) {

	doc := Document{
		FrontMatter: fm,
		Body:        body,
	}

	return &doc, nil
}

type Body struct {
	*bytes.Buffer
}

func (b *Body) String() string {
	return fmt.Sprintf("%s", b.Bytes())
}

type FrontMatter struct {
	Title     string
	Excerpt   string
	Image     string
	Layout    string
	Category  string
	Authors   []string
	Tags      []string
	Date      string
	URI       string
	Published bool
}

func (fm *FrontMatter) String() string {

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	err := fm_template.Execute(wr, fm)

	if err != nil {
		return err.Error()
	}

	wr.Flush()
	return string(b.Bytes())
}

func EmptyFrontMatter() *FrontMatter {

	fm := FrontMatter{
		Title:     "",
		Excerpt:   "",
		Image:     "",
		Layout:    "",
		Category:  "",
		Published: false,
		Authors:   make([]string, 0),
		Tags:      make([]string, 0),
		Date:      "",
		URI:       "",
	}

	return &fm
}
