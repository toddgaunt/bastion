// Pure, DOM-free logic for the Pairings app. See docs/specs/pairings-app.md.

export class InvalidCodeError extends Error {
  constructor(message) {
    super(message);
    this.name = 'InvalidCodeError';
  }
}

// ---------------------------------------------------------------------------
// Names (spec §3.3)

/** Display form: NFC, trimmed, internal whitespace collapsed. */
export function cleanName(raw) {
  return String(raw).normalize('NFC').trim().replace(/\s+/g, ' ');
}

/** Comparison form: display form, lowercased. */
export function normalizeName(raw) {
  return cleanName(raw).toLocaleLowerCase();
}

/** Alphabetical order by normalized name; ties broken by display form. */
export function sortNames(names) {
  return [...names].sort((a, b) => {
    const na = normalizeName(a);
    const nb = normalizeName(b);
    if (na !== nb) return na < nb ? -1 : 1;
    return a < b ? -1 : a > b ? 1 : 0;
  });
}

/**
 * Turns raw input values into a list of display names.
 * Empty inputs are dropped. `duplicates` lists each name (display form of its
 * first occurrence) that appears more than once under normalization.
 */
export function collectNames(rawValues) {
  const names = [];
  const firstSeen = new Map();
  const duplicates = [];
  for (const raw of rawValues) {
    const name = cleanName(raw);
    if (!name) continue;
    const key = normalizeName(name);
    if (firstSeen.has(key)) {
      const first = firstSeen.get(key);
      if (!duplicates.includes(first)) duplicates.push(first);
    } else {
      firstSeen.set(key, name);
    }
    names.push(name);
  }
  return { names, duplicates };
}

/** True when both lists contain the same names (ignoring order/case/blanks). */
export function sameNameSet(a, b) {
  const setA = new Set(a.map(normalizeName).filter(Boolean));
  const setB = new Set(b.map(normalizeName).filter(Boolean));
  if (setA.size !== setB.size) return false;
  for (const n of setA) if (!setB.has(n)) return false;
  return true;
}

// ---------------------------------------------------------------------------
// Randomness & generation (spec §5.1)

/** Uniform integer in [0, max) from crypto.getRandomValues (rejection sampled). */
export function secureRandomInt(max) {
  if (!globalThis.crypto?.getRandomValues) {
    throw new Error('crypto.getRandomValues is not available');
  }
  const limit = Math.floor(0x100000000 / max) * max;
  const buf = new Uint32Array(1);
  do {
    globalThis.crypto.getRandomValues(buf);
  } while (buf[0] >= limit);
  return buf[0] % max;
}

/** Fisher–Yates shuffle; returns a new array. */
export function shuffle(items, randomInt = secureRandomInt) {
  const out = [...items];
  for (let i = out.length - 1; i > 0; i--) {
    const j = randomInt(i + 1);
    [out[i], out[j]] = [out[j], out[i]];
  }
  return out;
}

/** Returns a copy of the map with keys in alphabetical order (spec §3.1). */
export function sortedMap(map) {
  const out = {};
  for (const key of sortNames(Object.keys(map))) out[key] = map[key];
  return out;
}

/**
 * Builds a random single-cycle pairing map from raw names.
 * Throws if there are fewer than 2 names or any duplicates.
 */
export function generatePairings(rawNames, randomInt = secureRandomInt) {
  const { names, duplicates } = collectNames(rawNames);
  if (duplicates.length) throw new Error(`Duplicate name(s): ${duplicates.join(', ')}.`);
  if (names.length < 2) throw new Error('Enter at least 2 names.');
  const order = shuffle(names, randomInt);
  const map = {};
  order.forEach((name, i) => {
    map[name] = order[(i + 1) % order.length];
  });
  return sortedMap(map);
}

// ---------------------------------------------------------------------------
// Validation (spec §3.1)

/** Returns null if `map` is a valid pairing map, otherwise a reason string. */
export function validatePairingMap(map) {
  if (map === null || typeof map !== 'object' || Array.isArray(map)) {
    return 'not an object';
  }
  const keys = Object.keys(map);
  if (keys.length < 2) return 'fewer than 2 names';

  const normalized = new Set();
  for (const key of keys) {
    if (!cleanName(key)) return 'empty name';
    const n = normalizeName(key);
    if (normalized.has(n)) return `duplicate name: ${key}`;
    normalized.add(n);
  }

  const values = new Set();
  for (const key of keys) {
    const value = map[key];
    if (typeof value !== 'string') return `non-string value for ${key}`;
    if (!Object.hasOwn(map, value)) return `${value} is not in the list`;
    if (value === key) return `${key} is paired with themselves`;
    if (values.has(value)) return `${value} appears more than once as a match`;
    values.add(value);
  }

  // Values are a permutation of keys; check it is one cycle, not several.
  let current = keys[0];
  for (let steps = 1; steps < keys.length; steps++) {
    current = map[current];
    if (current === keys[0]) return 'pairings form more than one cycle';
  }
  return null;
}

// ---------------------------------------------------------------------------
// Base64 / base64url (spec §3.2). Implemented here rather than with btoa/atob
// because those work on binary strings, not bytes.

const B64 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
const B64_LOOKUP = new Map([...B64].map((c, i) => [c, i]));

export function bytesToBase64(bytes) {
  let out = '';
  for (let i = 0; i < bytes.length; i += 3) {
    const b0 = bytes[i];
    const b1 = i + 1 < bytes.length ? bytes[i + 1] : 0;
    const b2 = i + 2 < bytes.length ? bytes[i + 2] : 0;
    const triple = (b0 << 16) | (b1 << 8) | b2;
    out += B64[(triple >> 18) & 63] + B64[(triple >> 12) & 63];
    out += i + 1 < bytes.length ? B64[(triple >> 6) & 63] : '=';
    out += i + 2 < bytes.length ? B64[triple & 63] : '=';
  }
  return out;
}

