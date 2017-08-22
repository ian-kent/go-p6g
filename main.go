package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	var example = `
grammar foo {
	token TOP { <bar> }
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

	parsed["foo"].Match("test")
}
