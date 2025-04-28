package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/mikerybka/util"
)

func main() {
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		_, err := pull()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// if ok {
		err = stop()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = run()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// }
	})
	http.ListenAndServe(util.RequireEnvVar("LISTEN_ADDR"), nil)
}

func pull() (bool, error) {
	cmd := exec.Command("docker", "pull", "mikerybka/server")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Status: ") {
			line = strings.TrimPrefix("Status: ", line)
			if strings.HasPrefix(line, "Image is up to date for mikerybka/server") {
				return false, nil
			} else {
				return true, nil
			}
		}
	}
	panic("no status output")
}

func stop() error {
	procs, err := dockerPS()
	if err != nil {
		return err
	}
	for _, p := range procs {
		if p.Image == "mikerybka/server" {
			cmd := exec.Command("docker", "stop", p.ID)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s: %s", err, out)
			}
		}
	}
	return nil
}

func dockerPS() ([]Proc, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Image}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	procs := []Proc{}
	for _, line := range lines {
		parts := strings.Split(line, " ")
		id := parts[0]
		image := parts[1]
		proc := Proc{
			ID:    id,
			Image: image,
		}
		procs = append(procs, proc)
	}
	return procs, nil
}

type Proc struct {
	ID    string
	Image string
}

func (p *Proc) Stop() error {
	cmd := exec.Command("docker", "stop", p.ID)
	return cmd.Run()
}

func run() error {
	cmd := exec.Command("docker", "run", "-d", "--network=host", "-v", "/root/data:/root/data", "mikerybka/server")
	return cmd.Run()
}
