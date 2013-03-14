
package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"flag"
	"encoding/json"
	"io/ioutil"
)

var addr = flag.String("addr", "0.0.0.0:8888", "ip and port to listen on")
var hashes map[string][]string // map[buildername][]hashes

func main() {
	flag.Parse()
	loadHashes()

	http.HandleFunc("/build-status", StatusHandler)
	http.HandleFunc("/get-work", GetHandler)
	http.HandleFunc("/post-results", PostHandler)
	http.HandleFunc("/commit-push", PushHandler)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
	saveHashes()
}

func loadHashes() {
	data, err := ioutil.ReadFile("hashes.json")
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &hashes); err != nil {
		log.Fatal(err)
	}
}

func saveHashes() {
	data, err := json.Marshal(hashes)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("hashes.json")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}
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
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()

	data = data[len("paload="):]

	var push Push
	if err := json.Unmarshal(data, &push); err != nil {
		log.Fatal(err)
	}
	fmt.Println(push)
}

type Push struct {
	Before string `json:"before"`
	After string `json:"after"`
	Ref string `json:"ref"`
	Commits []Commit `json:"commits"`
	repository Repository `json:"repository"`
}

type Commit struct {
	Id string `json:"id"`
	Message string `json:"message"`
	Timestamp string `json:"committed_date"`
	Url string `json:"commit_url"`
	Author map[string]string `json:"author"`
}

type Repository struct {
	Name string `json:"name"`
	Url string `json:"repo_url"`
	Homepage string `json:"homepage"`
}
