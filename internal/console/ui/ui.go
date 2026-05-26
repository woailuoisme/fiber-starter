package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	green  = "42"
	yellow = "214"
	red    = "196"
	cyan   = "44"
	white  = "252"
)

func Success(w io.Writer, format string, args ...any) {
	writeLine(w, style(w, green, true).Render(fmt.Sprintf(format, args...)))
}

func Warning(w io.Writer, format string, args ...any) {
	writeLine(w, style(w, yellow, true).Render(fmt.Sprintf(format, args...)))
}

func Error(w io.Writer, format string, args ...any) {
	writeLine(w, style(w, red, true).Render(fmt.Sprintf(format, args...)))
}

func Info(w io.Writer, format string, args ...any) {
	writeLine(w, style(w, cyan, true).Render(fmt.Sprintf(format, args...)))
}

func Highlight(w io.Writer, value string) string {
	return style(w, cyan, false).Render(value)
}

func Faint(w io.Writer, value string) string {
	return renderer(w).NewStyle().Faint(true).Render(value)
}

func Method(w io.Writer, methods []string) string {
	color := white
	switch {
	case contains(methods, "GET"):
		color = green
	case contains(methods, "POST"):
		color = yellow
	case contains(methods, "PUT") || contains(methods, "PATCH"):
		color = cyan
	case contains(methods, "DELETE"):
		color = red
	}
	return style(w, color, true).Render(fmt.Sprintf("%-12s", joinMethods(methods)))
}

func writeLine(w io.Writer, value string) {
	_, _ = fmt.Fprintln(w, value)
}

func style(w io.Writer, color string, bold bool) lipgloss.Style {
	return renderer(w).NewStyle().Foreground(lipgloss.Color(color)).Bold(bold)
}

func renderer(w io.Writer) *lipgloss.Renderer {
	r := lipgloss.NewRenderer(w)
	if !isTerminalWriter(w) || os.Getenv("NO_COLOR") != "" {
		r.SetColorProfile(termenv.Ascii)
	}
	return r
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func joinMethods(methods []string) string {
	if len(methods) == 0 {
		return ""
	}
	value := methods[0]
	for _, method := range methods[1:] {
		value += "|" + method
	}
	return value
}
