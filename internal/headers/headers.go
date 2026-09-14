package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	line, _, ok := bytes.Cut(data, []byte(crlf))
	if !ok {
		return 0, false, nil
	}

	if len(line) == 0 {
		return 2, true, nil
	}

	rawKey, rawValue, found := bytes.Cut(line, []byte(":"))

	if !found || len(rawKey) == 0 {
		return 0, false, fmt.Errorf("malformed header: %q", line)
	}

	if bytes.ContainsAny(rawKey, " \t") {
		return 0, false, fmt.Errorf("malformed header key: %q", rawKey)
	}

	if !isValidHeaderName(string(rawKey)) {
		return 0, false, fmt.Errorf("not a valid header key: %q", rawKey)
	}

	key := string(rawKey)
	val := string(bytes.TrimSpace(rawValue))

	h.Set(key, val)

	return len(line) + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func isValidHeaderName(key string) bool {
	if len(key) == 0 {
		return false
	}
	for _, r := range key {
		switch {
		case unicode.IsLetter(r) && r <= unicode.MaxASCII: // A-Z, a-z
		case unicode.IsDigit(r) && r <= unicode.MaxASCII: // 0-9
		case strings.ContainsRune("!#$%&'*+-.^_`|~", r):
		default:
			return false
		}
	}
	return true
}
