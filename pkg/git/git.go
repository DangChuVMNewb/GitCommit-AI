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

	fullDiff := string(stagedOut) + "\n" + string(unstagedOut)

	// Get Untracked files (New files)
	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedOut, _ := untrackedCmd.CombinedOutput()
	untrackedFiles := strings.Split(strings.TrimSpace(string(untrackedOut)), "\n")

	for _, file := range untrackedFiles {
		if file == "" {
			continue
		}
		// Read content of untracked file to give context to AI
		content, err := os.ReadFile(file)
		if err == nil {
			// Format like a git diff for new file
			fullDiff += fmt.Sprintf("\ndiff --git a/%s b/%s\nnew file mode 100644\n--- /dev/null\n+++ b/%s\n@@ -0,0 +1 @@\n%s\n", file, file, file, string(content))
		}
	}

	if strings.TrimSpace(fullDiff) == "" {
		return "", fmt.Errorf("no changes found (stage files, modify tracked files, or create new files)")
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
