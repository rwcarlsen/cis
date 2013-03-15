
package main

import (
	"flag"
	"log"
	"os/exec"

	"github.com/rwcarlsen/cis/builder"
)

var root = flag.String("path", ".", "path to main repo root dir")
var addr = flag.String("addr", "http://127.0.0.1:8888", "url for build server master")

func main() {
	flag.Parse()
	bldr := builder.New("robert-1", *root, *addr)

	cmd := exec.Command("go", "build", "./...")
	bldr.AddCmd("build-all", cmd)

	hashes, err := bldr.DoWork()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(" processed %v hashes:\n%v\n", len(hashes), hashes)
}
