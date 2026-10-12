# Bastion Security Scan — 2026-10-11

| | |
|---|---|
| **Target** | `github.com/toddgaunt/bastion` @ `e15e4c6` (master) |
| **Scope** | All Go sources (`cmd/`, `internal/`), templates, example site, build/version scripts, CI workflow, dependencies |
| **Method** | Full manual review + static tooling + live dynamic testing of the built binary + fuzzing |
| **Toolchain** | go1.26.9 (go.mod declares `go 1.25.4`) |

Non-security defects found during the same review are tracked separately in [Bugs.md](Bugs.md).

---

## 1. Executive summary

Bastion is a small (~3.5k LOC) markdown-to-HTML web server with per-article HTTP Basic auth and a JWT-protected admin `POST` endpoint for updating documents. The core design is fairly sound: article lookups are map-based (no path traversal), templates use `html/template`, JWT verification pins `HS256` with a random 256-bit per-process key, refresh tokens are single-use, and problem+json responses don't leak internal errors.

The significant risks are in **transport security, content parsing, authentication hardening, and resource limits**:

| ID | Severity | Title | Verified |
|----|----------|-------|----------|
| H-1 | **High** | TLS silently fails open to plaintext HTTP; TLS is off by default | ✅ live |
| M-1 | **Medium** | Protected-article auth bypass + password disclosure when a header value contains `=== x ===` | ✅ live |
| M-2 | **Medium** | Unauthenticated memory exhaustion via unbounded body on `POST /.auth/token` | ✅ live (300 MB req → 2 GB RSS peak) |
| M-3 | **Medium** | No HTTP server timeouts or header/body limits (Slowloris, connection exhaustion) | ✅ code + gosec G114 |
| M-4 | **Medium** | Username enumeration via timing (~130×), no brute-force throttling, bcrypt CPU DoS | ✅ live |
| M-5 | **Medium** | Content watcher exposes "hidden" files/dirs created at runtime and keeps serving content moved out of `content/` | ✅ live |
| L-1 | Low | Outdated `gomarkdown` with 3 reachable advisories; watcher goroutine has no panic recovery (process-wide crash) | ⚠️ govulncheck; fuzz found no crash |
| L-2 | Low | Missing security headers; invalid `Content-Type: text` on raw `.md` enables MIME sniffing | ✅ live |
| L-3 | Low | Directory listing enabled on `/.static/` | ✅ live |
| L-4 | Low | Unsanitized HTML/JS in content (raw HTML, `javascript:` links, `html` format) | ✅ live |
| L-5 | Low | Third-party MathJax script loaded from CDN with floating version and no SRI | ✅ code |
| L-6 | Low | Refresh tokens and claims written to logs | ✅ live |
| L-7 | Low | Token store never purged; no revocation/logout; token responses lack `Cache-Control: no-store` | ✅ code/live |
| L-8 | Low | `POST` update reports success on failure; writes with mode `0755`; follows symlinks | ✅ live |
| L-9 | Low | Protected article titles/routes are listed publicly on the index | ✅ live |
| L-10 | Low | Latent fail-open code paths (`scanner` package ignores auth/unlisted; `Admin` claim never checked) | ✅ code |
| L-11 | Low | Weak sample credentials committed in `env.sh`; plaintext article passwords (documented) | ✅ code |
| I-1 | Info | Dependency hygiene: deprecated `chi` v1 (advisory not reachable), `x/crypto` module advisories (not reachable) | govulncheck |
| I-2 | Info | CI hardening: unpinned/outdated actions, no `permissions:`, no vuln scanning, Go version mismatch | ✅ code |

---

## 2. Methodology

1. **Manual review** of every tracked source file, template, script and config.
2. **Static tools**
   - `go vet ./...` — clean
   - `go test -race ./...` — all pass, no races detected (note: watcher/scanner/main have no tests)
   - `govulncheck@v1.9.0` (vuln DB as of 2026-10-11) — 4 reachable symbol-level findings, 4 module-only findings
   - `gosec` (latest) — 23 issues (G114, G203, G304, G306, G117, G122, G104), triaged below
   - `staticcheck@v0.8.1 -checks all` — dead code / style only (fed into Bugs.md)
