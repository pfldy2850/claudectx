package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
)

// PrintSuccess prints a success message to stdout.
func PrintSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stdout, successStyle.Render(msg))
}

// PrintError prints an error message to stderr.
func PrintError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, errorStyle.Render("Error: "+msg))
}

// PrintWarn prints a warning message to stderr.
func PrintWarn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, warnStyle.Render("Warning: "+msg))
}

// PrintInfo prints an info message to stdout.
func PrintInfo(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stdout, infoStyle.Render(msg))
}

// FprintSuccess writes a success message to the given writer.
func FprintSuccess(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(w, successStyle.Render(msg))
}
