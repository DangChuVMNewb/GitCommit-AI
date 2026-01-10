package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

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

func createSmartHTTPClient() *http.Client {
	useCustomDNS := false
	// Chỉ check trên Linux/Android
	if runtime.GOOS == "linux" || runtime.GOOS == "android" {
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
				d := net.Dialer{Timeout: 10 * time.Second}
				// Force Google DNS
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

func GenerateMessage(diff string, apiKey string, lang string) (string, error) {
	modelName := "gemini-3-flash-preview"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	prompt := fmt.Sprintf(`You are an expert developer. Generate a professional git commit message strictly in %s language.
Rules:
1. Header: <emoji> <type>: <summary>
2. Body: Paragraph context + Bullet points.
3. List changed files with the header "Files changed:" (translated to %s).
4. Output MUST be in %s.

Diff:
%s`, lang, lang, lang, diff)

	reqBody := GeminiRequest{Contents: []Content{{Parts: []Part{{Text: prompt}}}}}
	jsonData, _ := json.Marshal(reqBody)

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
	return "", fmt.Errorf("no response")
}