3. **Dynamic testing**: built the binary, ran it against a scratch copy of `www.example.com/` with auth enabled (`BASTION_USERNAME/PASSWORD` set), and exercised every route with `curl` — auth, token lifecycle, document updates, the file watcher, TLS config handling, large bodies, XSS payloads, and parser edge cases.
4. **Fuzzing**: 90 s / 5.25 M executions of `markdown.ToHTML` at the pinned gomarkdown version with Bastion's exact parser extensions — no crash or hang found.

All test artifacts lived in a scratch directory outside the repository; the working tree is unchanged apart from these two report files.

---

## 3. Attack surface

| Route | Auth | Handler |
|-------|------|---------|
| `GET /` | none | `Index` — lists non-unlisted articles |
| `GET /*` | per-article Basic auth if `Username`/`Password` in header | `GetArticle` (HTML) |
| `GET /*.md` | same as above | `GetArticle` (raw canonical document) |
| `POST /*` | JWT in `Authorization` header | `UpdateDocument` — overwrites document on disk |
| `GET /.auth/login` | Basic auth (server credentials) | issues access JWT (5 s) + refresh token (24 h session) |
| `POST /.auth/token` | refresh token in JSON body | rotates tokens |
| `GET /.problems/{id}` | none | static problem docs |
| `GET /.static/*` | none | `http.FileServer` on `<site>/static` |

Trust boundaries: anonymous internet clients; article readers holding per-article credentials; the single admin (server credentials); local content authors with write access to `<site>/content`.

---

## 4. Findings

### H-1 — TLS fails open to plaintext HTTP; TLS disabled by default

