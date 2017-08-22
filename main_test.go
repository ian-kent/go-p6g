package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var tests = []struct {
	grammar string
	tokens  []string
}{
	{`
grammar foo {
	token TOP { <bar> }
	token bar { .* }
}
`, []string{
		"grammar", "foo", "{", "token", "TOP", "{",
		"<bar>", "}", "rule", "bar", "{", ".*",
		"}", "}",
	},
	}}

func TestMain(t *testing.T) {
	Convey("tokenise works", t, func() {
		for _, t := range tests {
			tokens, err := tokenise(t.grammar)
			So(err, ShouldBeNil)
			So(tokens, ShouldResemble, t.tokens)
		}
	})
}
