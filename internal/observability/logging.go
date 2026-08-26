package observability

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu     sync.Mutex
	out    *log.Logger
	fields map[string]string
}

func NewLogger() *Logger { return &Logger{out: log.New(os.Stdout, "", 0), fields: map[string]string{}} }
func (l *Logger) With(k, v string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := NewLogger()
	n.out = l.out
	n.fields = l.fields
	n.fields[k] = v
	return n
}
func (l *Logger) Event(level, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	m := map[string]interface{}{"ts": time.Now().UTC().Format(time.RFC3339Nano), "level": level, "msg": msg}
	for k, v := range l.fields {
		m[k] = v
	}
	for k, v := range fields {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	l.out.Print(string(b))
}
func (l *Logger) Info(msg string, fields map[string]interface{})  { l.Event("info", msg, fields) }
func (l *Logger) Error(msg string, fields map[string]interface{}) { l.Event("error", msg, fields) }
