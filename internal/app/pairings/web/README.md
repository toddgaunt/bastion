# Pairings

A static, client-side app for creating a random cycle of pairings (gift-exchange
style) and sharing a link so each person can see who they're paired with.
See [docs/specs/pairings-app.md](docs/specs/pairings-app.md).

## Run locally

```sh
./run.sh          # serves on the first free port from 8000 and prints the URL
./run.sh 9000     # or pick a port
```

Requires `python3`. The page must be served over HTTP (not opened as a file)
because it uses ES modules.

## Test

```sh
npm test      # runs node --test (Node 18+)
```

## Layout

- `index.html`, `styles.css` — the page
- `src/core.js` — pure logic: generation, encoding, validation, lookup
- `src/app.js` — DOM wiring
- `test/` — unit tests for `src/core.js`

To deploy, upload `index.html`, `styles.css`, and `src/` to any static host.
