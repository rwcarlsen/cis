package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	WorkPath    = "get-work"
	ResultsPath = "post-results"
)

type Result struct {
	Label  string
	Cmd    string
	Pass   bool
	Output []byte
	Error  string
}

type Builder struct {
	Root      string
	Name      string
	Server    string
	cmds      []*exec.Cmd
	cmdLabels []string
	cmdWithHash []bool
}

func New(name, root, addr string) *Builder {
	dir, _ := filepath.Abs(root)
	return &Builder{
		Root:   dir,
		Name:   name,
		Server: addr,
	}
}

// AddCmd adds a command to be executed every time this builder receives a
// unit of work from the master build server.  If withHash is true, the
// unit-of-work commit hash will be included as the last argument to the
// command.
func (b *Builder) AddCmd(label string, cmd *exec.Cmd, withHash bool) {
	b.cmds = append(b.cmds, cmd)
	b.cmdLabels = append(b.cmdLabels, label)
	b.cmdWithHash = append(b.cmdWithHash, withHash)
}

// DoWork fetches a set of commit hashes from the master build server, executes
// all commands for each commit, and reports the result back to the master.  It
// returns a list of commit hashes that were processed and any errors that
// occured.
func (b *Builder) DoWork() ([]string, error) {
	if err := os.Chdir(b.Root); err != nil {
		return nil, err
	}

	resp, err := http.Get(b.Server + "/" + WorkPath + "/" + b.Name)
	if err != nil {
		return nil, err
	}
	hashes := resp.Header[http.CanonicalHeaderKey("hashes")]

	for _, h := range hashes {
		results := b.runAll(h)
		if err := b.report(h, results); err != nil {
			return nil, err
		}
	}

	return hashes, nil
}

func (b *Builder) runAll(hash string) []Result {
	var output, stderr bytes.Buffer
	multi := io.MultiWriter(&output, &stderr)
	results := make([]Result, len(b.cmds))
	for i, tmp := range b.cmds {
		tmp2 := *tmp // this allows reuse/resetting of the same cmd obj
		cmd := &tmp2
		cmd.Stdout = &output
		cmd.Stderr = multi

		if b.cmdWithHash[i] {
			cmd.Args = append(cmd.Args, hash)
		}

		errtxt := ""
		if err := cmd.Run(); err != nil {
			errtxt = err.Error()
		}

		results[i] = Result{
			Label:  b.cmdLabels[i],
			Cmd:    cmd.Path,
			Pass:   stderr.Len()+len(errtxt) == 0,
			Output: output.Bytes(),
			Error:  errtxt,
		}
		output.Reset()
		stderr.Reset()
	}
	return results
}

func (b *Builder) report(hash string, results []Result) error {
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(data)
	full := fmt.Sprintf("%v/%v/%v/%v", b.Server, ResultsPath, b.Name, hash)
	_, err = http.Post(full, "text/json", body)
	if err != nil {
		return err
	}

	return nil
}

