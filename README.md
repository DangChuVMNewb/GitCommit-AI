# GitCommit AI 🚀

**GitCommit AI** is a powerful CLI tool that leverages Google's **Gemini AI** to automatically generate professional, conventional git commit messages. It streamlines your workflow by handling staging, message generation, and pushing with a single keystroke.

## ✨ Features

- **🤖 AI-Powered:** Generates semantic commit messages using Google's Gemini 1.5 Flash model.
- **⚡️ Fast Workflow:**
  - Auto-stages changes (`git add .`) automatically.
  - Non-blocking, single-keystroke interaction (no need to press Enter).
- **🎨 Visual Feedback:** Displays changed files with GitHub-style coloring (`+` Green for added/mod, `-` Red for deleted).
- **🌐 Multi-language:** Supports generating messages in any language (English, Vietnamese, Japanese, etc.).
- **🔧 Configurable:** Save your API key and default language preference.

## 📦 Installation

### Prerequisites
- **Go** (Golang) 1.20+ installed.
- **Git** installed.
- A **Google Gemini API Key** (Get it free at [Google AI Studio](https://aistudio.google.com/)).

### Build from Source

```bash
git clone https://github.com/dangchuvmnewb/gitcommit-ai.git
cd gitcommit-ai
go build -o gcommit main.go

# Move binary to your path (optional)
mv gcommit /usr/local/bin/
```

## ⚙️ Configuration

### 1. Set API Key
Run this command once to save your Gemini API key:

```bash
gcommit add-api "YOUR_GEMINI_API_KEY"
```

### 2. Set Default Language (Optional)
Set your preferred language for commit messages (e.g., Vietnamese `vi`, English `en`):

```bash
gcommit -def-lang vi
```

## 🚀 Usage

Simply run the tool in your git repository:

```bash
gcommit
```

### Options
- **Override Language temporarily:**
  ```bash
  gcommit -lang ja
  ```

### Interactive Menu
After generating the message, you will see a preview with changed files:

```text
Files changed:
+ main.go
- old_script.sh
```

Press a single key to act:
- **[y]**: Stage all files (`git add .`) and **Commit**.
- **[p]**: Stage, Commit, and **Push** to remote.
- **[n]**: Quit without doing anything.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open-source.
