package writer

import (
	"errors"
	"github.com/facebookgo/atomicfile"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FSWriter struct {
	Writer
	root string
}

func NewFSWriter(root string) (Writer, error) {

	abs_root, err := filepath.Abs(root)

	if err != nil {
		return nil, err
	}

	info, err := os.Stat(abs_root)

	if os.IsNotExist(err) {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("Root is not a directory")
	}

	w := FSWriter{
		root: abs_root,
	}

	return &w, nil
}

func (w *FSWriter) Write(path string, fh io.ReadCloser) error {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	rel_path := strings.Replace(abs_path, w.root, "", -1)
	rel_root := filepath.Dir(rel_path)

	_, err = os.Stat(rel_root)

	if os.IsNotExist(err) {

		err = os.MkdirAll(rel_root, 0755)

		if err != nil {
			return err
		}
	}

	out_path := filepath.Join(w.root, rel_path)

	out, err := atomicfile.New(out_path, os.FileMode(0644))

	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, fh)

	if err != nil {
		out.Abort()
		return err
	}

	return nil
}
