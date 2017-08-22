package main

import (
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type tokeniserTest struct {
	grammar string
	tokens  []string
}

type parserTest struct {
	grammar string
	result  *grammar
}

type matcherTest struct {
	name    string
	grammar string
	test    string
	result  *matchResult
}

var tokeniserTests = []tokeniserTest{
	{`
grammar foo {
	token TOP { <bar> }
	token bar { .* }
}
`,
		[]string{
			"grammar", "foo", "{", "token", "TOP", "{",
			"<bar>", "}", "token", "bar", "{", ".*",
			"}", "}",
		},
	},
	{
		`grammar foo { token TOP { <bar> } token bar { 'exact match' } }`,
		[]string{
			"grammar", "foo", "{", "token", "TOP", "{",
			"<bar>", "}", "token", "bar", "{", "'exact match'",
			"}", "}",
		},
	},
}

var parserTests = []parserTest{}
var matcherTests = []matcherTest{
	{
		"foo",
		`grammar foo { token TOP { <bar> } token bar { 'exact match' } }`,
		"exact match",
		&matchResult{
			match: "exact match",
			children: map[string]*matchResult{
				"bar": &matchResult{
					match: "exact match",
				},
			},
		},
	},
	{
		"foo",
		`grammar foo { token TOP { <bar> } token bar { .* } }`,
		"exact match",
		&matchResult{
			match: "exact match",
			children: map[string]*matchResult{
				"bar": &matchResult{
					match: "exact match",
				},
			},
		},
	},
}

func TestMain(t *testing.T) {
	Convey("tokenise works", t, func() {
		for _, t := range tokeniserTests {
			tokens, err := tokenise(t.grammar)
			So(err, ShouldBeNil)
			So(tokens, ShouldResemble, t.tokens)
		}
	})
	Convey("parser works", t, func() {
		for _, t := range parserTests {
			tokens, err := tokenise(t.grammar)
			So(err, ShouldBeNil)
			parsed, err := parse(tokens)
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble, t.result)
		}
	})
	Convey("matcher works", t, func() {
		for _, t := range matcherTests {
			tokens, err := tokenise(t.grammar)
			So(err, ShouldBeNil)
			parsed, err := parse(tokens)
			So(err, ShouldBeNil)
			So(parsed, ShouldContainKey, t.name)
			log.Println(parsed)
			_, res := parsed["foo"].Match(t.test)
			So(res, ShouldResemble, t.result)
		}
	})
}
