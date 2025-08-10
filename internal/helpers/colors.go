package helpers

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	// SuccessColor for successful operations
	SuccessColor = color.New(color.FgGreen, color.Bold)

	// ErrorColor for error messages
	ErrorColor = color.New(color.FgRed, color.Bold)

	// WarningColor for warning messages
	WarningColor = color.New(color.FgYellow, color.Bold)

	// InfoColor for informational messages
	InfoColor = color.New(color.FgCyan, color.Bold)

	// TitleColor for titles and headers
	TitleColor = color.New(color.FgMagenta, color.Bold)
)

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	SuccessColor.Printf("✅ "+format+"\n", args...)
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	ErrorColor.Printf("❌ "+format+"\n", args...)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	WarningColor.Printf("⚠️  "+format+"\n", args...)
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	InfoColor.Printf("ℹ️  "+format+"\n", args...)
}

// PrintTitle prints a title
func PrintTitle(format string, args ...interface{}) {
	TitleColor.Printf("🎯 "+format+"\n", args...)
}

// PrintProgress prints a progress message
func PrintProgress(current, total int, message string) {
	InfoColor.Printf("📊 [%d/%d] %s\n", current, total, message)
}

// PrintSeparator prints a visual separator
func PrintSeparator() {
	fmt.Println(strings.Repeat("─", 80))
}

// IsTerminal checks if output is going to a terminal
func IsTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
