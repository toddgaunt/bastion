import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  InvalidCodeError,
  base64ToBytes,
  buildShareUrl,
  bytesToBase64,
  decodeCode,
  encodePersonal,
  personalLinks,
  scramble,
  validatePersonalPayload,
  cleanName,
  codeFromHash,
  collectNames,
  decode,
  encode,
  extractCode,
  generatePairings,
  lookup,
  namesFromMap,
  normalizeName,
  sameNameSet,
  secureRandomInt,
  shuffle,
  toBase64Url,
  validatePairingMap,
} from '../src/core.js';

// Fast non-crypto RNG for bulk generation tests.
const rand = (max) => Math.floor(Math.random() * max);

// Encodes any JSON value the way the app does, to exercise validation paths.
const encodeJson = (value) => bytesToBase64(scramble(new TextEncoder().encode(JSON.stringify(value))));

// --- names ---------------------------------------------------------------

test('cleanName trims and collapses whitespace', () => {
  assert.equal(cleanName('  Mary   Ann \t'), 'Mary Ann');
});

test('normalizeName ignores case and spacing', () => {
  assert.equal(normalizeName(' ALICE '), normalizeName('alice'));
});

test('normalizeName applies NFC so composed and decomposed forms match', () => {
  assert.equal(normalizeName('José'), normalizeName('José'));
});

test('collectNames drops blank inputs', () => {
  assert.deepEqual(collectNames(['Bob', '', '   ', ' Alice ']).names, ['Bob', 'Alice']);
});

test('collectNames reports normalized duplicates once', () => {
  const { duplicates } = collectNames(['Bob', 'bob', ' BOB', 'Alice']);
  assert.deepEqual(duplicates, ['Bob']);
});

test('sameNameSet ignores order, case, whitespace, and blanks', () => {
  assert(sameNameSet(['Bob', 'Alice', ''], [' alice', 'BOB']));
});

test('sameNameSet detects added, removed, and renamed names', () => {
  assert(!sameNameSet(['Bob', 'Alice'], ['Bob', 'Alice', 'Mark']));
  assert(!sameNameSet(['Bob', 'Alice'], ['Bob']));
  assert(!sameNameSet(['Bob', 'Alice'], ['Bob', 'Alicia']));
});

// --- generation ----------------------------------------------------------

test('shuffle keeps every element', () => {
  const input = ['a', 'b', 'c', 'd', 'e'];
  assert.deepEqual([...shuffle(input, rand)].sort(), input);
});

test('generatePairings always produces a valid single cycle', () => {
  for (let n = 2; n <= 12; n++) {
    const names = Array.from({ length: n }, (_, i) => `Person ${i}`);
    for (let trial = 0; trial < 50; trial++) {
      const map = generatePairings(names, rand);
      assert.equal(validatePairingMap(map), null, `n=${n}`);
      assert.equal(Object.keys(map).length, n);
    }
  }
});

test('generatePairings with 2 names pairs them with each other', () => {
  assert.deepEqual(generatePairings(['Bob', 'Alice'], rand), { Alice: 'Bob', Bob: 'Alice' });
});

test('generatePairings stores cleaned display names', () => {
  const map = generatePairings(['  Mary   Ann ', 'Bob'], rand);
  assert.deepEqual(Object.keys(map), ['Bob', 'Mary Ann']);
});

test('generatePairings emits keys alphabetically', () => {
  const map = generatePairings(['mark', 'Alice', 'bob', 'Zoë'], rand);
  assert.deepEqual(Object.keys(map), ['Alice', 'bob', 'mark', 'Zoë']);
});

test('generatePairings rejects fewer than 2 names', () => {
  assert.throws(() => generatePairings(['Bob', ''], rand));
});

test('generatePairings rejects duplicates', () => {
  assert.throws(() => generatePairings(['Bob', 'bob', 'Alice'], rand));
});

test('generatePairings produces every possible cycle', () => {
  // 4 names have (4-1)! = 6 distinct cycles.
  const seen = new Set();
  for (let i = 0; i < 2000; i++) seen.add(JSON.stringify(generatePairings(['a', 'b', 'c', 'd'], rand)));
  assert.equal(seen.size, 6);
});

// --- validation ----------------------------------------------------------

