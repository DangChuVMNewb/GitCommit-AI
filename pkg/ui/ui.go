package ui

import "fmt"

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

func Step(msg string) {
	fmt.Printf("%s%s==> %s%s\n", ColorBlue, ColorBold, msg, ColorReset)
}

func Success(msg string) {
	fmt.Printf("%s✔ %s%s\n", ColorGreen, msg, ColorReset)
}

func Error(msg string) {
	fmt.Printf("%s✖ %s%s\n", ColorRed, msg, ColorReset)
}

func Info(msg string) {
	fmt.Printf("%s%s\n%s", ColorCyan, msg, ColorReset)
}
