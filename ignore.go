package ignore

import (
	"bufio"
	"errors"
	"os"
	"slices"
	"strings"
)

type GitIgnore struct {
	lines []string
}

type matchType int

const (
	CharLiteral matchType = iota
	PathSeparator
	QuestionMark
	Asterix
	CharRange
	LeadingDoubleAsterix  // Match in all directories
	MiddleDoubleAsterix   // Zero or more directories
	TrailingDoubleAsterix // Everything inside
)

func matchTypeToString(m matchType) string {
	switch m {
	case CharLiteral:
		return "CharLiteral"
	case PathSeparator:
		return "PathSeparator"
	case QuestionMark:
		return "QuestionMark"
	case Asterix:
		return "Asterix"
	case CharRange:
		return "CharRange"
	case LeadingDoubleAsterix:
		return "LeadingDoubleAsterix:"
	case MiddleDoubleAsterix:
		return "MiddleDoubleAsterix:"
	case TrailingDoubleAsterix:
		return "TrailingDoubleAsterix:"
	}

	return "Invalid match type!"
}

type matchToken struct {
	theType matchType
	chars   string
	ranges  characterRange
}

type characterRange struct {
	negate bool
	ranges []startAndEndIndex
}

type startAndEndIndex struct {
	start byte
	end   byte
}

// Returns the ending index of the character range.
// Handles backslashes since it's always used within a CompileLine()
func parseCharRange(text string, atIndex int) (characterRange, int, error) {
	if atIndex > len(text)-1 {
		return characterRange{}, 0, errors.New("parseCharRange received an invalid atIndex")
	}

	if text[atIndex] != '[' {
		return characterRange{}, 0, errors.New("parseCharRange called on a non-range")
	}

	i := atIndex + 1
	negate := text[i] == '!'
	if negate {
		i++
	}
	startIndex := i

	ret := characterRange{negate: negate}

	var currentStartRange byte
	isEscaped := false
	inRange := false
	for ; i < len(text); i++ {
		c := text[i]

		if !isEscaped && c == '\\' {
			isEscaped = true
			continue
		}

		peekIsDash := i+2 < len(text) && (text[i+1] == '-' && text[i+2] != ']')

		if !isEscaped && i != startIndex && c == ']' {
			return ret, i, nil
		}

		if peekIsDash {
			currentStartRange = c
			i++
			inRange = true
			continue
		}

		if inRange {
			ret.ranges = append(ret.ranges, startAndEndIndex{start: currentStartRange, end: c})
		} else {
			ret.ranges = append(ret.ranges, startAndEndIndex{start: c, end: c})
		}

		isEscaped = false
		inRange = false
	}

	return characterRange{}, 0, errors.New("unclosed character range")
}

func compileLine(line string) ([]matchToken, error) {
	ret := []matchToken{}

	length := len(line)
	shouldAppendTrailingDoubleAsterix := false
	if strings.HasSuffix(line, "/**") {
		shouldAppendTrailingDoubleAsterix = true
		length -= 3
	}

	i := 0
	if strings.HasPrefix(line, "**/") {
		ret = append(ret, matchToken{theType: LeadingDoubleAsterix})
		i += 3
	}

	isEscaped := false
	for ; i < length; i++ {
		c := line[i]

		if !isEscaped && c == '\\' {
			isEscaped = true
			continue
		}

		if isEscaped {
			ret = append(ret, matchToken{theType: CharLiteral, chars: string(c)})
			isEscaped = false
			continue
		}

		//nextThreeCharsMeanMiddleDoubleAsterix := i+3 > len(line)-1 || line[i:i+3] == "**"+string(os.PathSeparator)
		var nextThreeCharsMeanMiddleDoubleAsterix bool
		if i+4 > len(line)-1 {
			nextThreeCharsMeanMiddleDoubleAsterix = false
		} else {
			nextThreeCharsMeanMiddleDoubleAsterix = line[i+1:i+4] == "**"+string(os.PathSeparator)
		}
		switch c {
		case os.PathSeparator:
			if nextThreeCharsMeanMiddleDoubleAsterix {
				ret = append(ret, matchToken{theType: PathSeparator})
				ret = append(ret, matchToken{theType: MiddleDoubleAsterix})
				i += 2
			} else {
				ret = append(ret, matchToken{theType: PathSeparator})
			}
		case '*':
			ret = append(ret, matchToken{theType: Asterix})
		case '?':
			ret = append(ret, matchToken{theType: QuestionMark})
		case '[':
			theRanges, newIndex, err := parseCharRange(line, i)
			if err != nil {
				return ret, err
			}
			ret = append(ret, matchToken{theType: CharRange, ranges: theRanges})
			i = newIndex
		default:
			ret = append(ret, matchToken{theType: CharLiteral, chars: string(c)})
		}

		isEscaped = false
	}

	if shouldAppendTrailingDoubleAsterix {
		ret = append(ret, matchToken{theType: TrailingDoubleAsterix})
	}
	return ret, nil
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

		if strings.HasPrefix(line, "#") {
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
		return line == "" || strings.HasPrefix(line, "#") || strings.TrimRight(line, " ") == ""
	})

	gi := &GitIgnore{lines: lines}
	return gi
}

func charMatch(c byte, token matchToken) bool {
	if token.theType == Asterix {
		return false // Should never happen
	}

	switch token.theType {
	case QuestionMark:
		return true // Matches any single character
	case CharLiteral:
		if c != token.chars[0] {
			return false
		}
	case CharRange:
		for _, r := range token.ranges.ranges {
			if c >= r.start && c <= r.end {
				return !token.ranges.negate
			}
		}
	}

	return true
}

func MatchesLine(line, path string) (bool, error) {
	tokens, err := compileLine(line)
	if err != nil {
		return false, errors.New("invalid gitignore line")
	}

	findFirstMatch := false
	pathIndex := 0
	for _, token := range tokens {
		//fmt.Println("token: " + matchTypeToString(token.theType))
		if token.theType == Asterix {
			findFirstMatch = true
			continue
		}

		if pathIndex >= len(path) {
			return false, nil
		}

		if findFirstMatch {
			for ; pathIndex < len(path); pathIndex++ {
				if charMatch(path[pathIndex], token) {
					//fmt.Println("found first match at index:", pathIndex)
					pathIndex++
					break
				}
			}
			//fmt.Println("after:", pathIndex)
			findFirstMatch = false
			continue
		} else {
			if !charMatch(path[pathIndex], token) {
				//fmt.Println("fail match:", pathIndex)
				return false, nil
			}
		}

		findFirstMatch = false
		pathIndex++
	}

	return pathIndex == len(path), nil
}

func (gi *GitIgnore) MatchesPath(path string) bool {
	// Match through the lines backwards
	for i := len(gi.lines) - 1; i >= 0; i-- {
		line := gi.lines[i]

		negate := line[0] == '!'
		if negate {
			line = line[1:]
		}

		matches, err := MatchesLine(line, path)
		if err != nil {
			continue
		}

		if matches {
			return !negate
		}
	}

	return false
}