test('validatePairingMap accepts a valid cycle', () => {
  assert.equal(validatePairingMap({ alice: 'mark', mark: 'john', john: 'alice' }), null);
});

test('validatePairingMap rejects invalid maps', () => {
  const bad = [
    null,
    [],
    'alice',
    { alice: 'bob' },
    { alice: 'alice', bob: 'bob' },
    { a: 'b', b: 'a', c: 'd', d: 'c' },
    { a: 'b', b: 'zed' },
    { a: 'b', b: 'b', c: 'a' },
    { a: 'b', b: 'c', c: 'b' },
    { Bob: 'bob', bob: 'Bob' },
    { a: 'b', b: 1 },
    { '  ': 'b', b: '  ' },
  ];
  for (const map of bad) assert(validatePairingMap(map) !== null, `accepted ${JSON.stringify(map)}`);
});

// --- encoding ------------------------------------------------------------

test('base64 matches known vectors', () => {
  const enc = (s) => bytesToBase64(new TextEncoder().encode(s));
  assert.equal(enc(''), '');
  assert.equal(enc('f'), 'Zg==');
  assert.equal(enc('fo'), 'Zm8=');
  assert.equal(enc('foo'), 'Zm9v');
  assert.equal(enc('foobar'), 'Zm9vYmFy');
});

test('base64ToBytes round-trips arbitrary bytes', () => {
  const bytes = Uint8Array.from({ length: 256 }, (_, i) => i);
  assert.deepEqual([...base64ToBytes(bytesToBase64(bytes))], [...bytes]);
  assert.deepEqual([...base64ToBytes(toBase64Url(bytesToBase64(bytes)))], [...bytes]);
});

test('base64ToBytes rejects invalid input', () => {
  assert.throws(() => base64ToBytes('abc$'));
  assert.throws(() => base64ToBytes('abcde'));
});

test('encode produces the specified JSON format', () => {
  const code = encode({ bob: 'alice', alice: 'bob' });
  const json = new TextDecoder().decode(scramble(base64ToBytes(code)));
  assert.equal(json, '{"alice":"bob","bob":"alice"}');
});

test('encode/decode round-trips non-ASCII, quotes, and emoji', () => {
  const map = generatePairings(['José', 'Zoë', '李', 'O"Neil', 'Ana 🎈', 'back\\slash'], rand);
  assert.deepEqual(decode(encode(map)), map);
});

test('decode accepts base64url without padding', () => {
  const map = generatePairings(['José', 'Zoë', '李', 'Ann'], rand);
  assert.deepEqual(decode(toBase64Url(encode(map))), map);
});

test('decode throws InvalidCodeError for garbage', () => {
  const cases = ['', 'not base64!', bytesToBase64(Uint8Array.from([0xff, 0xfe])), encodeJson([1, 2]), encodeJson({ a: 'a' })];
  for (const code of cases) {
    let error;
    try {
      decode(code);
    } catch (err) {
      error = err;
    }
    assert(error instanceof InvalidCodeError, `did not reject ${JSON.stringify(code)}`);
  }
});

test('scramble is its own inverse', () => {
  const bytes = Uint8Array.from({ length: 300 }, (_, i) => (i * 7) & 255);
  assert.deepEqual([...scramble(scramble(bytes))], [...bytes]);
});

test('a base64-decoded code does not show any names', () => {
  const map = generatePairings(['Alice', 'Bob', 'Mark', 'José'], rand);
  for (const code of [encode(map), encodePersonal('Alice', 'Bob')]) {
    const raw = new TextDecoder().decode(base64ToBytes(code));
    for (const name of ['Alice', 'Bob', 'Mark', 'José']) assert.ok(!raw.includes(name), `${name} visible`);
    assert.ok(!raw.includes('{') && !raw.includes('['), 'JSON structure visible');
  }
});

// --- personal codes ------------------------------------------------------

test('personal codes round-trip', () => {
  assert.deepEqual(decodeCode(encodePersonal('José', '李')), { kind: 'personal', name: 'José', match: '李' });
  assert.deepEqual(decodeCode(toBase64Url(encodePersonal('Bob', 'Alice'))), { kind: 'personal', name: 'Bob', match: 'Alice' });
});

