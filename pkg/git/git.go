package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	diff := string(out)
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("no staged changes")
	}
	return diff, nil
}

func Commit(message string) error {
	return exec.Command("git", "commit", "-m", message).Run()
}

func Push() error {
	cmd := exec.Command("git", "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
