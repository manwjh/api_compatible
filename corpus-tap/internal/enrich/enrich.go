package enrich

import (
	"net/http"
	"strings"
)

type Subject struct {
	UserID  int
	TokenID int
	OK      bool
}

// Resolve maps platform Bearer token to New API user_id.
// Skeleton: CORPUS_TAP_DEV_USER_ID or deny list only; MySQL lookup is TODO.
func Resolve(r *http.Request, devUserID int, denyUsers map[int]struct{}) Subject {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return Subject{}
	}
	uid := devUserID
	if uid <= 0 {
		return Subject{}
	}
	if _, denied := denyUsers[uid]; denied {
		return Subject{}
	}
	return Subject{UserID: uid, OK: true}
}

func SessionKey(r *http.Request, tapRequestID string) string {
	if s := strings.TrimSpace(r.Header.Get("X-Corpus-Session-Id")); s != "" {
		return s
	}
	return tapRequestID
}
