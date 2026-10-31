// Package demo serves bundled example dossiers. Demo data is never produced
// for a live audit: it is only ever returned by GET /demo/{id}.
package demo

import (
	_ "embed"
	"encoding/json"
	"errors"
	"regexp"
)

//go:embed demo-cases.json
var raw []byte

// ErrNotFound is returned for an unknown demo id.
var ErrNotFound = errors.New("demo case not found")

// idPattern matches the route contract in api/openapi.yaml.
var idPattern = regexp.MustCompile(`^[a-z0-9-]{4,40}$`)

// Case is one bundled example: the synthetic caption plus its dossier.
type Case struct {
	ID      string          `json:"id"`
	Caption string          `json:"caption"`
	Dossier json.RawMessage `json:"dossier"`
}

type file struct {
	Cases []Case `json:"cases"`
}

var loaded = mustLoad()

func mustLoad() map[string]Case {
	var f file
	if err := json.Unmarshal(raw, &f); err != nil {
		panic("demo: embedded demo-cases.json is invalid: " + err.Error())
	}
	out := make(map[string]Case, len(f.Cases))
	for _, c := range f.Cases {
		if !idPattern.MatchString(c.ID) {
			panic("demo: case id violates the route pattern: " + c.ID)
		}
		out[c.ID] = c
	}
	return out
}

// Get returns the case for id, or ErrNotFound.
func Get(id string) (Case, error) {
	if !idPattern.MatchString(id) {
		return Case{}, ErrNotFound
	}
	c, ok := loaded[id]
	if !ok {
		return Case{}, ErrNotFound
	}
	return c, nil
}

// IDs lists the available demo ids, used by tests and the frontend build check.
func IDs() []string {
	out := make([]string, 0, len(loaded))
	for id := range loaded {
		out = append(out, id)
	}
	return out
}
