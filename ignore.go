package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type GitIgnore struct {
	lines []string
}

func CompileIgnoreFile(path string) (*GitIgnore, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ret := &GitIgnore{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.TrimRight(line, " ") == "" {
			continue
		}

		ret.lines = append(ret.lines, line)
	}

	return ret, nil
}

func CompileIgnoreLines(lines ...string) *GitIgnore {
	lines = slices.DeleteFunc(lines, func(line string) bool {
		return line == "" || strings.TrimRight(line, " ") == ""
	})

	gi := &GitIgnore{lines: lines}
	return gi
}

func MatchesLine(line, path string) bool {
	match, err := filepath.Match(line, path)
	if err != nil {
		return false
	}

	return match
}

func (gi *GitIgnore) MatchesPath(path string) bool {
	// Match through the lines backwards
	for i := len(gi.lines) - 1; i >= 0; i-- {
		line := gi.lines[i]

		negate := line[0] == '!'
		if negate {
			line = line[1:]
		}

		if MatchesLine(line, path) {
			return !negate
		}
	}

	return false
}
