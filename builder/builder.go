
package builder

import (
	"os"
	"os/exec"
	"net/http"
	"path"
	"path/filepath"
	"bytes"
	"encoding/json"
	"io"
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

func (b *Builder) DoWork(label string, cmd *exec.Cmd) error {
	if err := os.Chdir(b.Root); err != nil {
		return err
	}

	resp, err := http.Get(path.Join(b.Server, WorkPath, b.Name))
	if err != nil {
		return err
	}
	hashes := resp.Header[http.CanonicalHeaderKey("hashes")]

	for _, h := range hashes {
		if err := GitReset(); err != nil {
			return err
		} else if err := GitCheckout(h); err != nil {
			return err
		}

		results := b.runAll()
		if err := b.report(h, results); err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) runAll() []Result {
	var output, stderr bytes.Buffer
	multi := io.MultiWriter(&output, &stderr)
	results := make([]Result, len(b.cmds))
	for i, cmd := range b.cmds {
		cmd.Stdout = &output
		cmd.Stderr = multi
		err := cmd.Run()
		results[i] = Result{
			Label: b.cmdLabels[i],
			Cmd: cmd.Path,
			Pass: stderr.Len() == 0,
			Output: output.Bytes(),
			Error: err.Error(),
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
	_, err = http.Post(path.Join(b.Server, ResultsPath), "text/json", body)
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


