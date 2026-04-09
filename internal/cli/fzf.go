package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

func (c *Cli) triggerFzf() (string, error) {
	if c.term != nil {
		c.term.Restore()
	}

	defer c.term.EnableRaw()

	fmt.Print("\r")

	cmd := exec.Command("fzf")
	cmd.Stdin = strings.NewReader(c.listFiles())

	out, err := cmd.Output()

	if err != nil {
		c.term.EnableRaw()
		return "", err
	}

	selected := strings.TrimSpace(string(out))
	return selected, nil
}

func (c *Cli) listFiles() string {
	cmd := exec.Command("find", ".", "-type", "f")
	out, _ := cmd.Output()
	return string(out)
}
