
package builder

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

const (
	MaxHist = 100
	MaxWork = 10
)

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
	Results map[string][]Result
}

type Server struct {
	Path string
	Commits []*Entry
}

func LoadServer(fpath string) (*Server, error) {
	s := &Server{Path: fpath}

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return s, nil
	}

	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) Save() error {
	data, err := json.MarshalIndent(s, "", "    ")
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
func (s *Server) ReportWork(name, hash string, results []Result) error {
	for _, e := range s.Commits {
		if e.Hash == hash {
			e.Results[name] = results
			return nil
		}
	}
	return fmt.Errorf("Work reported for untracked commit id: %v", hash)
}

func (s *Server) dupCommit(hash string) bool {
	for _, e := range s.Commits {
		if e.Hash == hash {
			return true
		}
	}
	return false
}

// Add adds commits to the list of tracked hashes that will be distributed
// to builders.  Commit order should be from oldest to newest.
func (s *Server) Add(commits ...Commit) {
	for _, commit := range commits {
		if !s.dupCommit(commit.Hash) {
			e := &Entry{Commit: commit, Results: make(map[string][]Result)}
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
	n := min(MaxWork, len(s.Commits))
	for _, e := range s.Commits[:n] {
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
	Results []Result
	N int
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