/** Accepts standard base64 or base64url, with or without padding. */
export function base64ToBytes(input) {
  let s = String(input).replace(/\s+/g, '').replace(/-/g, '+').replace(/_/g, '/');
  s = s.replace(/=+$/, '');
  if (s.length % 4 === 1 || /[^A-Za-z0-9+/]/.test(s)) {
    throw new InvalidCodeError('not valid base64');
  }
  const bytes = [];
  for (let i = 0; i < s.length; i += 4) {
    const chunk = s.slice(i, i + 4);
    let triple = 0;
    for (let j = 0; j < 4; j++) {
      triple = (triple << 6) | (j < chunk.length ? B64_LOOKUP.get(chunk[j]) : 0);
    }
    bytes.push((triple >> 16) & 255);
    if (chunk.length > 2) bytes.push((triple >> 8) & 255);
    if (chunk.length > 3) bytes.push(triple & 255);
  }
  return new Uint8Array(bytes);
}

export function toBase64Url(b64) {
  return b64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

// ---------------------------------------------------------------------------
// Scrambling (spec §3.2). Obfuscation only: it keeps a code pasted into a
// base64 decoder from showing the pairings, nothing more.

const SCRAMBLE_KEY = new TextEncoder().encode('obfuscate');

/** XORs each byte with a position-dependent key byte. Its own inverse. */
export function scramble(bytes) {
  return bytes.map((b, i) => b ^ SCRAMBLE_KEY[i % SCRAMBLE_KEY.length] ^ ((i * 31) & 255));
}

// ---------------------------------------------------------------------------
// Encode / decode

function encodePayload(payload) {
  const bytes = new TextEncoder().encode(JSON.stringify(payload));
  return bytesToBase64(scramble(bytes));
}

function decodePayload(code) {
  const bytes = scramble(base64ToBytes(code));
  let json;
  try {
    json = new TextDecoder('utf-8', { fatal: true }).decode(bytes);
  } catch {
    throw new InvalidCodeError('not valid UTF-8');
  }
  try {
    return JSON.parse(json);
  } catch {
    throw new InvalidCodeError('not valid JSON');
  }
}

/** Returns null if `payload` is a valid personal payload, otherwise a reason. */
export function validatePersonalPayload(payload) {
  if (!Array.isArray(payload) || payload.length !== 2) return 'not a [name, match] pair';
  const [name, match] = payload;
  if (typeof name !== 'string' || typeof match !== 'string') return 'non-string name';
  if (!cleanName(name) || !cleanName(match)) return 'empty name';
  if (normalizeName(name) === normalizeName(match)) return 'paired with themselves';
  return null;
}

/** Pairing map → standard base64 code (spec §3.1). */
export function encode(map) {
  return encodePayload(sortedMap(map));
}

/** One person's pairing → standard base64 personal code (spec §3.1a). */
export function encodePersonal(name, match) {
  return encodePayload([name, match]);
}

/**
 * Code (base64 or base64url) → `{ kind: 'list', map }` or
 * `{ kind: 'personal', name, match }`. Throws InvalidCodeError.
 */
export function decodeCode(code) {
  const payload = decodePayload(code);
  if (Array.isArray(payload)) {
    const problem = validatePersonalPayload(payload);
    if (problem) throw new InvalidCodeError(problem);
    return { kind: 'personal', name: payload[0], match: payload[1] };
  }
  const problem = validatePairingMap(payload);
  if (problem) throw new InvalidCodeError(problem);
  return { kind: 'list', map: payload };
}

/** Code → validated pairing map. Throws InvalidCodeError for anything else. */
export function decode(code) {
  const result = decodeCode(code);
  if (result.kind !== 'list') throw new InvalidCodeError('personal code, not a list');
  return result.map;
}

// ---------------------------------------------------------------------------
// Links (spec §4)

/**
 * Pulls the code out of user input: a full share link, a bare `#code=…`
 * fragment, or a raw code. Not parsed with URLSearchParams because that turns
 * `+` (valid in standard base64) into a space.
 */
export function extractCode(input) {
  const text = String(input).trim();
  const match = text.match(/[#?&]code=([^&#\s]*)/);
  if (match) {
    try {
      return decodeURIComponent(match[1]);
    } catch {
      return match[1];
    }
  }
  return text;
}

/** Code from a location hash such as `#code=abc`, or null if there is none. */
export function codeFromHash(hash) {
  const match = String(hash).match(/^#?(?:.*&)?code=([^&]*)/);
  if (!match) return null;
  try {
    return decodeURIComponent(match[1]);
  } catch {
    return match[1];
  }
}

/** `base` is the page URL without a fragment. */
export function buildShareUrl(base, code) {
  return `${String(base).split('#')[0]}#code=${toBase64Url(code)}`;
}

/** One `{ name, url }` per participant, alphabetically (spec §5.5). */
export function personalLinks(base, map) {
  return sortNames(Object.keys(map)).map((name) => ({
    name,
    url: buildShareUrl(base, encodePersonal(name, map[name])),
  }));
}

// ---------------------------------------------------------------------------
// Lookup (spec §5.3, §5.4)

/** The display name `name` is paired with, or null if `name` isn't listed. */
export function lookup(map, name) {
  const target = normalizeName(name);
  if (!target) return null;
  for (const key of Object.keys(map)) {
    if (normalizeName(key) === target) return map[key];
  }
  return null;
}

/** All names in a map, alphabetically (used when loading a list). */
export function namesFromMap(map) {
  return sortNames(Object.keys(map));
}
