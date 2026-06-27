// Day 17 debugging challenge — FIXED version.
//
// Same stdlib "JWT-like" token as bugged/, but verify() now checks the `exp`
// claim after the signature. An expired token is rejected with 401. Stdlib
// only; deterministic demo via net/http/httptest.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var secret = []byte("super-secret-signing-key-32-bytes!!")

type payload struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func mintToken(p payload) string {
	body, _ := json.Marshal(p)
	bodyB64 := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(bodyB64))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return bodyB64 + "." + sigB64
}

// verify checks the signature AND the expiry.
func verify(token string) (*payload, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, false
	}
	bodyB64, sigB64 := parts[0], parts[1]

	// Signature check — constant-time.
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(bodyB64))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil || !hmac.Equal(got, want) {
		return nil, false
	}

	body, err := base64.RawURLEncoding.DecodeString(bodyB64)
	if err != nil {
		return nil, false
	}
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, false
	}

	// FIX: a valid signature is necessary but not sufficient.
	// Reject the token if it has expired.
	if p.Exp == 0 || time.Now().Unix() >= p.Exp {
		return nil, false
	}

	return &p, true
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		if _, ok := verify(token); !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "welcome to your profile")
}

func main() {
	expired := mintToken(payload{
		Sub:  "user-123",
		Role: "admin",
		Exp:  time.Now().Add(-1 * time.Hour).Unix(),
	})

	handler := authMiddleware(http.HandlerFunc(profileHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		fmt.Printf("status=%d (expired token accepted!)\n", rec.Code)
	} else {
		fmt.Printf("status=%d (expired token rejected)\n", rec.Code)
	}
}
