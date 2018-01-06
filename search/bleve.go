package search

import (
	"github.com/whosonfirst/go-whosonfirst-markdown"
)

type BleveIndexer struct {
	Indexer
}

func (i *BleveIndexer) IndexDocument(doc *markdown.Document) (*SearchDocument, error) {

	return NewSearchDocument(doc)
}
