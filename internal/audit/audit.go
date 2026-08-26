package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type Event struct {
	ID           string            `json:"id"`
	Actor        string            `json:"actor"`
	Action       string            `json:"action"`
	Resource     string            `json:"resource"`
	Metadata     map[string]string `json:"metadata"`
	At           time.Time         `json:"at"`
	PreviousHash string            `json:"previous_hash"`
	Hash         string            `json:"hash"`
}
type Log struct {
	mu     sync.Mutex
	events []Event
	last   string
}

func New() *Log { return &Log{events: []Event{}} }
func (l *Log) Append(actor, action, resource string, meta map[string]string) Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{ID: time.Now().UTC().Format("20060102T150405.000000000"), Actor: actor, Action: action, Resource: resource, Metadata: meta, At: time.Now().UTC(), PreviousHash: l.last}
	b, _ := json.Marshal(e)
	sum := sha256.Sum256(append([]byte(l.last), b...))
	e.Hash = hex.EncodeToString(sum[:])
	l.last = e.Hash
	l.events = append(l.events, e)
	return e
}
func (l *Log) List() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.events
}
func Verify(events []Event) bool {
	prev := ""
	for _, e := range events {
		if e.PreviousHash != prev {
			return false
		}
		copyE := e
		copyE.Hash = ""
		b, _ := json.Marshal(copyE)
		sum := sha256.Sum256(append([]byte(prev), b...))
		if hex.EncodeToString(sum[:]) != e.Hash {
			return false
		}
		prev = e.Hash
	}
	return true
}
