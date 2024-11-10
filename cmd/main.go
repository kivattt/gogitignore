package ignore

import (
	"fmt"
	"log"

	"github.com/kivattt/gogitignore"
)

func main() {
	gi, err := ignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		log.Fatal(err)
	}

	bool2Str := func(b bool) string {
		if b {
			return "true"
		}
		return "false"
	}

	printThing := func(text string) {
		fmt.Println(text, bool2Str(gi.MatchesPath(text)))
	}

	printThing("hello")
	printThing("meme.go")
}
