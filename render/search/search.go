package search

import (
       "github.com/whosonfirst/go-whosonfirst-markdown"
)

type SearchDocument struct {
     Title string
     Authors []string
     Date string
     Links []string
     Body string
}

type Indexer interface {
     IndexDocument(doc *markdown.Document) (*SearchDocument, error)
}