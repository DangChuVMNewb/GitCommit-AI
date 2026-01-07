package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// --- Constants ---
const (
	ColorBlue   = "\033[94m"
	ColorGreen  = "\033[92m"
	ColorRed    = "\033[91m"
	ColorYellow = "\033[93m"
	ColorPurple = "\033[95m"
	ColorCyan   = "\033[96m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"
)

// --- Structs ---
type Config struct {
	ApiKey   string `json:"api_key"`
	Language string `json:"language"`
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}
type Content struct {
	Parts []Part `json:"parts"`
}
type Part struct {
	Text string `json:"text"`
}
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}
type Candidate struct {
	Content Content `json:"content"`
}

// --- Config Functions ---
func getConfigDir() string {
	var configDir string
	if runtime.GOOS == "windows" {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	} else {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "commitai")
}

func getConfigFile() string {
	return filepath.Join(getConfigDir(), "config.json")
}

func loadConfig() Config {
	file := getConfigFile()
	data, err := ioutil.ReadFile(file)
	var config Config
	if err == nil {
		json.Unmarshal(data, &config)
	}
	if config.Language == "" {
		config.Language = "en"
	}
	return config
}

func saveConfig(config Config) {
	dir := getConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(getConfigFile(), data, 0644)
}

// --- Helper Functions ---
func printStep(msg string) {
	fmt.Printf("%s%s==> %s%s\n", ColorBlue, ColorBold, msg, ColorReset)
}
func printSuccess(msg string) {
	fmt.Printf("%s✔ %s%s\n", ColorGreen, msg, ColorReset)
}
func printError(msg string) {
	fmt.Printf("%s✖ %s%s\n", ColorRed, msg, ColorReset)
}

func getGitDiff() (string, error) {
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

func gitPush() {
	printStep("Pushing to remote...")
	cmd := exec.Command("git", "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		printError("Push failed: " + err.Error())
	} else {
		printSuccess("Pushed successfully!")
	}
}

// --- Network & AI Logic ---

// createSmartHTTPClient tự động dùng Google DNS nếu hệ thống không có /etc/resolv.conf (Termux)
func createSmartHTTPClient() *http.Client {
	useCustomDNS := false
	if runtime.GOOS == "linux" || runtime.GOOS == "android" {
		// Kiểm tra file cấu hình DNS chuẩn
		if _, err := os.Stat("/etc/resolv.conf"); os.IsNotExist(err) {
			useCustomDNS = true
		}
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if useCustomDNS {
		dialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 10 * time.Second,
				}
				// Force dùng Google DNS
				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 60 * time.Second,
	}
}

func generateCommitMessage(diff string, apiKey string, lang string) (string, error) {
	modelName := "gemini-3-flash-preview"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	langMap := map[string]string{
		"en": "English", "vi": "Vietnamese", "jp": "Japanese",
		"cn": "Chinese", "kr": "Korean",
	}
	targetLang, exists := langMap[lang]
	if !exists {
		targetLang = lang
	}

	prompt := fmt.Sprintf(`You are an expert developer. Generate a git commit message.
Rules:
1. Format: "<emoji> <type>: <subject>" (Conventional Commits).
2. Gitmoji: ✨ feat, 🐛 fix, 📝 docs, ♻️ refactor, 🎨 style, 🚀 perf, 🧪 test.
3. First line max 70 chars.
4. Language: %s.
5. Return ONLY the raw message.

Diff:
%s`, targetLang, diff)

	reqBody := GeminiRequest{Contents: []Content{{Parts: []Part{{Text: prompt}}}}}}
	jsonData, _ := json.Marshal(reqBody)

	// Sử dụng Smart Client
	client := createSmartHTTPClient()
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	json.NewDecoder(resp.Body).Decode(&geminiResp)

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text), nil
	}
	return "", fmt.Errorf("no response from AI")
}

// --- Main Function ---
func main() {
	defLangPtr := flag.String("def-lang", "", "Set default language")
	langPtr := flag.String("lang", "", "Temporary language")
	
	if len(os.Args) > 1 && os.Args[1] == "add-api" {
		if len(os.Args) < 3 {
			printError("Usage: gcommit add-api \"YOUR_KEY\"")
			os.Exit(1)
		}
		config := loadConfig()
		config.ApiKey = os.Args[2]
		saveConfig(config)
		printSuccess("API Key saved!")
		return
	}

	flag.Parse()
	config := loadConfig()

	if *defLangPtr != "" {
		config.Language = *defLangPtr
		saveConfig(config)
		printSuccess("Default language set: " + *defLangPtr)
		return
	}

	effectiveLang := config.Language
	if *langPtr != "" {
		effectiveLang = *langPtr
	}

	apiKey := config.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("API_KEY")
	}

	if apiKey == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\n%s%s⚠️  API Key missing!%s\n", ColorYellow, ColorBold, ColorReset)
		fmt.Printf("Enter Gemini API Key: ")
		input, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(input)
		if apiKey == "" {
			printError("Empty key. Exiting.")
			os.Exit(1)
		}
		config.ApiKey = apiKey
		saveConfig(config)
	}

	diff, err := getGitDiff()
	if err != nil {
		printError("No staged changes. Use 'git add' first.")
		os.Exit(1)
	}
	if len(diff) > 20000 {
		diff = diff[:20000]
	}

	printStep(fmt.Sprintf("Analyzing (Lang: %s)...", effectiveLang))
	message, err := generateCommitMessage(diff, apiKey, effectiveLang)
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("\n%s========================================%s\n", ColorPurple, ColorReset)
		fmt.Printf("%s%sSuggested Message:%s\n", ColorGreen, ColorBold, ColorReset)
		fmt.Printf("%s%s%s\n", ColorCyan, message, ColorReset)
		fmt.Printf("%s----------------------------------------%s\n", ColorPurple, ColorReset)

		fmt.Printf("[%sY%s] Commit\n", ColorGreen, ColorReset)
		fmt.Printf("[%sP%s] Commit & Push\n", ColorYellow, ColorReset)
		fmt.Printf("[%sR%s] Regenerate\n", ColorBlue, ColorReset)
		fmt.Printf("[%sE%s] Edit\n", ColorPurple, ColorReset)
		fmt.Printf("[%sN%s] Quit\n", ColorRed, ColorReset)
		fmt.Print("\n> ")

		input, _ := reader.ReadString('\n')
		choice := strings.ToLower(strings.TrimSpace(input))

		if choice == "y" {
			exec.Command("git", "commit", "-m", message).Run()
			printSuccess("Committed!")
			break
		} else if choice == "p" {
			if err := exec.Command("git", "commit", "-m", message).Run(); err == nil {
				printSuccess("Committed!")
				gitPush()
			} else {
				printError("Commit failed.")
			}
			break
		} else if choice == "r" {
			printStep("Regenerating...")
			message, _ = generateCommitMessage(diff, apiKey, effectiveLang)
		} else if choice == "e" {
			fmt.Print("Enter new message: ")
			newMessage, _ := reader.ReadString('\n')
			if strings.TrimSpace(newMessage) != "" {
				message = strings.TrimSpace(newMessage)
			}
		} else if choice == "n" || choice == "q" {
			break
		}
	}
}