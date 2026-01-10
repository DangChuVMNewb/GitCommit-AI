package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetDiff() (string, error) {
	// Get staged changes
	stagedCmd := exec.Command("git", "diff", "--cached")
	stagedOut, _ := stagedCmd.CombinedOutput()

	// Get unstaged changes (tracked files)
	unstagedCmd := exec.Command("git", "diff")
	unstagedOut, _ := unstagedCmd.CombinedOutput()

	// Combine them
	fullDiff := string(stagedOut) + "\n" + string(unstagedOut)

	if strings.TrimSpace(fullDiff) == "" {
		return "", fmt.Errorf("no changes found (stage files or modify tracked files)")
	}
	return fullDiff, nil
}

func GetStatus() ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func StageAll() error {
	return exec.Command("git", "add", ".").Run()
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
