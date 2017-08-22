package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

var debugEnabled = true

func debug(msg string, args ...interface{}) {
	if debugEnabled {
		log.Printf(msg, args...)
	}
}

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

type matchResult struct {
	match    string
	children map[string]*matchResult
}

func (g grammar) Match(s string) (int, *matchResult) {
	n, match := g.top.match(s, 0)
	debug("grammar->Match [n=%d, match=%+v]", n, spew.Sdump(match))
	if match != nil {
		match = match.children["TOP"]
	}
	return n, match
}

type matcher struct {
	name  string
	typ   int
	atoms []*matchAtom
}

func (t matcher) match(s string, offset int) (n int, result *matchResult) {
	debug("matcher->Match [offset=%d]", offset)
	switch t.typ {
	case matchTypeRule:
		debug("matcher->Match rule")
		mR := matchResult{
			children: make(map[string]*matchResult),
		}
		o := offset
		for _, a := range t.atoms {
			n, m := a.match(s, o)
			if m == nil {
				return 0, nil
			}
			o += n
			mR.children[t.name] = m
		}
		debug("matcher->Match result [offset=%d, o=%d]", offset, o)
		mR.match = s[offset : o-offset]
		return o, &mR
	case matchTypeToken:
		debug("matcher->Match token")
		mR := matchResult{
			children: make(map[string]*matchResult),
		}
		o := offset
		for _, a := range t.atoms {
			n, m := a.match(s, o)
			if m == nil {
				return 0, nil
			}
			o += n
			mR.children[t.name] = m
		}
		debug("matcher->Match result [offset=%d, o=%d]", offset, o)
		mR.match = s[offset : o-offset]
		return o, &mR
	default:
		debug("unknown matcher type")
	}
	return 0, nil
}

func (t *matcher) init(g *grammar) error {
	for _, a := range t.atoms {
		if _, err := a.init(g); err != nil {
			return err
		}
	}
	return nil
}

type matchAtom struct {
	atom  string
	ref   *matcher
	exact bool
}

func (t *matchAtom) match(s string, offset int) (n int, result *matchResult) {
	switch {
	case t.exact:
		if len(s) < offset+len(t.atom) {
			return 0, nil
		}
		if s[offset:len(t.atom)] == t.atom {
			return len(t.atom), &matchResult{match: t.atom}
		}
		return 0, nil
	case t.ref != nil:
		return t.ref.match(s, offset)
	default:
		// TODO
	}

	return 0, nil
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

	if strings.HasPrefix(t.atom, "'") &&
		strings.HasSuffix(t.atom, "'") {
		t.exact = true
		t.atom = t.atom[1 : len(t.atom)-1]
	}

	return nil, nil
}

type parsed map[string]*grammar

func parse(tokens []string) (parsed parsed, err error) {
	debug("parse")
	parsed = make(map[string]*grammar)
	var gram *grammar
	var tokOrRule *matcher
	var ctx int

	var state = expectingGrammar

	for _, tok := range tokens {
		debug("parse [token=%s]", tok)

		switch state {
		case expectingGrammar:
			debug("parse -> expecting grammar")
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
			debug("parse -> expecting label")
			switch ctx {
			case ctxGrammar:
				gram.name = tok
			case ctxmatcher:
				tokOrRule.name = tok
			}
			state = expectingOpenCurly
			continue
		case expectingOpenCurly:
			debug("parse -> expecting open curly")
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
			debug("parse -> expecting matcher")
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
			debug("parse -> expecting matcher body")
			switch tok {
			case "}":
				state = expectingMatch
				gram.entries[tokOrRule.name] = tokOrRule
				tokOrRule = &matcher{}
			default:
				tokOrRule.atoms = append(tokOrRule.atoms, &matchAtom{atom: tok})
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

	defer func() {
		if len(buf) > 0 {
			tokens = append(tokens, buf)
		}
	}()

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
