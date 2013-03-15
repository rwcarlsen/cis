
package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"flag"
	"encoding/json"
	"io/ioutil"
	"html/template"

	"github.com/gorilla/mux"
	"github.com/rwcarlsen/cis/builder"
)

var addr = flag.String("addr", "0.0.0.0:8888", "ip and port to listen on")
const (
	MaxHist = 100
	MaxWork = 10
)

var serv *Server
var tmpl = template.Must(template.ParseFiles("status.html"))

func main() {
	flag.Parse()
	serv = LoadServer("./server-data.json")

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
}

func PushHandler(w http.ResponseWriter, r *http.Request) {
	data := []byte(r.FormValue("payload"))

	var p Push
	if err := json.Unmarshal(data, &p); err != nil {
		log.Print(err)
		return
	}

	if len(p.Commits) > 0 {
		// only add the most recent (the new head) commit of the push
		serv.Add(p.Commits[len(p.Commits)-1])
	}
}

type Push struct {
	Before string `json:"before"`
	After string `json:"after"`
	Ref string `json:"ref"`
	Commits []Commit `json:"commits"`
}

type Commit struct {
	Hash string `json:"id"`
	Message string `json:"message"`
	Timestamp string `json:"timestamp"`
	Url string `json:"url"`
	Author map[string]string `json:"author"`
}

type Entry struct {
	Commit
	Results map[string][]builder.Result
}

type Server struct {
	Path string
	Commits []*Entry
}

func LoadServer(fpath string) *Server {
	s := &Server{Path: fpath}

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return s
	}

	if err := json.Unmarshal(data, &s); err != nil {
		log.Fatal(err)
	}
	return s
}

func (s *Server) Save() error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	f, err := os.Create(s.Path)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// ReportWork adds work completed by the named builder for the commit id
// hash.
func (s *Server) ReportWork(name, hash string, results []builder.Result) error {
	for _, e := range s.Commits {
		if e.Hash == hash {
			e.Results[name] = results
			return nil
		}
	}
	return fmt.Errorf("Work reported for untracked commit id: %v", hash)
}

// Add adds commits to the list of tracked hashes that will be distributed
// to builders.  Commit order should be from oldest to newest.
func (s *Server) Add(commits ...Commit) {
	for _, commit := range commits {
		// prevent duplicates
		for _, e := range s.Commits {
			if e.Hash == commit.Hash {
				break
			}
			e := &Entry{Commit: commit, Results: make(map[string][]builder.Result)}
			s.Commits = append(s.Commits, e)
		}
	}

	if len(s.Commits) > MaxHist {
		i := len(s.Commits) - MaxHist
		s.Commits = append([]*Entry{}, s.Commits[i:]...)
	}
}

// GetLog returns the result log for the specified commit hash, builder
// name, and command label.
func (s *Server) GetLog(hash, name, label string) ([]byte, error) {
    err := fmt.Errorf("%v, %v, %v combo has no result", hash, name, label)
	for _, e := range s.Commits {
		if e.Hash == hash {
			results := e.Results[name]
			if len(results) > 0 {
				for _, r := range results {
					if r.Label == label {
						return r.Output, nil
					}
				}
				return nil, err
			}
			return nil, err
		}
	}
	return nil, err
}

// GetWork returns a list of the most recent commit id's that have not yet
// been processed by the named builder
func (s *Server) GetWork(name string) []string {
	refs := []string{}
	for _, e := range s.Commits[:MaxWork] {
		if len(e.Results[name]) == 0 {
			refs = append(refs, e.Hash)
		}
	}
	return refs
}

// Builders returns a list of all builders for which any results have ever
// been received.
func (s *Server) Builders() []*BuilderInfo {
	bm := map[string]bool{}
	b := []*BuilderInfo{}
	for _, e := range s.Commits {
		for name, results := range e.Results {
			if !bm[name] {
				bm[name] = true
				b = append(b, &BuilderInfo{name, results, len(results)})
			}
		}
	}
	return b
}

type BuilderInfo struct {
	Name string
	Results []builder.Result
	N int
}

