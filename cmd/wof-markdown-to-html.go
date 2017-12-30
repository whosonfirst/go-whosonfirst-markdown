package main

import (
	"bufio"
	"flag"
	_ "fmt"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Fatal(err)
		}

		ext := filepath.Ext(abs_path)

		if ext != ".md" {
		   	log.Printf("%s doesn't look like a Markdown file\n", abs_path)
			continue
		}

		fh, err := os.Open(abs_path)

		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(fh)

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

		root := filepath.Dir(abs_path)
		index := filepath.Join(root, "index.html")

		out, err := os.Create(index)

		if err != nil {
			log.Fatal(err)
		}

		out.Write(safe)
		out.Close()

	}
}
