# Day 17 — Middleware & JWT Resources

Curated, no fluff. Read the OWASP cheat sheet and the JWT RFC at least once —
they pay for themselves the first time you review auth code.

## JWT fundamentals

- [jwt.io](https://jwt.io/) — paste a token, see the decoded header/payload live. The fastest way to *prove to yourself* the payload isn't encrypted.
- [RFC 7519 — JSON Web Token](https://datatracker.ietf.org/doc/html/rfc7519) — the spec. The `exp`, `iat`, `iss`, `sub` registered claims are defined here.
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html) — the security checklist: alg pinning, expiry, storage, revocation. Java-titled but the guidance is language-agnostic.

## Go libraries & stdlib

- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) — the Go JWT library. Read the v5 docs for `ParseWithClaims`, `RegisteredClaims`, and `WithValidMethods`.
- [crypto/hmac](https://pkg.go.dev/crypto/hmac) — `hmac.New` and the all-important `hmac.Equal` for constant-time comparison.
- [go-playground/validator](https://github.com/go-playground/validator) — the validation library behind both `binding` and `validate` tags; the README lists every tag.
- [Gin custom middleware](https://gin-gonic.com/docs/examples/custom-middleware/) — the official `c.Next()` / `c.Set()` middleware example.

## Going deeper

- [Alex Edwards — Let's Go Further](https://lets-go-further.alexedwards.net/) — the definitive Go book for building real APIs: middleware, authentication, permissions/RBAC, all in idiomatic stdlib-first Go.
