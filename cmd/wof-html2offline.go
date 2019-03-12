package main

// https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API/Using_Service_Workers
// https://hacks.mozilla.org/2015/11/offline-service-workers/
// https://serviceworke.rs/strategy-network-or-cache_service-worker_doc.html
// https://developer.mozilla.org/en-US/docs/Web/API/Cache

// really there is no reason for this tool to be in this package, specifically
// it is meant to be a general-purpose "take this page offline" tool but we'll
// leave it here for now... (20190312/thisisaaronland)

import (
	"bufio"
	"bytes"
	"flag"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var sw string
var sw_init string

func init() {

	sw = `var CACHE = '{{ .CacheName }}';

self.addEventListener('install', function(evt) {
  console.log('The service worker is being installed.');
  evt.waitUntil(precache());
});

self.addEventListener('fetch', function(evt) {
  console.log('The service worker is serving the asset.');
  evt.respondWith(fromNetwork(evt.request, 400).catch(function () {
    return fromCache(evt.request);
  }));
});

function precache() {
  return caches.open(CACHE).then(function (cache) {
    return cache.addAll([
	{{ range  $uri := .ToCache }}{{ $uri }},
	{{ end }}
    ]);
  });
}

function fromNetwork(request, timeout) {
  return new Promise(function (fulfill, reject) {
    var timeoutId = setTimeout(reject, timeout);
    fetch(request).then(function (response) {
      clearTimeout(timeoutId);
      fulfill(response);
    }, reject);
  });
}

function fromCache(request) {
  return caches.open(CACHE).then(function (cache) {
    return cache.match(request).then(function (matching) {
      return matching || Promise.reject('no-match');
    });
  });
}`

	sw_init = `window.addEventListener("load", function load(event){
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('{{ .ServiceWorkerURL }}').then(function(registration) {
    console.log('Service worker registration succeeded:', registration);
  }, /*catch*/ function(error) {
    console.log('Service worker registration failed:', error);
  });
} else {
  console.log('Service workers are not supported.');
}
}, false);`

}

func Parse(in io.Reader, html_wr io.Writer, serviceworker_wr io.Writer) error {

	sw_t, err := template.New("service-worker").Parse(sw)

	if err != nil {
		return err
	}

	init_t, err := template.New("service-worker-init").Parse(sw_init)

	if err != nil {
		return err
	}

	doc, err := html.Parse(in)

	if err != nil {
		return err
	}

	type ServiceWorkerVars struct {
		CacheName string
		ToCache   []string
	}

	type ServiceWorkerInitVars struct {
		ServiceWorkerURL string
	}

	to_cache := make([]string, 0)

	var f func(node *html.Node, writer io.Writer)

	f = func(n *html.Node, w io.Writer) {

		if n.Type == html.ElementNode {

			switch n.Data {

			case "head":

				vars := ServiceWorkerInitVars{
					ServiceWorkerURL: "sw.js",
				}

				var buf bytes.Buffer
				wr := bufio.NewWriter(&buf)

				err := init_t.Execute(wr, vars)

				if err != nil {
					log.Println(err)
					return
				}

				wr.Flush()
				
				script_type := html.Attribute{"", "type", "text/javascript"}

				script := html.Node{
					Type:      html.ElementNode,
					DataAtom:  atom.Script,
					Data:      "script",
					Namespace: "",
					Attr:      []html.Attribute{script_type},
				}

				body := html.Node{
					Type: html.TextNode,
					Data: string(buf.Bytes()),
				}

				script.AppendChild(&body)
				n.AppendChild(&script)

			case "img":

				for _, attr := range n.Attr {

					if attr.Key == "src" {
						to_cache = append(to_cache, attr.Val)
						break
					}
				}

			case "link":

				link := attrs2map(n.Attr...)

				rel, rel_ok := link["rel"]
				href, href_ok := link["href"]

				if rel_ok && href_ok && rel == "stylesheet" {
					to_cache = append(to_cache, href)
				}

			case "script":

				script := attrs2map(n.Attr...)

				script_type, script_type_ok := script["type"]
				src, src_ok := script["src"]

				if script_type_ok && src_ok && script_type == "text/javascript" {
					to_cache = append(to_cache, src)
				}

			default:
				// pass
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, html_wr)
		}
	}

	f(doc, html_wr)

	vars := ServiceWorkerVars{
		CacheName: "network-or-cache",
		ToCache:   to_cache,
	}

	err = sw_t.Execute(serviceworker_wr, vars)

	if err != nil {
		return err
	}

	return html.Render(html_wr, doc)
}

func attrs2map(attrs ...html.Attribute) map[string]string {

	attrs_map := make(map[string]string)

	for _, a := range attrs {
		attrs_map[a.Key] = a.Val
	}

	return attrs_map
}

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Fatal(err)
		}

		in, err := os.Open(abs_path)

		if err != nil {
			log.Fatal(err)
		}

		sw_out := ioutil.Discard
		html_out := os.Stdout

		err = Parse(in, html_out, sw_out)

		if err != nil {
			log.Fatal(err)
		}
	}

}