test('decodeCode tells list and personal codes apart', () => {
  const map = { bob: 'alice', alice: 'bob' };
  assert.deepEqual(decodeCode(encode(map)), { kind: 'list', map: { alice: 'bob', bob: 'alice' } });
  assert.equal(decodeCode(encodePersonal('bob', 'alice')).kind, 'personal');
});

test('decode rejects personal codes', () => {
  assert.throws(() => decode(encodePersonal('bob', 'alice')), InvalidCodeError);
});

test('validatePersonalPayload rejects invalid payloads', () => {
  const bad = [[], ['bob'], ['bob', 'alice', 'mark'], ['bob', 2], ['', 'alice'], ['Bob', ' bob '], { 0: 'a', 1: 'b' }];
  for (const payload of bad) assert.notEqual(validatePersonalPayload(payload), null, JSON.stringify(payload));
  for (const payload of bad.filter(Array.isArray)) assert.throws(() => decodeCode(encodeJson(payload)), InvalidCodeError);
});

test('personalLinks gives each person a link to only their match', () => {
  const map = generatePairings(['Mark', 'alice', 'Bob', 'José'], rand);
  const links = personalLinks('https://example.com/app/', map);
  assert.deepEqual(links.map((l) => l.name), ['alice', 'Bob', 'José', 'Mark']);
  for (const { name, url } of links) {
    assert.ok(url.startsWith('https://example.com/app/#code='));
    assert.deepEqual(decodeCode(extractCode(url)), { kind: 'personal', name, match: map[name] });
  }
});

// --- links ---------------------------------------------------------------

test('extractCode handles full links, fragments, and raw codes', () => {
  assert.equal(extractCode('https://example.com/app/#code=abc-_'), 'abc-_');
  assert.equal(extractCode('  #code=xyz  '), 'xyz');
  assert.equal(extractCode('a+b/c=='), 'a+b/c==');
  assert.equal(extractCode('https://example.com/?code=a+b'), 'a+b');
});

test('codeFromHash reads only the code parameter', () => {
  assert.equal(codeFromHash('#code=abc'), 'abc');
  assert.equal(codeFromHash('#x=1&code=abc'), 'abc');
  assert.equal(codeFromHash('#reveal'), null);
  assert.equal(codeFromHash(''), null);
});

test('buildShareUrl uses base64url in the fragment', () => {
  const url = buildShareUrl('https://example.com/app/#old', 'a+b/c==');
  assert.equal(url, 'https://example.com/app/#code=a-b_c');
});

test('a share link decodes back to the same map', () => {
  const map = generatePairings(['José', 'Zoë', '李', 'Ann', 'Bob'], rand);
  const url = buildShareUrl('https://example.com/', encode(map));
  assert.deepEqual(decode(extractCode(url)), map);
});

// --- lookup --------------------------------------------------------------

test('lookup returns only who the person is paired with', () => {
  const map = { bob: 'alice', alice: 'bob' };
  assert.equal(lookup(map, 'bob'), 'alice');
});

test('lookup is case and whitespace insensitive and returns display form', () => {
  const map = { 'Mary Ann': 'Bob', Bob: 'Mary Ann' };
  assert.equal(lookup(map, '  mary   ANN '), 'Bob');
  assert.equal(lookup(map, 'BOB'), 'Mary Ann');
});

test('lookup returns null for unknown or empty names', () => {
  const map = { bob: 'alice', alice: 'bob' };
  assert.equal(lookup(map, 'mark'), null);
  assert.equal(lookup(map, '   '), null);
});

test('namesFromMap returns every name alphabetically', () => {
  assert.deepEqual(namesFromMap({ mark: 'Alice', Alice: 'bob', bob: 'mark' }), ['Alice', 'bob', 'mark']);
});

// --- randomness ----------------------------------------------------------

test('secureRandomInt stays in range and covers every value', () => {
  const seen = new Set();
  for (let i = 0; i < 1000; i++) {
    const n = secureRandomInt(7);
    assert.ok(Number.isInteger(n) && n >= 0 && n < 7);
    seen.add(n);
  }
  assert.equal(seen.size, 7);
});

test('generatePairings uses crypto randomness by default', () => {
  assert.equal(validatePairingMap(generatePairings(['a', 'b', 'c', 'd', 'e'])), null);
});
