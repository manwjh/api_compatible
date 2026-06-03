package redact

import (
	"net/http"
)

func Headers(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, vv := range h {
		if http.CanonicalHeaderKey(k) == "Authorization" {
			continue
		}
		cp := make([]string, len(vv))
		copy(cp, vv)
		out[k] = cp
	}
	return out
}

// Body applies minimal R4 redaction on stored payloads (skeleton: pass-through).
func Body(b []byte) []byte {
	return b
}
