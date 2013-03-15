
package builder

import (
	"os"
	"os/exec"
	"net/http"
	"path/filepath"
	"bytes"
	"encoding/json"
	"io"
	"fmt"
)

const (
	WorkPath = "get-work"
	ResultsPath = "post-results"
)

type Result struct {
	Label string
	Cmd string
	Pass bool
	Output []byte
	Error string
}

type Builder struct {
	Root string
	Name string
	Server string
	cmds []*exec.Cmd
	cmdLabels []string
}

func New(name, root, addr string) *Builder {
	dir, _ := filepath.Abs(root)
	return &Builder{
		Root: dir,
		Name: name,
		Server: addr,
	}
}

func (b *Builder) AddCmd(label string, cmd *exec.Cmd) {
	b.cmds = append(b.cmds, cmd)
	b.cmdLabels = append(b.cmdLabels, label)
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
		if err := GitReset(); err != nil {
			return nil, err
		} else if err := GitCheckout(h); err != nil {
			return nil, err
		}

		results := b.runAll()
		if err := b.report(h, results); err != nil {
			return nil, err
		}
	}

	return hashes, nil
}

func (b *Builder) runAll() []Result {
	var output, stderr bytes.Buffer
	multi := io.MultiWriter(&output, &stderr)
	results := make([]Result, len(b.cmds))
	for i, tmp := range b.cmds {
		tmp2 := *tmp // this allows reuse/resetting of the same cmd obj
		cmd := &tmp2
		cmd.Stdout = &output
		cmd.Stderr = multi

		errtxt := ""
		if err := cmd.Run(); err != nil {
			errtxt = err.Error()
		}

		results[i] = Result{
			Label: b.cmdLabels[i],
			Cmd: cmd.Path,
			Pass: stderr.Len() + len(errtxt) == 0 ,
			Output: output.Bytes(),
			Error: errtxt,
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

func GitCheckout(refspec string) error {
	cmd := exec.Command("git", "checkout", refspec)
	return cmd.Run()
}

func GitReset() error {
	cmd := exec.Command("git", "reset", "--hard", "HEAD")
	return cmd.Run()
}


