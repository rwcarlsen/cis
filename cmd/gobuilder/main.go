
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
	if len(flag.Args()) != 1 {
		log.Fatal("Need 1 arg: builder name")
	}
	bldr := builder.New(flag.Arg(0), *root, *addr)

	cmd := exec.Command("git", "fetch", "origin")
	bldr.AddCmd("fetch", cmd, false)

	cmd = exec.Command("git", "checkout")
	bldr.AddCmd("checkout", cmd, true)

	cmd = exec.Command("go", "test", "./...")
	bldr.AddCmd("test-all", cmd, false)

	hashes, err := bldr.DoWork()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("processed %v hashes:\n%v\n", len(hashes), hashes)
}


