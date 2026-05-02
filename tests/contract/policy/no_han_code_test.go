package tests

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/require"
)

func TestNoHanOutsideComments_GoAndShell(t *testing.T) {
	repoRoot := testkit.RepoRoot(t)

	t.Run("go", func(t *testing.T) {
		err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
			require.NoError(t, err)
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "vendor" || name == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := assertNoHanOutsideGoComments(path, b); err != nil {
				require.NoError(t, err)
			}
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("sh", func(t *testing.T) {
		err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
			require.NoError(t, err)
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "vendor" || name == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".sh") {
				return nil
			}

			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := assertNoHanOutsideShellComments(path, b); err != nil {
				require.NoError(t, err)
			}
			return nil
		})
		require.NoError(t, err)
	})
}

func assertNoHanOutsideGoComments(path string, src []byte) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return err
	}

	commentRanges := make([][2]int, 0, len(file.Comments))
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			start := fset.PositionFor(c.Pos(), false).Offset
			end := fset.PositionFor(c.End(), false).Offset
			if start >= 0 && end >= start && end <= len(src) {
				commentRanges = append(commentRanges, [2]int{start, end})
			}
		}
	}

	isInComment := func(i int) bool {
		for _, r := range commentRanges {
			if i >= r[0] && i < r[1] {
				return true
			}
		}
		return false
	}

	for i := 0; i < len(src); {
		if isInComment(i) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(src[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		if isHan(r) {
			line, col := lineCol(src, i)
			snippet := extractLine(src, i)
			return &hanError{Path: path, Line: line, Col: col, Snippet: snippet}
		}
		i += size
	}

	return nil
}

func assertNoHanOutsideShellComments(path string, src []byte) error {
	lines := bytes.Split(src, []byte{'\n'})
	for idx, raw := range lines {
		lineNo := idx + 1
		line := raw

		trimmed := bytes.TrimLeft(line, " \t")
		if len(trimmed) == 0 {
			continue
		}

		if trimmed[0] == '#' {
			continue
		}

		for i := 0; i < len(line); {
			r, size := utf8.DecodeRune(line[i:])
			if r == utf8.RuneError && size == 1 {
				i++
				continue
			}
			if isHan(r) {
				col := i + 1
				snippet := string(line)
				return &hanError{Path: path, Line: lineNo, Col: col, Snippet: snippet}
			}
			i += size
		}
	}
	return nil
}

func isHan(r rune) bool {
	return unicode.Is(unicode.Han, r)
}

type hanError struct {
	Path    string
	Line    int
	Col     int
	Snippet string
}

func (e *hanError) Error() string {
	return e.Path + ":" + itoa(e.Line) + ":" + itoa(e.Col) + ": Han character outside comments: " + e.Snippet
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func lineCol(src []byte, off int) (line int, col int) {
	line = 1
	col = 1
	for i := 0; i < off && i < len(src); i++ {
		if src[i] == '\n' {
			line++
			col = 1
			continue
		}
		col++
	}
	return line, col
}

func extractLine(src []byte, off int) string {
	start := off
	for start > 0 && src[start-1] != '\n' {
		start--
	}
	end := off
	for end < len(src) && src[end] != '\n' {
		end++
	}
	return string(src[start:end])
}
