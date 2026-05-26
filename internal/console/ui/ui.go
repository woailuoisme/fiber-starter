package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

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

func YellowStyle(w io.Writer) lipgloss.Style {
	return style(w, yellow, true)
}

func CyanStyle(w io.Writer) lipgloss.Style {
	return style(w, cyan, true)
}

func GreenStyle(w io.Writer) lipgloss.Style {
	return style(w, green, false)
}

func GrayStyle(w io.Writer) lipgloss.Style {
	return renderer(w).NewStyle().Foreground(lipgloss.Color("244"))
}

// ProgressBar 代表一个美丽的命令行进度条
type ProgressBar struct {
	total   int
	current int
	width   int
	writer  io.Writer
}

// NewProgressBar 创建一个新的进度条实例
func NewProgressBar(w io.Writer, total int) *ProgressBar {
	return &ProgressBar{
		total:  total,
		width:  40,
		writer: w,
	}
}

// SetWidth 设置进度条的字符宽度
func (p *ProgressBar) SetWidth(w int) {
	if w > 0 {
		p.width = w
	}
}

// Advance 将进度递增指定值，并在终端实时渲染
func (p *ProgressBar) Advance(n int) {
	p.current += n
	if p.current > p.total {
		p.current = p.total
	}
	p.render()
}

// SetProgress 将进度设为指定绝对值并刷新渲染
func (p *ProgressBar) SetProgress(curr int) {
	p.current = curr
	if p.current > p.total {
		p.current = p.total
	}
	if p.current < 0 {
		p.current = 0
	}
	p.render()
}

// Finish 完成进度条，并换行
func (p *ProgressBar) Finish() {
	p.current = p.total
	p.render()
	_, _ = fmt.Fprintln(p.writer)
}

func (p *ProgressBar) render() {
	if p.total <= 0 {
		return
	}

	percent := float64(p.current) / float64(p.total)
	filledWidth := int(percent * float64(p.width))
	if filledWidth > p.width {
		filledWidth = p.width
	}

	// 样式设计
	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))              // 绿色代表已完成
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))              // 暗灰代表未完成
	percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true) // 亮橙代表百分比

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", p.width-filledWidth)

	// 如果输出不是终端 TTY，则输出纯文本进度行，避免 \r 回车符破坏 CI 系统的日志
	if !IsTerminal(p.writer) {
		_, _ = fmt.Fprintf(p.writer, "Progress: %d/%d (%d%%)\n", p.current, p.total, int(percent*100))
		return
	}

	// ANSI \r (Carriage Return) 会将光标重置回行首，实现原地更新
	_, _ = fmt.Fprintf(
		p.writer, "\r%s%s  %s",
		filledStyle.Render(filled),
		emptyStyle.Render(empty),
		percentStyle.Render(fmt.Sprintf("%3d%% (%d/%d)", int(percent*100), p.current, p.total)),
	)
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
	if !IsTerminal(w) || os.Getenv("NO_COLOR") != "" {
		r.SetColorProfile(termenv.Ascii)
	}
	return r
}

func IsTerminal(w any) bool {
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
