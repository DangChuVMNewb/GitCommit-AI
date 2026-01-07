package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
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
	message, err := ai.GenerateMessage(diff, apiKey, effectiveLang)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	fmt.Printf("\n%s\n", message)
	fmt.Printf("\n[Y] Commit, [P] Push, [N] Quit: ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "y" {
		git.Commit(message)
		ui.Success("Committed!")
	} else if choice == "p" {
		git.Commit(message)
		git.Push()
		ui.Success("Pushed!")
	}
}