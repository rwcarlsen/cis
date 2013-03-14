
package builder

import (
	"os/exec"
	"net/http"
	"path"
	"path/filepath"

	"github.com/rwcarlsen/cis/server"
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
	Results map[string][]Result // map[hash][]Result
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
		}
		if err := CheckoutCmd(h); err != nil {
			return err
		}
		results := b.runAll(h)
	}
}

func (b *Builder) runAll() []Result {
	var stdout, stderr bytes.Buffer
	for i, cmd := range b.cmds {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		r := result{
			Label: b.cmdLabels[i],
			Cmd: 
	}

}

func (b *Builder) ReportWork() error {
	return nil
}

func GitCheckout(refspec string) error {
	cmd := exec.Command("git", "checkout", refspec)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func GitReset() error {
	cmd := exec.Command("git", "reset", "--hard", "HEAD")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}


