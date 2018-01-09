package search

import (
       "errors"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	_ "log"
)

type SQLiteIndexer struct {
	Indexer
}

func NewSQLiteIndexer(path string) (Indexer, error) {
     return nil, errors.New("Please write me")
}

func (i *SQLiteIndexer) Query(q string) (interface{}, error) {
     return nil, errors.New("Please write me")
}

func (i *SQLiteIndexer) IndexDocument(doc *markdown.Document) (*SearchDocument, error) {
     return nil, errors.New("Please write me")
}
