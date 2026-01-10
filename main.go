package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dangchuvmnewb/gitcommit-ai/pkg/ai"
	"github.com/dangchuvmnewb/gitcommit-ai/pkg/config"
	"github.com/dangchuvmnewb/gitcommit-ai/pkg/git"
	"github.com/dangchuvmnewb/gitcommit-ai/pkg/ui"
)

func main() {
	defLangPtr := flag.String("def-lang", "", "Set default language")
	langPtr := flag.String("lang", "", "Temporary language")
	
	if len(os.Args) > 1 && os.Args[1] == "add-api" {
		if len(os.Args) < 3 {
			ui.Error("Usage: gcommit add-api \"YOUR_KEY\"")
			os.Exit(1)
		}
		cfg := config.Load()
		cfg.ApiKey = os.Args[2]
		config.Save(cfg)
		ui.Success("API Key saved!")
		return
	}

	flag.Parse()
	cfg := config.Load()

	if *defLangPtr != "" {
		cfg.Language = *defLangPtr
		config.Save(cfg)
		ui.Success("Default language set: " + *defLangPtr)
		return
	}

	effectiveLang := cfg.Language
	if *langPtr != "" {
		effectiveLang = *langPtr
	}

	apiKey := cfg.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("API_KEY")
	}

	if apiKey == "" {
		fmt.Printf("Enter API Key: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(input)
		if apiKey == "" {
			ui.Error("Empty Key. Exiting.")
			os.Exit(1)
		}
		cfg.ApiKey = apiKey
		config.Save(cfg)
	}

	diff, err := git.GetDiff()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Step(fmt.Sprintf("Generating Pro Message (Lang: %s)...", effectiveLang))
	aiMessage, err := ai.GenerateMessage(diff, apiKey, effectiveLang)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	// Process files for GitHub-style display
	files, _ := git.GetStatus()
	var fileListDisplay strings.Builder
	var fileListPlain strings.Builder

	header := "Files changed:"
	if strings.Contains(strings.ToLower(effectiveLang), "vi") {
		header = "Các tập tin đã thay đổi:"
	}

	fileListDisplay.WriteString("\n" + ui.ColorYellow + header + ui.ColorReset + "\n")
	fileListPlain.WriteString("\n" + header + "\n")

	for _, f := range files {
		// git status --porcelain returns "XY filename"
		// X is staged, Y is unstaged. If either is 'D', it's a deletion logic usually,
		// but simplified: D at start means deleted.
		code := f[:2]
		name := f[3:]
		
		var lineDisplay, linePlain string
		if strings.Contains(code, "D") {
			// Deleted
			lineDisplay = fmt.Sprintf("%s- %s%s", ui.ColorRed, name, ui.ColorReset)
			linePlain = fmt.Sprintf("- %s", name)
		} else {
			// Added / Modified / Renamed
			lineDisplay = fmt.Sprintf("%s+ %s%s", ui.ColorGreen, name, ui.ColorReset)
			linePlain = fmt.Sprintf("+ %s", name)
		}
		
		fileListDisplay.WriteString(lineDisplay + "\n")
		fileListPlain.WriteString(linePlain + "\n")
	}

	fullDisplayMsg := aiMessage + "\n" + fileListDisplay.String()
	finalCommitMsg := aiMessage + "\n" + fileListPlain.String()

	ui.Info(fmt.Sprintf("\n%s", fullDisplayMsg))
	fmt.Printf("\n[Y] Commit, [P] Push, [N] Quit: ")

	// Disable input buffering and echoing to read single key
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	var b = make([]byte, 1)
	os.Stdin.Read(b)
	choice := strings.ToLower(string(b))

	// Restore terminal mode
	exec.Command("stty", "-F", "/dev/tty", "sane").Run()
	fmt.Println() // Add newline after input

	if choice == "y" {
		git.StageAll()
		git.Commit(finalCommitMsg)
		ui.Success("Committed!")
	} else if choice == "p" {
		git.StageAll()
		git.Commit(finalCommitMsg)
		git.Push()
		ui.Success("Pushed!")
	}
}