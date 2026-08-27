package api

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type Page struct {
	Limit  int
	Cursor string
}

func ParsePage(limit string, cursor string) Page {
	n, _ := strconv.Atoi(limit)
	if n < 1 {
		n = 50
	}
	if n > 500 {
		n = 500
	}
	return Page{Limit: n, Cursor: cursor}
}
func EncodeCursor(value string) string { return base64.RawURLEncoding.EncodeToString([]byte(value)) }
func DecodeCursor(value string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("invalid cursor: %w", err)
	}
	return string(b), nil
}
func Fields(value string, allowed map[string]bool) []string {
	out := []string{}
	for _, f := range strings.Split(value, ",") {
		f = strings.TrimSpace(f)
		if allowed[f] {
			out = append(out, f)
		}
	}
	return out
}
