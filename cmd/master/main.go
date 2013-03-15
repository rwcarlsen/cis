
package main

import (
	"log"
	"net/http"
	"flag"
	"encoding/json"
	"io/ioutil"
	"html/template"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rwcarlsen/cis/builder"
)

var addr = flag.String("addr", "0.0.0.0:8888", "ip and port to listen on")
var branch = flag.String("branch", "master", "branch to track commits for")
var fpath = flag.String("fpath", "./master.json", "path to file of saved server state")

var serv *builder.Server
var tmpl = template.Must(template.ParseFiles("status.html"))

func main() {
	flag.Parse()
	var err error
	serv, err = builder.LoadServer(*fpath)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/build-status", StatusHandler)
	r.HandleFunc("/get-work/{builder:.*}", GetHandler)
	r.HandleFunc("/post-results/{builder:.*}/{hash:.*}", PostHandler)
	r.HandleFunc("/push-update", PushHandler)
	r.HandleFunc("/log/{hash:.*}/{builder:.*}/{label:.*}", LogHandler)

	http.Handle("/", r)
	log.Printf("listening on %v", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, serv); err != nil {
		log.Print(err)
	}
}

func LogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	name := vars["builder"]
	label := vars["label"]

	data, err := serv.GetLog(hash, name, label)
	if err != nil {
		log.Print(err)
		return
	}

	w.Write(data)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["builder"]

	hashes := serv.GetWork(name)
	for _, h := range hashes {
		w.Header().Add("hashes", h)
	}
	log.Printf("Sent %v commits of work to builder %v", len(hashes), name)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		return
	}

	var results []builder.Result
	if err := json.Unmarshal(data, &results); err != nil {
		log.Print(err)
		return
	}

	vars := mux.Vars(r)
	name := vars["builder"]
	hash := vars["hash"]
	if err := serv.ReportWork(name, hash, results); err != nil {
		log.Print(err)
	}
	log.Printf("Received hash %v results from builder %v", hash, name)
}

func PushHandler(w http.ResponseWriter, r *http.Request) {
	data := []byte(r.FormValue("payload"))

	var p builder.Push
	if err := json.Unmarshal(data, &p); err != nil {
		log.Print(err)
		return
	}

	if !strings.Contains(p.Ref, *branch) {
		log.Printf("received push for untracked branch '%v'", p.Ref)
		return
	}

	if len(p.Commits) > 0 {
		c := p.Commits[len(p.Commits)-1]
		log.Printf("Received pushed commit: %v", c.Hash)
		// only add the most recent (the new head) commit of the push
		serv.Add(c)
	}
}

