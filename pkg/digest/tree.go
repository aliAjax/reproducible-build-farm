package digest

import (
	"sort"
	"strings"
)

type Tree struct{ Entries []Entry }

func NewTree(entries []Entry) Tree {
	out := append([]Entry(nil), entries...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return Tree{Entries: out}
}
func (t Tree) Root() Digest { return Merkle(t.Entries) }
func (t Tree) Lookup(name string) (Entry, bool) {
	for _, e := range t.Entries {
		if e.Name == name {
			return e, true
		}
	}
	return Entry{}, false
}
func (t Tree) Prefix(prefix string) Tree {
	out := []Entry{}
	for _, e := range t.Entries {
		if strings.HasPrefix(e.Name, prefix) {
			out = append(out, e)
		}
	}
	return Tree{Entries: out}
}
func (t Tree) Size() int64 {
	var n int64
	for _, e := range t.Entries {
		n += e.Size
	}
	return n
}
