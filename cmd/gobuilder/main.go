
package main

import (
	"flag"
	"log"
	"os/exec"

	"github.com/rwcarlsen/cis/builder"
)

var root = flag.String("path", ".", "path to main repo root dir")
var addr = flag.String("addr", "localhost:8888", "url for build server master")

func main() {
	flag.Parse()
	bldr := builder.New("robert-1", *root, *addr)

	cmd := exec.Command("go", "build", "./...")
	bldr.AddCmd("go-build-all", cmd)

	if err := bldr.DoWork(); err != nil {
		log.Fatal(err)
	}
}
