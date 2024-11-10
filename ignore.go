package ignore

import (
	"bufio"
	"os"
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

func CompileIgnoreLines(lines ...string) (*GitIgnore, error) {
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	slices.DeleteFunc(linesCopy, func(line string) bool {
		return line == "" || strings.TrimRight(line, " ") == ""
	})

	gi := &GitIgnore{lines: linesCopy}
	return gi, nil
}

func (gi *GitIgnore) MatchesPath(path string) bool {

}
