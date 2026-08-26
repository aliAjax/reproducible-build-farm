package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type Digest string

func Of(data []byte) Digest     { sum := sha256.Sum256(data); return Digest(hex.EncodeToString(sum[:])) }
func OfString(s string) Digest  { return Of([]byte(s)) }
func (d Digest) String() string { return string(d) }
func (d Digest) Valid() bool {
	if len(d) != 64 {
		return false
	}
	_, err := hex.DecodeString(string(d))
	return err == nil
}

type Entry struct {
	Name   string
	Digest Digest
	Size   int64
}

func Merkle(entries []Entry) Digest {
	copyEntries := append([]Entry(nil), entries...)
	sort.Slice(copyEntries, func(i, j int) bool { return copyEntries[i].Name < copyEntries[j].Name })
	h := sha256.New()
	for _, e := range copyEntries {
		fmt.Fprintf(h, "%s:%s:%d\\n", e.Name, e.Digest, e.Size)
	}
	return Digest(hex.EncodeToString(h.Sum(nil)))
}
