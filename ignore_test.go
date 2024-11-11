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

	text := "[abc]"
	got, i, _ := parseCharRange(text, 0)
	if i != len(text)-1 {
		t.Fatal("Incorrect end index, expected " + strconv.Itoa(len(text)-1) + ", but got: " + strconv.Itoa(i))
	}

	expected := characterRange{
		negate: false,
		ranges: []startAndEndIndex{
			{start: 'a', end: 'a'},
			{start: 'b', end: 'b'},
			{start: 'c', end: 'c'},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		printRanges(expected, got)
		t.Fatal("Incorrect parseCharRange output")
	}

	text = "[!abc]"
	got, i, _ = parseCharRange(text, 0)
	if i != len(text)-1 {
		t.Fatal("Incorrect end index, expected " + strconv.Itoa(len(text)-1) + ", but got: " + strconv.Itoa(i))
	}

	expected = characterRange{
		negate: true,
		ranges: []startAndEndIndex{
			{start: 'a', end: 'a'},
			{start: 'b', end: 'b'},
			{start: 'c', end: 'c'},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		printRanges(expected, got)
		t.Fatal("Incorrect parseCharRange output")
	}

	text = "[--0abc-z]"
	got, i, _ = parseCharRange(text, 0)
	if i != len(text)-1 {
		t.Fatal("Incorrect end index, expected " + strconv.Itoa(len(text)-1) + ", but got: " + strconv.Itoa(i))
	}

	expected = characterRange{
		negate: false,
		ranges: []startAndEndIndex{
			{start: '-', end: '0'},
			{start: 'a', end: 'a'},
			{start: 'b', end: 'b'},
			{start: 'c', end: 'z'},
		},
	}

	if !reflect.DeepEqual(expected, got) {
		printRanges(expected, got)
		t.Fatal("Incorrect parseCharRange output")
	}

	text = "[-]"
	got, i, _ = parseCharRange(text, 0)
	if i != len(text)-1 {
		t.Fatal("Incorrect end index, expected " + strconv.Itoa(len(text)-1) + ", but got: " + strconv.Itoa(i))
	}

	expected = characterRange{
		negate: false,
		ranges: []startAndEndIndex{
			{start: '-', end: '-'},
		},
	}

	if !reflect.DeepEqual(expected, got) {
		printRanges(expected, got)
		t.Fatal("Incorrect parseCharRange output")
	}

	text = "[-]"
	got, i, _ = parseCharRange(text, 0)
	if i != len(text)-1 {
		t.Fatal("Incorrect end index, expected " + strconv.Itoa(len(text)-1) + ", but got: " + strconv.Itoa(i))
	}

	expected = characterRange{
		negate: false,
		ranges: []startAndEndIndex{
			{start: '-', end: '-'},
		},
	}

	if !reflect.DeepEqual(expected, got) {
		printRanges(expected, got)
		t.Fatal("Incorrect parseCharRange output")
	}

	text = "[-"
	_, _, err := parseCharRange(text, 0)
	if err == nil {
		t.Fatal("Expected unclosed error, but got nil")
	}
}

func TestCompileLine(t *testing.T) {
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

	got, _ := compileLine("**/")
	expected := []matchToken{
		{theType: LeadingDoubleAsterix},
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatal("Incorrect CompileLine output")
	}

	got, _ = compileLine("/**")
	expected = []matchToken{
		{theType: TrailingDoubleAsterix},
	}
	if !reflect.DeepEqual(expected, got) {
		printMatchTokens(expected, got)
		t.Fatal("Incorrect CompileLine output")
	}

	got, _ = compileLine("*.go")
	expected = []matchToken{
		{theType: Asterix},
		{theType: CharLiteral, chars: "."},
		{theType: CharLiteral, chars: "g"},
		{theType: CharLiteral, chars: "o"},
	}
	if !reflect.DeepEqual(expected, got) {
		printMatchTokens(expected, got)
		t.Fatal("Incorrect CompileLine output")
	}

	got, _ = compileLine("*.go")
	expected = []matchToken{
		{theType: Asterix},
		{theType: CharLiteral, chars: "."},
		{theType: CharLiteral, chars: "g"},
		{theType: CharLiteral, chars: "o"},
	}
	if !reflect.DeepEqual(expected, got) {
		printMatchTokens(expected, got)
		t.Fatal("Incorrect CompileLine output")
	}

	got, _ = compileLine("")
	expected = []matchToken{}
	if !reflect.DeepEqual(expected, got) {
		printMatchTokens(expected, got)
		t.Fatal("Incorrect CompileLine output")
	}
}
