
package builder

import (
	"os/exec"
	"net/http"
	"path"
	"path/filepath"
	"bytes"
	"encoding/json"

	"github.com/rwcarlsen/cis/server"
)

const (
	WorkPath = "get-work"
	ResultsPath = "post-results"
)


type Result struct {
	Label string
	Cmd string
	Stdout []byte
	Stderr []byte
	Error error
}

type Builder struct {
	Root string
	Name string
	Server string
	cmds []*Cmd
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

func (b *Builder) DoWork(label string, cmd *exec.Cmd) error {
	resp, err := http.Get(path.Join(b.Server, server.WorkPath))
	if err != nil {
		return err
	}

	hashes := resp.Header[http.CanonicalHeaderKey("hashes")]

	if err := os.Chdir(b.Root); err != nil {
		return err
	}
	for _, h := range hashes {
		if err := GitReset(); err != nil {
			return err
		} else if err := CheckoutCmd(h); err != nil {
			return err
		}

		results := b.runAll(h)
		if err := b.report(results); err != nil {
			return err
		}
	}
}

func (b *Builder) runAll() []Result {
	var stdout, stderr bytes.Buffer
	results := make([]Result, len(b.cmds))
	for i, cmd := range b.cmds {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		results[i] = result{
			Label: b.cmdLabels[i],
			Cmd: cmd.Path,
			Stdout: stdout.Bytes(),
			Stderr: stderr.Bytes(),
			Error: err,
		}
		stdout.Reset()
		stdin.Reset()
	}
	return results
}

func (b *Builder) report(results []Result) error {
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(data)
	resp, err := http.Post(path.Join(b.Server, server.ResultsPath), "text/json", body)
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


