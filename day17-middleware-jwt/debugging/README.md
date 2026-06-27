# Debugging Challenge — The Token That Never Dies

An auth middleware verifies a signed token's signature perfectly — the crypto is
correct, the HMAC matches, `hmac.Equal` is used. And yet an **expired** token
sails straight through to the protected handler. The code compiles, runs, and
guards nothing. This is the signature-is-not-enough gotcha of Day 17.

To stay offline and stdlib-only, we simulate a JWT: a token is
`base64url(payload).base64url(HMAC-SHA256(payloadB64, secret))`. The payload
carries an `exp` (Unix expiry) claim — exactly like a real JWT.

## Symptom

A request carrying a token that expired an hour ago should get `401`. Instead
the bugged middleware returns `200` and serves the profile. An attacker who
captured an old token keeps access forever.

## Repro

Bugged (wrong output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day17-middleware-jwt/debugging/bugged
go run .
```

Expected (buggy) output:

```
status=200 (expired token accepted!)
```

Fixed (correct output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day17-middleware-jwt/debugging/fixed
go run .
```

Expected (correct) output:

```
status=401 (expired token rejected)
```

## Hint

The signature check is not the bug — `hmac.Equal` is used correctly and the
HMAC genuinely matches. Ask a different question: *what does a valid signature
actually prove?* It proves the payload wasn't tampered with. It says nothing
about **when** the token was issued or whether it's still in its validity
window. Read the claims back out of the payload after verifying the signature
and see which one nobody looks at.

<details>
<summary>Solution & why</summary>

A valid signature is **necessary but not sufficient**. It proves the token was
minted by someone holding the secret and hasn't been altered. It does *not*
prove the token is still current. The bugged `verify` checks the signature, then
returns the claims — and never inspects `exp`:

```go
// BUG: signature valid → accept. Expiry is never checked.
var p payload
if err := json.Unmarshal(body, &p); err != nil {
    return nil, false
}
return &p, true // p.Exp could be in the past — we don't care
```

The fix adds an expiry gate after the signature passes:

```go
// FIX: a valid signature is necessary but not sufficient.
if p.Exp == 0 || time.Now().Unix() >= p.Exp {
    return nil, false // expired (or no expiry set at all) → reject
}
return &p, true
```

**Why this is a security hole, not a nitpick:** expiry is your *only* built-in
defense against a leaked or stolen token in stateless JWT. There's no server-
side session to delete. If you skip `exp`, every token that ever leaks — from a
log, a proxy cache, a browser history — grants permanent access. The whole point
of short-lived access tokens (15 min) plus refresh tokens is that the access
token *dies on its own*. Forget the `exp` check and that entire model collapses.

Note we also reject `exp == 0` (a token with no expiry). A JWT without `exp` is
valid forever — treat a missing expiry as invalid, not as "skip the check."

**Rules of thumb:**

- Always validate `exp` (and `nbf`/`iat` if you set them). A good signature does
  not mean a usable token.
- Use `hmac.Equal` for the signature compare — constant-time, immune to timing
  attacks. Never `==` on the raw signature bytes/string.
- Verify the signing method before trusting the token (pin HMAC; reject
  `alg:none`) so an attacker can't strip the signature entirely.
- Treat a missing `exp` as a failure, not a free pass.

</details>
