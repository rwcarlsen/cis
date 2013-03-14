
package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"flag"
	"encoding/json"
	"io/ioutil"

	//"github.com/rwcarlsen/cis/builder"
)

var addr = flag.String("addr", "0.0.0.0:8888", "ip and port to listen on")
const (
	HistCount = 10
)

var set *CommitSet

func main() {
	flag.Parse()
	set = LoadCommits()

	http.HandleFunc("/build-status", StatusHandler)
	http.HandleFunc("/get-work", GetHandler)
	http.HandleFunc("/post-results", PostHandler)
	http.HandleFunc("/push-update", PushHandler)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}

	_ = set.Save()
}

type CommitSet struct {
	set []Commit
	// maps builder name to hash of last commit processed
	builders map[string]string
}

func LoadCommits() *CommitSet {
	c := &CommitSet{}
	data, err := ioutil.ReadFile("hashes.json")
	if err != nil {
		return c
	}

	if err := json.Unmarshal(data, &c); err != nil {
		log.Fatal(err)
	}
	return c
}

func (c *CommitSet) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f, err := os.Create("hashes.json")
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *CommitSet) Add(p Push) {
	if len(c.set) == HistCount {
		c.set = append([]Commit{p.Commits[0]}, c.set[:len(c.set)-1]...)
	} else {
		c.set = append([]Commit{p.Commits[0]}, c.set...)
	}
}

func (c *CommitSet) WorkFor(builder string) []string {
	return nil
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Write(data)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	

}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	
}

func PushHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received push notification")
	data := []byte(r.FormValue("payload"))
	fmt.Printf("%+v\n", r.Form)

	var push Push
	if err := json.Unmarshal(data, &push); err != nil {
		log.Fatal(err)
	}

	fmt.Println("---------------decoded:--------------")
	fmt.Printf("%+v\n", push)
}

type Push struct {
	Before string `json:"before"`
	After string `json:"after"`
	Ref string `json:"ref"`
	Commits []Commit `json:"commits"`
	Repo Repository `json:"repository"`
}

type Commit struct {
	Id string `json:"id"`
	Message string `json:"message"`
	Timestamp string `json:"timestamp"`
	Url string `json:"url"`
	Author map[string]string `json:"author"`
}

type Repository struct {
	Name string `json:"name"`
	Url string `json:"url"`
}
