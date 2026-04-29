package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

func (c *Cli) triggerFzf(list string) (string, error) {
	if c.term != nil {
		c.term.Restore()
	}

	defer c.term.EnableRaw()

	fmt.Print("\r")

	cmd := exec.Command("fzf")
	cmd.Stdin = strings.NewReader(list)

	out, err := cmd.Output()

	if err != nil {
		c.term.EnableRaw()
		return "", err
	}

	selected := strings.TrimSpace(string(out))
	return selected, nil
}
