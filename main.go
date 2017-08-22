package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
)

// var _, _ = G("foo",
// 	T("TOP", "<baz>"),
// 	R("bar", ".*"),
// 	R("baz", "'exact match'"),
// )

func main() {
	var example = `
grammar foo {
	token TOP { <baz> }
	rule bar { .* }
	rule baz { 'exact match' }
}
`

	tokens, err := tokenise(example)
	if err != nil {
		log.Fatal(err)
	}

	parsed, err := parse(tokens)
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(parsed)

	n, match := parsed["foo"].Match("test")
	spew.Dump(n, match)

	n, match = parsed["foo"].Match("exact match")
	spew.Dump(n, match)
}
