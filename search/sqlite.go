package search

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-whosonfirst-markdown"
	"log"
	"strings"
)

type SQLiteIndexer struct {
	Indexer
	conn *sql.DB
	dsn  string
}

func NewSQLiteIndexer(dsn string) (Indexer, error) {

	// It seems likely to me that once we understand how this works
	// we will replace most of the code below with go-whosonfirst-sqlite
	// and a series of markdown specific tables but not today...
	// (20180110/thisisaaronland)

	conn, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	sql := "SELECT name FROM sqlite_master WHERE type='table'"

	rows, err := conn.Query(sql)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	has_table := false

	for rows.Next() {

		var name string
		err := rows.Scan(&name)

		if err != nil {
			return nil, err
		}

		if name == "documents" {
			has_table = true
			break
		}
	}

	if !has_table {

		// this needs a "tags" table but don't bother adding that until
		// we figure out what do about using or not using go-whosonfirst-sqlite
		// above (20180110/thisisaaronland)

		schema := `CREATE TABLE documents (
		       id TEXT PRIMARY KEY,
		       title TEXT
		       category TEXT,
		       date TEXT,
		       body TEXT,
		       code TEXT
		);

		CREATE INDEX documents_by_date ON documents (date);
		CREATE INDEX documents_by_body ON documents (body);

		CREATE TABLE authors (
		       post_id TEXT,
		       author TEXT,
		       date TEXT
		);

		CREATE UNIQUE INDEX authors_by_author ON authors (post_id, author);
		CREATE INDEX authors_by_date ON authors (author, date);

		CREATE TABLE links (
		       post_id TEXT,
		       link TEXT,
		       date TEXT
		);

		CREATE UNIQUE INDEX links_by_link ON links (post_id, link);
		CREATE INDEX links_by_date ON links (date);
		`
		_, err = conn.Exec(schema)

		if err != nil {
			return nil, err
		}
	}

	i := SQLiteIndexer{
		conn: conn,
		dsn:  dsn,
	}

	return &i, nil
}

func (i *SQLiteIndexer) Conn() (*sql.DB, error) {
	return i.conn, nil
}

func (i *SQLiteIndexer) DSN() string {
	return i.dsn
}

func (i *SQLiteIndexer) Close() error {
	return i.conn.Close()
}

func (i *SQLiteIndexer) Query(q string) (interface{}, error) {
	return nil, errors.New("Please write me")
}

func (i *SQLiteIndexer) IndexDocument(doc *markdown.Document) (*SearchDocument, error) {

	search_doc, err := NewSearchDocument(doc)

	if err != nil {
		return nil, err
	}

	conn, err := i.Conn()

	if err != nil {
		return nil, err
	}

	tx, err := conn.Begin()

	if err != nil {
		return nil, err
	}

	str_body := strings.Join(search_doc.Body, " ")
	str_code := strings.Join(search_doc.Code, " ")

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO documents (
		id, title, category, date, body, code
	) VALUES (
		?, ?, ?, ?, ?, ?
	)`)

	log.Println(sql)

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(search_doc.Id, search_doc.Title, search_doc.Category, search_doc.Date, str_body, str_code)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return nil, errors.New("Please write me")
}
