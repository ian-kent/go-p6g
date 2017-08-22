package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	expectingGrammar = iota
	expectingLabel
	expectingMatch
	expectingMatchContents
	expectingOpenCurly
	expectingCloseCurly

	matchTypeToken = iota
	matchTypeRule

	ctxGrammar = iota
	ctxmatcher
)

type grammar struct {
	name    string
	entries map[string]*matcher
	top     *matcher
}

func (g grammar) Match(s string) bool {
	g.top.Match(s)
	return false
}

type matcher struct {
	name  string
	typ   int
	match []*matchAtom
}

func (t matcher) Match(s string) bool {
	switch t.typ {
	case matchTypeRule:
	case matchTypeToken:
	}
	return false
}

func (t *matcher) init(g *grammar) error {
	for _, a := range t.match {
		if _, err := a.init(g); err != nil {
			return err
		}
	}
	return nil
}

type matchAtom struct {
	atom string
	ref  *matcher
}

func (t *matchAtom) init(g *grammar) (ref *matcher, err error) {
	if strings.HasPrefix(t.atom, "<") &&
		strings.HasSuffix(t.atom, ">") {
		refname := t.atom[1 : len(t.atom)-1]
		tr, ok := g.entries[refname]
		if !ok {
			return nil, fmt.Errorf("ref %s not found for atom %s", refname, t.atom)
		}
		t.ref = tr
		return tr, nil
	}

	return nil, nil
}

type parsed map[string]*grammar

func parse(tokens []string) (parsed parsed, err error) {
	parsed = make(map[string]*grammar)
	var gram *grammar
	var tokOrRule *matcher
	var ctx int

	var state = expectingGrammar

	for _, tok := range tokens {
		switch state {
		case expectingGrammar:
			if tok != "grammar" {
				err = fmt.Errorf("expecting grammar, got %s", tok)
				return
			}
			state = expectingLabel
			ctx = ctxGrammar
			gram = &grammar{
				entries: make(map[string]*matcher),
			}
			continue
		case expectingLabel:
			switch ctx {
			case ctxGrammar:
				gram.name = tok
			case ctxmatcher:
				tokOrRule.name = tok
			}
			state = expectingOpenCurly
			continue
		case expectingOpenCurly:
			if tok != "{" {
				err = fmt.Errorf("expecting open curly, got %s", tok)
				return
			}
			switch ctx {
			case ctxGrammar:
				state = expectingMatch
			case ctxmatcher:
				state = expectingMatchContents
			}
			continue
		case expectingMatch:
			tokOrRule = &matcher{}
			switch tok {
			case "token":
				tokOrRule.typ = matchTypeToken
			case "rule":
				tokOrRule.typ = matchTypeRule
			case "}":
				state = expectingGrammar
				if _, ok := gram.entries["TOP"]; !ok {
					err = fmt.Errorf("expecting TOP in grammar %s", gram.name)
					return
				}
				gram.top = gram.entries["TOP"]
				parsed[gram.name] = gram
				gram = &grammar{
					entries: make(map[string]*matcher),
				}
				continue
			default:
				err = fmt.Errorf("expecting keyword 'token' or 'rule', got %s", tok)
				return
			}
			ctx = ctxmatcher
			state = expectingLabel
		case expectingMatchContents:
			switch tok {
			case "}":
				state = expectingMatch
				gram.entries[tokOrRule.name] = tokOrRule
				tokOrRule = &matcher{}
			default:
				tokOrRule.match = append(tokOrRule.match, &matchAtom{atom: tok})
			}
		default:
			err = fmt.Errorf("unexpected state, got %s", tok)
			return
		}
	}

	for _, g := range parsed {
		for _, t := range g.entries {
			if err := t.init(g); err != nil {
				return nil, fmt.Errorf("error parsing token or rule %s", t.name)
			}
		}
	}

	return
}

func tokenise(s string) (tokens []string, err error) {
	var buf string
	var pos = -1
	var str bool
	//var esc bool

	rdr := bufio.NewReader(bytes.NewReader([]byte(s)))

	for {
		r, l, err := rdr.ReadRune()
		if err != nil {
			if err == io.EOF {
				return tokens, nil
			}
			return tokens, err
		}

		pos += l

		if str {
			if r == '\'' {
				str = false
			}
			buf += string(r)
			continue
		} else if r == '\'' {
			str = true
			buf += string(r)
			continue
		}

		if strings.TrimSpace(string(r)) == "" {
			if buf == "" {
				continue
			}
			tokens = append(tokens, buf)
			buf = ""
			continue
		}

		buf += string(r)

		if pos > len(s) {
			break
		}
	}

	return tokens, errors.New("tokenise error")
}
