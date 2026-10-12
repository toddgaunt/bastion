# Bugs

Non-security defects found during the 2026-10-11 review (commit `e15e4c6`). Security findings are in [claude-security-scan-2026.10.11.md](claude-security-scan-2026.10.11.md); where a bug also has a security angle, it's cross-referenced.

✅ = reproduced against the running binary · 🔍 = found by code review / static analysis

## Handlers

| ID | | Location | Bug | Fix |
|----|---|----------|-----|-----|
| B-1 | ✅ | [articles.go:155-161](internal/handlers/articles.go#L155-L161) | `UpdateDocument` builds an error from `Store.Update` but never returns it, so it always replies `200 OK`. Reproduced for a read-only file and for a non-existent article. (Sec L-8) | `return errors.Note{...}.Wrap(err)`; map "article does not exist" to 404. |
| B-2 | ✅ | [auth.go:63-68](internal/handlers/auth.go#L63-L68) | Expired or not-yet-valid tokens produce `500 Internal Server Error`, because `errors.Note{}` has no status. | Use `StatusCode: http.StatusUnauthorized`. |
| B-3 | ✅ | [auth.go:54-55](internal/handlers/auth.go#L54-L55) | `Authorize` treats the entire `Authorization` header as the JWT, so standard `Authorization: Bearer <jwt>` is rejected with 401. | `strings.CutPrefix(h, "Bearer ")` and accept that form. |
| B-4 | 🔍 | [auth.go:115-116](internal/handlers/auth.go#L115-L116), [auth.go:173-176](internal/handlers/auth.go#L173-L176) | On an encode failure the handler calls `enc.Encode(resp)` **again**, then `handleError` writes a second status and body after `WriteHeader(200)` was already sent (superfluous `WriteHeader`, corrupt response). | Encode into a buffer first; on error, return the original `err`. |
| B-5 | 🔍 | [auth.go:94-109](internal/handlers/auth.go#L94-L109) | `Login` duplicates `generateTokens` line for line. | Call `env.generateTokens(claims)`. |
| B-6 | ✅ | [articles.go:108](internal/handlers/articles.go#L108) | Raw documents are served as `Content-Type: text`, which isn't a valid MIME type. HTML responses have no charset, and the problems page sets no Content-Type. (Sec L-2) | `text/plain; charset=utf-8` / `text/html; charset=utf-8`; use `Header().Set`. |
| B-7 | 🔍 | [articles.go:116](internal/handlers/articles.go#L116), [index.go:26](internal/handlers/index.go#L26), [problems.go:56](internal/handlers/problems.go#L56) | `template.Execute` errors are ignored, so a failed render returns a 200 with a partial or empty page. | Check the error and return `statusInternal.Wrap(err)`. |
| B-8 | 🔍 | [packages.go](internal/handlers/packages.go) | `GetPackage` is an unfinished stub with empty branches (staticcheck SA9003) and is never routed. | Finish or delete it. |

## Content parsing / articles

| ID | | Location | Bug | Fix |
|----|---|----------|-----|-----|
| B-9 | ✅ | [article.go:131-139](internal/content/article.go#L131-L139) | `article.Err` gets overwritten by later steps. An invalid `Pinned` error is wiped out by `toBool(unlisted)`, and a bad `Created`/`Updated` date is wiped out by `GenerateHTML`. Reproduced: `tests/bad-date` (`Updated: not-a-date`) is served with **200** instead of an error. | Return early after each step that sets `Err`, or collect errors with `errors.Join`. |
| B-10 | 🔍 | [article.go:58-73](internal/content/article.go#L58-L73) | `SetTimestamps` assigns the zero time even when parsing fails, and only the first `Updated` value is used, although [docs/documents.md](docs/documents.md) says `Updated` may repeat. | Only assign on success; use `Properties.Values("Updated")` and take the latest. |
| B-11 | ✅ | [document.go:55-66](internal/content/document.go#L55-L66) | The delimiter regex `===.*===` isn't anchored to a line, so `===` inside any header value splits the document in the wrong place. The regex is also recompiled on every call. (Sec M-1) | Match a whole line, e.g. `(?m)^===\s*(\S+)\s*===\s*$`; compile it once at package level. |
| B-12 | 🔍 | [document.go:138-165](internal/content/document.go#L138-L165) | An unknown format (e.g. `=== md ===`, typo) silently renders an empty body. | Return an error for unsupported formats. |
| B-13 | 🔍 | [document.go:92](internal/content/document.go#L92) | `Errorf("failed write property %s: %s: err", k, v, err)` has 2 verbs for 3 args, so the message ends in `%!(EXTRA …)` and the cause is lost. (`go vet` misses it because `Errorf` is a package var.) | `"%s: %s: %w"`. |
| B-14 | 🔍 | [document.go:175-181](internal/content/document.go#L175-L181) | `Properties.Add` looks keys up lower-cased but stores them under the key as given, and its first assignment is dead. `Value`/`Has` always lower-case, so a `Properties` built in code with `"Title"` (as in `document_test.go`) can't be read back with `Value("Title")`. | Normalize once: `k := strings.ToLower(key); p[k] = append(p[k], value)`. |
| B-15 | ✅ | [document.go:81-95](internal/content/document.go#L81-L95) | `MarshalDocument` writes lower-cased keys, so `GET /x.md` shows `title:` / `password:` rather than the author's original casing. | Store a canonical display casing (e.g. `textproto.CanonicalMIMEHeaderKey`). |
| B-16 | 🔍 | [document.go:131](internal/content/document.go#L131) | `Tag: math` (lower-case) doesn't enable MathJax, because the value match is case-sensitive. | `strings.EqualFold`. |

## Watcher / store

| ID | | Location | Bug | Fix |
|----|---|----------|-----|-----|
| B-17 | ✅ | [watcher.go:214-217](internal/content/watcher/watcher.go#L214-L217) | A directory moved or copied into `content/` with files already in it gets watched, but its existing files are never indexed (reproduced: 404) until a restart. | On directory `Create`, walk it the same way `watchArticles` does. |
| B-18 | ✅ | [watcher.go:194-200](internal/content/watcher/watcher.go#L194-L200) | Moving a directory out of `content/` leaves all of its articles in the map (reproduced). Runtime-created dotfiles are indexed even though startup skips them. (Sec M-5) | Delete by prefix; apply the hidden-name filter in the event loop. |
| B-19 | 🔍 | [watcher.go:182](internal/content/watcher/watcher.go#L182), [watcher.go:230-235](internal/content/watcher/watcher.go#L230-L235) | If fsnotify closes its channels, `<-watcher.Events` / `<-watcher.Errors` return immediately forever, a 100 % CPU busy loop that floods the log with "watcher error". | Check `ok` on both receives and exit the loop when a channel is closed. |
| B-20 | 🔍 | [watcher.go:126](internal/content/watcher/watcher.go#L126), [watcher.go:216](internal/content/watcher/watcher.go#L216) | Errors from `watcher.Add` are ignored. Hitting `fs.inotify.max_user_watches` means directories silently stop updating. | Log at Error, or fail at startup. |
| B-21 | 🔍 | [watcher.go:69-91](internal/content/watcher/watcher.go#L69-L91) (also `scanner.go`) | The list is sorted twice with the unstable `sort.Slice`, so the alphabetical tiebreak for articles with the same date is not preserved. | One `slices.SortFunc` that compares `Created` and then `Title`. |
| B-22 | ✅ | [watcher.go:69-91](internal/content/watcher/watcher.go#L69-L91) | The index lists articles that are in an error state (e.g. `tests/auth-nopassword`, which then 500s). Articles without `Created` show as `0001-01-01`. | Skip `Err != nil` in `GetAll`; hide or omit the zero date in the template. |
| B-23 | 🔍 | [watcher.go:38](internal/content/watcher/watcher.go#L38), [watcher.go:128](internal/content/watcher/watcher.go#L128) | Every non-hidden file under `content/` (images, `.txt`, backups) is parsed as a document and stored, but `Get` only resolves `*.md`. That wastes memory, and the init log is full of parse errors. | Only index `*.md` files. |
| B-24 | 🔍 | [watcher.go:153-159](internal/content/watcher/watcher.go#L153-L159) | `wg.Add(1)` runs before an early `return`, which would leave the WaitGroup unbalanced if `Fatal` is ever changed to an error return. | Call `wg.Add(1)` right before `go func()`. |
| B-25 | 🔍 | [scanner/scanner.go](internal/content/scanner/scanner.go) | Dead package: unused, keys articles differently from the watcher, and ignores `Username`/`Password`/`Unlisted`. Its log line uses the literal `"route", "err"`. (Sec L-10) | Delete it, along with `config.Content.ScanInterval`. |

## Server / main

| ID | | Location | Bug | Fix |
|----|---|----------|-----|-----|
| B-26 | 🔍 | [main.go:94](cmd/bastion/main.go#L94), [main.go:168-171](cmd/bastion/main.go#L168-L171) | `done` is a nil channel and `close(done)` would panic. That line is unreachable because `logger.Fatal` exits, so there is no graceful shutdown at all. | `done := make(chan bool)`; use `http.Server.Shutdown` on SIGINT/SIGTERM. |
| B-27 | 🔍 | [main.go:83-91](cmd/bastion/main.go#L83-L91) | `isFlagPassed` is unused (staticcheck U1000). | Delete it. |
| B-28 | ✅ | [router.go:21-23](cmd/bastion/router.go#L21-L23) | `/favicon.ico` 301-redirects to `/.static/favicon.ico`, which doesn't exist in the example site (404 after a permanently cached redirect). | Ship a favicon, or use a non-permanent redirect / remove the route. |

## Errors / logging

| ID | | Location | Bug | Fix |
|----|---|----------|-----|-----|
| B-30 | 🔍 | [errors.go:311-333](internal/errors/errors.go#L311-L333) | `Fields()` merge: `fields` is empty when the first loop runs, so its collision branch can never fire. | Simplify. |
| B-31 | 🔍 | [log/recorder.go:33-42](internal/log/recorder.go#L33-L42) | `recorder.With` creates `&recorder{}` with a nil `keyValues` map, so it **panics** on the first assignment. It also drops the `entries` pointer, so `Print` on the child would nil-deref. | `n := &recorder{keyValues: map[any]any{}, entries: r.entries}`. |
| B-32 | 🔍 | [log/log.go:27-34](internal/log/log.go#L27-L34) | The error from `zap.NewProduction()` is discarded (a nil logger would panic later). `defer logger.Sync()` inside the constructor does nothing useful. | Handle the error; call `Sync` at shutdown. |
| B-33 | 🔍 | [auth/auth.go:30-37](internal/auth/auth.go#L30-L37), [auth/jwt.go:73-79](internal/auth/jwt.go#L73-L79) | A short read with a nil error returns `nil` error with a partially filled key. That's unreachable with `crypto/rand`, but the check is wrong. | `if _, err := io.ReadFull(rand.Reader, key); err != nil`. |

## Tests

| ID | | Location | Bug |
|----|---|----------|-----|
| B-34 | 🔍 | [auth_test.go:268-283](internal/handlers/auth_test.go#L268-L283), [document_test.go:65-66](internal/content/document_test.go#L65-L66) | `TestToken` and `TestMarshalUnmarshalDocument` are empty and always pass. |
| B-35 | 🔍 | — | No tests for `watcher`, `GetArticle` (incl. per-article auth), `UpdateDocument`, or the parser edge cases above (B-9, B-11, B-12). |
| B-36 | 🔍 | [auth_test.go:18](internal/handlers/auth_test.go#L18), [errors.go:13-14](internal/handlers/errors.go#L13-L14), [watcher.go:142](internal/content/watcher/watcher.go#L142), [articles.go:134](internal/handlers/articles.go#L134), [index.go:13](internal/handlers/index.go#L13) | Unused vars, funcs and consts (staticcheck U1000). |
