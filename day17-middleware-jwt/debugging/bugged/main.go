// Day 17 debugging challenge — BUGGED version.
//
// A tiny stdlib "JWT-like" token: base64url(payload) + "." + base64url(sig)
// where sig = HMAC-SHA256(payloadB64, secret). The auth middleware verifies the
// signature with a constant-time compare (good) but FORGETS to check the `exp`
// claim — so an EXPIRED token is accepted. No gin, no golang-jwt: stdlib only,
// builds and runs offline. Deterministic demo via net/http/httptest.
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

// payload is our minimal claims set. `Exp` is a Unix timestamp.
type payload struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

// mintToken builds payloadB64.sigB64 signed with HMAC-SHA256.
func mintToken(p payload) string {
	body, _ := json.Marshal(p)
	bodyB64 := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(bodyB64))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return bodyB64 + "." + sigB64
}

// verify parses the token and checks the signature, returning the claims.
// BUG: it never checks p.Exp — an expired token passes verification.
func verify(token string) (*payload, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, false
	}
	bodyB64, sigB64 := parts[0], parts[1]

	// Signature check (constant-time — this part is correct).
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

	// BUG: signature is valid, so we accept the token...
	// ...but we never verify that p.Exp is still in the future.
	return &p, true
}

// authMiddleware extracts the bearer token and verifies it.
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
	// Mint a token that expired an hour ago.
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
