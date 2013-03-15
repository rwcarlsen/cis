
package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"encoding/json"

	"github.com/rwcarlsen/cis/builder"
)

var addr = flag.String("addr", "127.0.0.1:8888", "url for build server master")

func main() {
	flag.Parse()

	branch := flag.Arg(0)
	commit := flag.Arg(1)

	p := builder.Push{
		Ref: branch,
		After: commit,
		Before: commit,
		Commits: []builder.Commit{
			builder.Commit{
				Hash: commit,
				Message: "<forced injection>",
			},
		},
	}

	data, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("posting to url %v", *addr)

	vals := url.Values{}
	vals.Add("payload", string(data))

	full := &url.URL{
		Scheme: "http",
		Path: "/push-update",
		Host: *addr,
		RawQuery: vals.Encode(),
	}

	r, err := http.NewRequest("POST", full.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = (&http.Client{}).Do(r)
	if err != nil {
		log.Fatal(err)
	}
}
