package main

import "fmt"

func G(name string, inner ...func() *matcher) (*grammar, error) {
	g := &grammar{name: name}
	for _, i := range inner {
		in := i()
		g.entries[in.name] = in
	}
	t, ok := g.entries["TOP"]
	if !ok {
		err := fmt.Errorf("expecting TOP in grammar %s", name)
		return g, err
	}
	g.top = t
	return g, nil
}

func T(name string, args ...string) func() *matcher {
	return func() *matcher {
		panic("TODO")
		return nil
	}
}

func R(name string, args ...string) func() *matcher {
	return func() *matcher {
		panic("TODO")
		return nil
	}
}
