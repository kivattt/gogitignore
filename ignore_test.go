package ignore

import (
	"fmt"
	"reflect"
	"testing"
)

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
	gi := CompileIgnoreLines("", "hello", "world", " ", "a  ")
	AssertLinesEqual(t, gi, "hello", "world", "a  ")
}

func TestMatchesPath(t *testing.T) {
}
