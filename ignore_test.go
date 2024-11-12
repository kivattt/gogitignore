package ignore

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func bool2Str(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func AssertLinesEqual(t *testing.T, gi *GitIgnore, lines ...string) {
	if !reflect.DeepEqual(gi.lines, lines) {
		fmt.Println("Expected:")
		for _, line := range lines {
			fmt.Println("    \"" + line + "\"")
		}
		fmt.Println("But got:")
		for _, line := range gi.lines {
			fmt.Println("    \"" + line + "\"")
		}
		t.Fatal("Lines did not equal")
	}
}

func TestCompileIgnoreLines(t *testing.T) {
	gi := CompileIgnoreLines("", "hello", "world", " ", "a  ", "# comment")
	AssertLinesEqual(t, gi, "hello", "world", "a  ")
}

func TestMatchesPath(t *testing.T) {
}

func TestParseCharRange(t *testing.T) {
	_, _, err := parseCharRange("[-", 0)
	if err == nil {
		t.Fatal("Expected unclosed error, but got nil")
	}

	type test struct {
		text     string
		expected characterRange
	}

	tests := []test{
		{
			text: "[ab]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: 'a', end: 'a'},
					{start: 'b', end: 'b'},
				},
			},
		},
		{
			text: "[!ab]",
			expected: characterRange{
				negate: true,
				ranges: []startAndEndIndex{
					{start: 'a', end: 'a'},
					{start: 'b', end: 'b'},
				},
			},
		},
		{
			text: "[--0abc-z]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: '-', end: '0'},
					{start: 'a', end: 'a'},
					{start: 'b', end: 'b'},
					{start: 'c', end: 'z'},
				},
			},
		},
		{
			text: "[-]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: '-', end: '-'},
				},
			},
		},
		{
			text: "[a-z0-9]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: 'a', end: 'z'},
					{start: '0', end: '9'},
				},
			},
		},
		{
			text: "[a-z[0-9]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: 'a', end: 'z'},
					{start: '[', end: '['},
					{start: '0', end: '9'},
				},
			},
		},
		{
			text: "[][!]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: ']', end: ']'},
					{start: '[', end: '['},
					{start: '!', end: '!'},
				},
			},
		},
		{
			text: "[]-]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: ']', end: ']'},
					{start: '-', end: '-'},
				},
			},
		},
		{
			text: "[a-z\\\\]",
			expected: characterRange{
				negate: false,
				ranges: []startAndEndIndex{
					{start: 'a', end: 'z'},
					{start: '\\', end: '\\'},
				},
			},
		},
	}

	printRanges := func(expected, got characterRange) {
		fmt.Println("Expected:")
		fmt.Println("    negate: " + bool2Str(expected.negate))
		for _, e := range expected.ranges {
			fmt.Println("    start: " + string(e.start) + ", end: " + string(e.end))
		}

		fmt.Println("But got:")
		fmt.Println("    negate: " + bool2Str(got.negate))
		for _, e := range got.ranges {
			fmt.Println("    start: " + string(e.start) + ", end: " + string(e.end))
		}
	}

	for _, aTest := range tests {
		got, i, _ := parseCharRange(aTest.text, 0)
		if i != len(aTest.text)-1 {
			t.Fatal("Incorrect end index for text \"" + aTest.text + "\", expected " + strconv.Itoa(len(aTest.text)-1) + ", but got: " + strconv.Itoa(i))
		}
		if !reflect.DeepEqual(got, aTest.expected) {
			printRanges(aTest.expected, got)
			t.Fatal("Incorrect parseCharRange output for text \"" + aTest.text + "\"")
		}
	}
}

func TestCompileLine(t *testing.T) {
	type test struct {
		text     string
		expected []matchToken
	}

	tests := []test{
		{
			text:     "",
			expected: []matchToken{},
		},
		{
			text: "**/",
			expected: []matchToken{
				{theType: LeadingDoubleAsterix},
			},
		},
		{
			text: "/**",
			expected: []matchToken{
				{theType: TrailingDoubleAsterix},
			},
		},
		{
			text: "*.go",
			expected: []matchToken{
				{theType: Asterix},
				{theType: CharLiteral, chars: "."},
				{theType: CharLiteral, chars: "g"},
				{theType: CharLiteral, chars: "o"},
			},
		},
		{
			text: "dir/**/file",
			expected: []matchToken{
				{theType: CharLiteral, chars: "d"},
				{theType: CharLiteral, chars: "i"},
				{theType: CharLiteral, chars: "r"},
				{theType: PathSeparator},
				{theType: MiddleDoubleAsterix},
				{theType: PathSeparator},
				{theType: CharLiteral, chars: "f"},
				{theType: CharLiteral, chars: "i"},
				{theType: CharLiteral, chars: "l"},
				{theType: CharLiteral, chars: "e"},
			},
		},
		{
			text: "[a-z]",
			expected: []matchToken{
				{theType: CharRange, ranges: characterRange{
					negate: false, ranges: []startAndEndIndex{
						{start: 'a', end: 'z'},
					},
				}},
			},
		},
		{
			text: "a[a-z0-9]b",
			expected: []matchToken{
				{theType: CharLiteral, chars: "a"},
				{theType: CharRange, ranges: characterRange{
					negate: false, ranges: []startAndEndIndex{
						{start: 'a', end: 'z'},
						{start: '0', end: '9'},
					},
				}},
				{theType: CharLiteral, chars: "b"},
			},
		},
		{
			text: "[a-z\\0-9]",
			expected: []matchToken{
				{theType: CharRange, ranges: characterRange{
					negate: false, ranges: []startAndEndIndex{
						{start: 'a', end: 'z'},
						{start: '0', end: '9'},
					},
				}},
			},
		},
		{
			text: "[a-z\\\\0-9]",
			expected: []matchToken{
				{theType: CharRange, ranges: characterRange{
					negate: false, ranges: []startAndEndIndex{
						{start: 'a', end: 'z'},
						{start: '\\', end: '\\'},
						{start: '0', end: '9'},
					},
				}},
			},
		},
	}

	printMatchTokens := func(expected, got []matchToken) {
		fmt.Println("Expected:")
		for _, e := range expected {
			fmt.Println("    type: " + matchTypeToString(e.theType) + ", chars: \"" + e.chars + "\"")
		}
		fmt.Println("But got:")
		for _, e := range got {
			fmt.Println("    type: " + matchTypeToString(e.theType) + ", chars: \"" + e.chars + "\"")
		}
	}

	for _, aTest := range tests {
		got, _ := compileLine(aTest.text)
		if !reflect.DeepEqual(aTest.expected, got) {
			printMatchTokens(aTest.expected, got)
			t.Fatal("Incorrect CompileLine output for text \"" + aTest.text + "\"")
		}
	}
}