**Location:** [cmd/bastion/main.go:157-166](cmd/bastion/main.go#L157-L166), [cmd/bastion/config.go:29-33](cmd/bastion/config.go#L29-L33), [www.example.com/config.json](www.example.com/config.json)

```go
if !config.Network.TLS.Disable && (config.Network.TLS.Cert != "" && config.Network.TLS.Key != "") {
    ... ListenAndServeTLS ...
} else {
    logger.Print(log.Warn, "TLS is disabled")
    ... http.ListenAndServe(addr, r)   // plaintext
}
```

If the operator explicitly sets `"tls": {"disable": false}` but the `cert` or `key` path is empty (typo, missing env templating, etc.), the server **silently serves plaintext HTTP** with only a warning-level log line. The shipped default config and `defaultConfig` also have TLS disabled.

Every credential in this system travels in a header that is trivially recoverable in cleartext: HTTP Basic credentials for protected articles and for `/.auth/login`, admin JWTs in `Authorization`, and refresh tokens in the `/.auth/token` body. A network observer can capture the admin password and then rewrite any article via `POST /*`.

**Reproduction (verified):** config with `"tls":{"disable":false,"cert":"/etc/ssl/missing.pem","key":""}` → log shows `TLS is disabled` / `serving without TLS`, and `curl http://localhost:PORT/` returns 200.

**Recommendation**
- Fail closed: if `disable` is false and either path is empty, `Fatal` at startup.
- Make TLS enabled by default; require an explicit `disable: true` for local development.
- When TLS is on, add HSTS (`Strict-Transport-Security: max-age=31536000`) and optionally an HTTP→HTTPS redirect listener.
- Consider refusing to start with authentication enabled while TLS is disabled, unless an explicit "behind TLS-terminating proxy" option is set.

---

### M-1 — Protected-article auth bypass and password disclosure via `===` in a header value

**Location:** [internal/content/document.go:54-73](internal/content/document.go#L54-L73)

```go
re := regexp.MustCompile(`===.*===`)
index := re.FindIndex(data)            // first match ANYWHERE, not a whole line
properties, err := parseProperties(data[:index[0]])
```

The format delimiter is located with an unanchored regex. If any header value contains `=== … ===` (e.g. `Title: Setext headings use === markdown ===`), the header ends at that point. The `Username`/`Password` lines below it become **body content**, so no `Authenticator` is created and the article is served **without authentication**. Because those lines are now part of the body, the page **also publishes the article's password**.

**Reproduction (verified):**

```
Title: Notes === markdown ===
Username: u
Password: s3cret
=== markdown ===
PROTECTED BODY
```

`GET /delim2` → `200`, body contains `Password: s3cret` and `PROTECTED BODY`. A variant with `Title: a === b === c` renders an empty HTML body (unknown format `b`), but `GET /delim.md` returns the full raw text, password included, unauthenticated.

This requires a content author to write such a header, so it is not directly attacker-triggerable. However, it fails open silently, on a feature whose only purpose is access control.

**Recommendation**
- Parse line by line: the delimiter must be an entire line matching `^===\s*(\S+)\s*===\s*$`.
- Reject unknown formats as an article error instead of rendering an empty body.
- Defensive: if the *body* of a document contains lines like `Password:` before any delimiter, or the header is suspiciously short, treat it as a parse error. At minimum, add a test for this case.

---

### M-2 — Unauthenticated memory exhaustion on `POST /.auth/token`

**Location:** [internal/handlers/auth.go:131-132](internal/handlers/auth.go#L131-L132)

```go
dec := json.NewDecoder(r.Body)
err := dec.Decode(&refreshToken)   // no size limit
```

The refresh endpoint is unauthenticated and decodes an arbitrarily large JSON string into memory. The decoder buffers the raw input and then allocates the decoded string.

**Reproduction (verified):** a single 300 MB request (`"AAAA…"`) returned 401, but drove the process's peak RSS (`VmHWM`) to **~2.0 GB**, and RSS stayed at ~1.4 GB afterwards. A handful of concurrent requests will OOM a small VPS. (`UpdateDocument` also uses an unbounded `io.ReadAll`, but only after JWT verification.)

**Recommendation**
- Wrap bodies with `http.MaxBytesReader(w, r.Body, 4<<10)` on `/.auth/token`, plus a sane limit (e.g. 1–5 MB) on `POST /*`.
- Add a global limit via middleware as defence in depth.

---

### M-3 — No server timeouts or connection limits

**Location:** [cmd/bastion/main.go:160](cmd/bastion/main.go#L160), [cmd/bastion/main.go:165](cmd/bastion/main.go#L165) (gosec G114)

`http.ListenAndServe` / `ListenAndServeTLS` use a zero-value `http.Server`, so there is no `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` or `IdleTimeout`. Slowloris-style clients can hold connections and goroutines open indefinitely, which compounds M-2 and M-4.

**Recommendation**

```go
srv := &http.Server{
    Addr:              addr,
    Handler:           r,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       15 * time.Second,
    WriteTimeout:      30 * time.Second,
    IdleTimeout:       60 * time.Second,
    MaxHeaderBytes:    16 << 10,
}
```

Also consider chi's `middleware.Throttle` or `middleware.Timeout`, or deploy behind a reverse proxy that enforces these limits.

---

### M-4 — Username enumeration by timing; no brute-force protection; bcrypt CPU DoS

**Location:** [internal/auth/simple.go:40-46](internal/auth/simple.go#L40-L46)

```go
if sa.username != username {          // returns immediately
    return Claims{}, errors.New(...)
}
err := bcrypt.CompareHashAndPassword(sa.hash, []byte(password))   // ~50 ms
```

bcrypt only runs when the username matches, which creates a large timing oracle. The same `simple` authenticator protects both `/.auth/login` and every protected article.

**Reproduction (verified):**

| Endpoint | Wrong username | Correct username, wrong password |
|----------|---------------|-------------------------------|
| `GET /authenticate` (article) | ~0.35 ms | ~53 ms |
| `GET /.auth/login` (admin) | ~0.40 ms | ~53 ms |

An attacker can confirm the admin username and per-article usernames, then brute-force passwords. Nothing throttles, locks out or backs off. Article passwords are often weak; the README example is `bananas`. Separately, each request with a valid username burns ~50 ms of CPU, so a modest request rate can saturate the CPU (unauthenticated).

**Recommendation**
- Always run bcrypt: compare the username with `subtle.ConstantTimeCompare`, and call `CompareHashAndPassword` against the real hash (or a dummy hash) regardless of whether the username matched.
- Add per-IP and per-username rate limiting with exponential backoff on `/.auth/login` and on failed article auth (e.g. `golang.org/x/time/rate`, or `httprate`).
- Log failed auth attempts with source IP at `Warn` level so fail2ban-style tooling can act.
- Consider hashing per-article passwords at parse time once (already done), but cache the successful auth result per connection/session so legitimate readers don't pay 50 ms on every page view.

---

### M-5 — Watcher exposes "hidden" content and keeps serving removed directories

**Location:** [internal/content/watcher/watcher.go:178-229](internal/content/watcher/watcher.go#L178-L229)

The initial `filepath.Walk` skips names beginning with `.` ("reserved for built-in routes"), but the fsnotify event loop **does not apply that filter**. It also deletes only the exact path from a `Remove`/`Rename` event, never its children.

**Reproduction (verified):**

1. `content/.draft.md` created while running → `GET /.draft` returns **200** with the draft content.
2. `mkdir content/.private && echo … > content/.private/x.md` while running → `GET /.private/x` returns **200**.
3. `mv content/old2 /elsewhere` (moving a directory out of the tree) → `GET /old2/a` **still returns 200** with the old content, until restart. (Deleting with `rm -rf` is handled, because each child file fires its own event.)

Operators reasonably expect that dot-prefixed drafts are private and that moving a directory out of `content/` unpublishes it. Neither holds while the server is running. The behaviour also differs between startup and runtime, so a restart changes what is public.

**Recommendation**
- Apply the same hidden-name check (to every path component relative to the root) inside the event loop.
- On `Remove`/`Rename`, also delete every key with the prefix `path + "/"`, and `watcher.Remove()` the directory.
- On `Create` of a directory, walk it to index existing files (see Bugs.md B-5).
- Consider periodic full rescans as a safety net.

---

### L-1 — Outdated `gomarkdown` with reachable advisories; no panic recovery in the watcher

**Location:** [go.mod](go.mod) (`gomarkdown v0.0.0-20201030010234`), [internal/content/document.go:164](internal/content/document.go#L164), [internal/content/watcher/watcher.go:178](internal/content/watcher/watcher.go#L178)

govulncheck reports three advisories reachable from `Document.GenerateHTML`:

| Advisory | Issue | Fixed in |
|----------|-------|----------|
| GO-2026-5208 | Out-of-bounds read in `SmartypantsRenderer` (Bastion's `nil` renderer → `CommonFlags` includes Smartypants) | `v0.0.0-20260411013819-759bbc3e3207` |
| GO-2024-3205 | Infinite loop in parser | `v0.0.0-20240729212818-a2a9c4f76ef5` |
| GO-2023-2074 | Parser out-of-bounds read on malformed input | `v0.0.0-20230922105210-14b16010c2ee` |

Markdown is rendered inside the watcher goroutine, which has **no `recover()`**. A panic there terminates the **entire server**, unlike handler panics, which `net/http` recovers per connection. An infinite loop would wedge the watcher and stop all content updates.

Exploitation needs the ability to write content: local filesystem access, or the admin token via `POST /*`. A 90 s / 5.25 M-exec fuzz run of the pinned version with Bastion's parser options found no crash. The advisories should still be treated as real.

**Recommendation**
- `go get github.com/gomarkdown/markdown@latest`
- Wrap per-file article generation in `defer func(){ if r := recover(); r != nil { article.Err = … } }()`.
- Add `middleware.Recoverer` to the router for consistent 500s and logging.

---

### L-2 — Missing security headers; invalid raw `.md` content type

**Location:** [cmd/bastion/router.go](cmd/bastion/router.go), [internal/handlers/articles.go:108](internal/handlers/articles.go#L108), [internal/handlers/problems.go:56-57](internal/handlers/problems.go#L56-L57)

Responses carry no `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options` / `frame-ancestors`, `Referrer-Policy` or HSTS (verified with `curl -D -`).

Raw documents are served with `Content-Type: text`, which is not a valid MIME type. Browsers treat that as missing and **sniff** the body, so an article in `html` format, or one containing `<script>`, can be rendered as HTML at its `.md` URL. The problems page sets no Content-Type at all.

**Recommendation:** add a small header middleware:

```
Content-Security-Policy: default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; object-src 'none'; base-uri 'none'; frame-ancestors 'none'
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
```

Serve raw documents as `text/plain; charset=utf-8` (or `text/markdown; charset=utf-8`), and HTML as `text/html; charset=utf-8`. Use `Header().Set`, not `Add`.

---

### L-3 — Directory listing on `/.static/`

**Location:** [cmd/bastion/main.go:100](cmd/bastion/main.go#L100), [cmd/bastion/router.go:39](cmd/bastion/router.go#L39)

`http.FileServer(http.Dir(...))` lists directories, and serves dotfiles and symlink targets. **Verified:** `GET /.static/` and `GET /.static/sub/` return directory indexes that reveal every file. Anything accidentally dropped into `static/` (backups, `.git`, `.env`) becomes discoverable and downloadable. Path traversal outside `static/` is *not* possible; `http.Dir` handles it correctly.

**Recommendation:** wrap the filesystem to return 404 for directories without `index.html` and for any path component starting with `.`. Alternatively, use `http.FileServerFS` over an `os.Root` (Go 1.24+) to also block symlink escapes.

---

### L-4 — Content HTML/JS is rendered unsanitized

**Location:** [internal/content/document.go:144-164](internal/content/document.go#L144-L164) (gosec G203)

The markdown renderer passes raw HTML through, `javascript:` links are emitted as-is, and `=== html ===` documents are inserted verbatim as `template.HTML`. **Verified:** `<script>alert(3)</script>` and `[x](javascript:alert(4))` in a body are emitted live. Title and Description *are* correctly escaped.

Content authors are trusted by design, so this is low severity today. It does mean that anyone holding a leaked admin token, or anyone given write access to `content/`, gets stored XSS against every visitor, including users of protected articles. The risk becomes significant if content ever comes from less-trusted contributors.

**Recommendation:** decide explicitly whether raw HTML is a feature. If not, enable `html.SkipHTML` and filter unsafe URL schemes; for the `html` format, or generally, run output through `bluemonday.UGCPolicy()`. A CSP (L-2) that disallows inline scripts limits the impact either way.

---

### L-5 — MathJax loaded from CDN without integrity pinning

**Location:** [internal/content/document.go:38-40](internal/content/document.go#L38-L40)

`https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-svg.js` is a floating major version with no `integrity=` attribute. A CDN or npm-package compromise would execute arbitrary script on every `Tag: Math` page.

**Recommendation:** pin an exact version (e.g. `mathjax@3.2.2`) and add `integrity="sha384-…" crossorigin="anonymous"`, or self-host it under `/.static/`.

---

### L-6 — Tokens and claims written to logs

**Location:** [internal/handlers/auth.go:147](internal/handlers/auth.go#L147), [internal/handlers/auth.go:163](internal/handlers/auth.go#L163)

The `Refreshed authentication` log line records the presented refresh token and the full claims. Failed refreshes embed the submitted token string in the error, and that error is logged. **Verified** in server output (`"token":"6FGOkz1ZP_tjMY1sNEj1PQ=="`). The logged *successful* token has already been consumed, which limits its value. Still, logging bearer secrets is a bad pattern: a timing race, a future change to non-rotating tokens, or a mistyped valid token would leak usable secrets to log storage. Passwords were confirmed **not** to appear in logs.

**Recommendation:** never log token values. Log a short hash prefix (e.g. first 8 hex chars of SHA-256) if correlation is needed.

---

### L-7 — Session/token lifecycle gaps

**Location:** [internal/handlers/auth.go:17-37](internal/handlers/auth.go#L17-L37)

- The in-memory `tokens` map is only cleaned when a token is used. Unused refresh tokens, one per login, accumulate forever (slow memory growth; needs valid credentials).
- There is no logout/revocation endpoint. A stolen refresh token stays valid for the rest of its 24 h session.
- `/.auth/login` and `/.auth/token` responses don't set `Cache-Control: no-store`.
- Login is a `GET`. That is acceptable with Basic auth, but `POST` is conventional for credential exchange and avoids accidental prefetch or caching.

**Recommendation:** purge expired entries on a timer or on insert, add `POST /.auth/logout` that deletes the refresh token, and set `Cache-Control: no-store` on token responses.

---

### L-8 — Document update integrity issues

**Location:** [internal/handlers/articles.go:155-161](internal/handlers/articles.go#L155-L161), [internal/content/watcher/watcher.go:58](internal/content/watcher/watcher.go#L58)

- The error from `Store.Update` is wrapped but **never returned**, so the handler always replies `200 OK`. **Verified:** a read-only target file and a non-existent article both return 200 with nothing written. An admin cannot tell whether a change (for example, adding a `Password:` to protect an article) actually took effect. That is a security-relevant false assurance.
- `os.WriteFile(..., 0755)` (gosec G306): the mode applies only on create, but marking content executable is wrong; use `0644` or `0600`.
- The write follows symlinks inside `content/`, so the admin endpoint can overwrite whatever a symlink points at. Writes also aren't atomic, and a reader can observe a truncated file. Use `os.Root` (or `O_NOFOLLOW`) and write-to-temp-then-rename.

---

### L-9 — Protected articles are advertised on the public index

**Location:** [internal/content/watcher/watcher.go:69-91](internal/content/watcher/watcher.go#L69-L91)

`GetAll` filters `Unlisted` but not protected articles. **Verified:** `/` lists `Authenticate` with its route and title. Titles and descriptions of private content can be sensitive by themselves.

**Recommendation:** exclude articles with an `Authenticator` from the index and nav, or make listing protected articles an explicit opt-in.

---

### L-10 — Latent fail-open code paths

- [internal/content/scanner/scanner.go:95-178](internal/content/scanner/scanner.go#L95-L178): the older `Scanner` store builds articles **without** reading `Username`/`Password` or `Unlisted`. If someone wires it back in (it still implements `content.Store`, and `config.ScanInterval` still exists), every protected article becomes public. Delete the package, or make both stores share `content.GenerateArticle`.
- [internal/handlers/auth.go:52-74](internal/handlers/auth.go#L52-L74): `Authorize` never checks `claims.Admin`. It is safe today, because only the server authenticator mints tokens, but any future non-admin token would gain write access. Check `claims.Admin` before `UpdateDocument`.

---

### L-11 — Credential handling hygiene

- [env.sh](env.sh) commits `BASTION_USERNAME=username` / `BASTION_PASSWORD=password`. Anyone who sources it in production gets trivially guessable admin credentials. Rename it to `env.sh.example`, or have it generate a random password.
- Per-article passwords are stored in plaintext in content files (documented in the README), and the raw `.md` endpoint echoes them back to any reader who authenticated. Consider accepting a bcrypt hash (`PasswordHash:`) and stripping auth properties from the raw output.
- `configVariable` `location: "literal"` puts the admin password in `config.json`. Document that `env` is preferred, and check the file's permissions at startup.

---

### I-1 — Dependency notes

| Module | Finding | Reachable? | Action |
|--------|---------|------------|--------|
| `github.com/go-chi/chi v1.5.0` | GO-2026-4316 open redirect in `middleware.RedirectSlashes`; **no fix** for the v1 module | **No**: `RedirectSlashes` isn't used (govulncheck matches the module broadly) | Migrate to `github.com/go-chi/chi/v5` (maintained) |
| `golang.org/x/crypto v0.52.0` | GO-2026-6355, -6354, -6303 (`ssh`), GO-2026-5932 (`openpgp`) | **No**: only `bcrypt` is imported | Bump to ≥ v0.56.0 to keep scanners quiet |
| `github.com/dvsekhvalnov/jose2go v1.7.0` | No open advisories. PBES2 `p2c` DoS path verified unreachable: Bastion passes a `[]byte` key, and PBES2 `Unwrap` requires `string`, so it errors before PBKDF2. `alg: none` is rejected because the key is non-nil. | — | OK. Consider rejecting 5-part (JWE) tokens before calling `DecodeBytes`. |
| `golang.org/x/exp 2023-02` | Only `constraints` used | — | Replace with `cmp.Ordered` from the stdlib and drop the dependency |
| Go stdlib | govulncheck with go1.26.9 found no stdlib issues | — | go.mod pins `go 1.25.4`. Make sure release builds use a current patch toolchain (add a `toolchain` directive). |

### I-2 — CI / supply chain

[.github/workflows/go.yml](.github/workflows/go.yml):
- `actions/checkout@v3` / `setup-go@v3` are outdated (Node 16) and pinned by mutable tag. Pin to full commit SHAs at current majors.
- No `permissions:` block, so the default `GITHUB_TOKEN` scope applies. Add `permissions: contents: read`.
- `go-version: 1.21` while go.mod requires 1.25.4 (works only via toolchain auto-download). Use `go-version-file: go.mod`.
- No `govulncheck` step. Add `golang/govulncheck-action` so findings like L-1 fail the build.
- Dependabot is active (good); consider enabling it for `github-actions` too.

---

## 5. Verified non-issues (controls that held up)

- **Path traversal:** article lookup is a map lookup on a `filepath.Clean`ed key; nothing user-controlled reaches the filesystem. `/.static` uses `http.Dir`, which rejects `..`.
- **JWT:** `HS256` with a 32-byte `crypto/rand` key per process; the `alg` header is checked after verification; `none` and HS384/HS512 variants are rejected; access tokens live 5 s; refresh tokens are 128-bit random, **single-use** (reuse → 401, verified), and bounded to the original 24 h session.
- **CSRF / CORS:** no cookies are used and credentials travel only in headers; no CORS headers are emitted, so cross-origin `POST` with `Authorization` is blocked by preflight.
- **Template injection / reflected XSS:** Title, Description, site name and style go through `html/template` (verified escaped). Problem responses go through `json.Marshal`, which escapes `<`, `>` and `&`. Internal error text is not returned to clients.
- **Incomplete article credentials** (only a username or only a password) fail closed with 500 (verified with `tests/auth-nopassword`, `tests/auth-nousername`).
- **bcrypt** is used for comparison; passwords over 72 bytes are rejected at hash time, so such articles fail closed.
- **Passwords are not logged** (grepped server logs after authenticated and failed requests).
- **Concurrency:** `go test -race` passes; article map and token map access is mutex-protected.

## 6. Triage of tool output not covered above

| Tool / rule | Location | Verdict |
|-------------|----------|---------|
| gosec G304 (file inclusion) | `article.go:96`, `scanner.go:129` | Paths come from a directory walk of the operator's content dir, so this is not attacker-controlled. Symlink TOCTOU (G122) applies only to local authors. |
| gosec G117 (secret in JSON) | `auth.go:115,116,173,176` | Intended: these are the token responses. See L-7 for `no-store`. |
| gosec G104 (unchecked errors) | multiple | Correctness only; tracked in Bugs.md. |
| staticcheck U1000 / SA9003 | dead code | Tracked in Bugs.md. |

## 7. Prioritized remediation plan

1. **Now:** fail closed on TLS misconfiguration (H-1); add `MaxBytesReader` + `http.Server` timeouts (M-2, M-3); fix the delimiter parser (M-1); upgrade gomarkdown and add `recover` (L-1).
2. **Next:** constant-time auth + rate limiting (M-4); watcher hidden-path and subtree removal fixes (M-5); return errors from `UpdateDocument` (L-8); security-headers middleware + correct content types (L-2); disable directory listing (L-3).
3. **Then:** stop logging tokens (L-6), token-store cleanup/logout (L-7), hide protected articles from the index (L-9), delete the `scanner` package and check `Admin` (L-10), MathJax SRI (L-5), CI hardening + govulncheck (I-2), migrate to chi/v5 (I-1).
