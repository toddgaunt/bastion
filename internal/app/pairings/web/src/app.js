// UI wiring for the Pairings app. All pairing logic lives in core.js.
import {
  buildShareUrl,
  codeFromHash,
  collectNames,
  decodeCode,
  encode,
  extractCode,
  generatePairings,
  lookup,
  namesFromMap,
  normalizeName,
  personalLinks,
  sameNameSet,
} from './core.js';

const INITIAL_ROWS = 3;
const GROUP_SHARE_TEXT = "Open this link and enter your name to see who you're paired with.";
const PERSONAL_SHARE_TEXT = "Open this link to see who you're paired with.";
const canShare = typeof navigator.share === 'function';
const $ = (id) => document.getElementById(id);

const els = {
  generateView: $('generate-view'),
  revealView: $('reveal-view'),
  live: $('live'),
  // generate
  loadPanel: $('load-panel'),
  loadForm: $('load-form'),
  loadInput: $('load-input'),
  loadError: $('load-error'),
  namesForm: $('names-form'),
  nameList: $('name-list'),
  addName: $('add-name'),
  perPerson: $('per-person'),
  helpToggle: $('per-person-help-toggle'),
  help: $('per-person-help'),
  namesError: $('names-error'),
  output: $('output'),
  staleWarning: $('stale-warning'),
  participantCount: $('participant-count'),
  groupOutput: $('group-output'),
  personalOutput: $('personal-output'),
  personalLinks: $('personal-links'),
  codeLabel: $('code-label'),
  organizerHint: $('organizer-hint'),
  codeOutput: $('code-output'),
  linkOutput: $('link-output'),
  copyCode: $('copy-code'),
  copyLink: $('copy-link'),
  shareLink: $('share-link'),
  // reveal
  revealInvalid: $('reveal-invalid'),
  personalReveal: $('personal-reveal'),
  personalFor: $('personal-for'),
  personalRevealBtn: $('personal-reveal-btn'),
  revealForm: $('reveal-form'),
  manualCode: $('manual-code'),
  revealCode: $('reveal-code'),
  revealName: $('reveal-name'),
  revealError: $('reveal-error'),
  revealResult: $('reveal-result'),
  resultName: $('result-name'),
  hideResult: $('hide-result'),
};

/** Names used for the code currently in the output panel. */
let generatedNames = null;
/** Pairing map currently in the output panel. */
let generatedMap = null;
/** Decoded code from the URL fragment, for the reveal view (see decodeCode). */
let linked = null;

function announce(message) {
  // Clear first so repeating the same message is still announced.
  els.live.textContent = '';
  setTimeout(() => {
    els.live.textContent = message;
  }, 50);
}

function showError(el, message) {
  el.textContent = message;
  el.hidden = !message;
}

// ---------------------------------------------------------------------------
// Name rows

function nameInputs() {
  return [...els.nameList.querySelectorAll('input')];
}

function nameValues() {
  return nameInputs().map((input) => input.value);
}

function createRow(value = '') {
  const li = document.createElement('li');

  const label = document.createElement('label');
  label.className = 'visually-hidden';

  const input = document.createElement('input');
  input.type = 'text';
  input.autocomplete = 'off';
  input.value = value;
  input.setAttribute('aria-describedby', 'names-error');

  const remove = document.createElement('button');
  remove.type = 'button';
  remove.className = 'secondary remove-name';
  remove.innerHTML = '<span aria-hidden="true">×</span>';

  li.append(label, input, remove);
  return li;
}

/** Keeps ids, labels, and remove-button state in sync with row order. */
function renumberRows() {
  const rows = [...els.nameList.children];
  rows.forEach((li, i) => {
    const n = i + 1;
    const [label, input, remove] = li.children;
    input.id = `name-${n}`;
    label.htmlFor = input.id;
    label.textContent = `Name ${n}`;
    input.placeholder = `Name ${n}`;
    remove.setAttribute('aria-label', `Remove name ${n}`);
    remove.disabled = rows.length === 1;
  });
}

function addRow(value = '', { focus = false } = {}) {
  const li = createRow(value);
  els.nameList.append(li);
  renumberRows();
  if (focus) li.querySelector('input').focus();
  return li;
}

function setRows(names) {
  els.nameList.replaceChildren();
  const count = Math.max(names.length, 1);
  for (let i = 0; i < count; i++) addRow(names[i] ?? '');
}

function removeRow(li) {
  if (els.nameList.children.length <= 1) return;
  const next = li.nextElementSibling ?? li.previousElementSibling;
  li.remove();
  renumberRows();
  next?.querySelector('input').focus();
  onNamesChanged();
}

function clearNameErrors() {
  showError(els.namesError, '');
  for (const input of nameInputs()) input.removeAttribute('aria-invalid');
}

function onNamesChanged() {
  clearNameErrors();
  updateStaleWarning();
}

els.addName.addEventListener('click', () => {
  addRow('', { focus: true });
});

els.nameList.addEventListener('click', (event) => {
  const button = event.target.closest('.remove-name');
  if (button) removeRow(button.closest('li'));
});

els.nameList.addEventListener('input', onNamesChanged);

els.nameList.addEventListener('keydown', (event) => {
  if (event.key !== 'Enter' || event.target.tagName !== 'INPUT') return;
  event.preventDefault();
  const li = event.target.closest('li');
  const next = li.nextElementSibling;
  if (next) next.querySelector('input').focus();
  else addRow('', { focus: true });
});

// ---------------------------------------------------------------------------
// Link per person option

els.helpToggle.addEventListener('click', () => {
  const open = els.help.hidden;
  els.help.hidden = !open;
  els.helpToggle.setAttribute('aria-expanded', String(open));
});

els.perPerson.addEventListener('change', () => {
  // Same pairings, other kind of output (spec §5.2).
  if (generatedMap) renderOutput();
});

// ---------------------------------------------------------------------------
// Generate

function updateStaleWarning() {
  if (els.output.hidden || !generatedNames) return;
  const wasHidden = els.staleWarning.hidden;
  els.staleWarning.hidden = sameNameSet(nameValues(), generatedNames);
  if (wasHidden && !els.staleWarning.hidden) {
    announce('The names have changed. This code no longer matches the list.');
  }
}

function createPersonalRow({ name, url }, index) {
  const li = document.createElement('li');

  const label = document.createElement('label');
  label.className = 'person-name';
  label.textContent = name;

  const row = document.createElement('div');
  row.className = 'row';

  const input = document.createElement('input');
  input.type = 'text';
  input.readOnly = true;
  input.spellcheck = false;
  input.value = url;
  input.id = `personal-link-${index + 1}`;
  label.htmlFor = input.id;

  const copy = document.createElement('button');
  copy.type = 'button';
  copy.className = 'secondary';
  copy.textContent = 'Copy';
  copy.setAttribute('aria-label', `Copy link for ${name}`);
  copy.addEventListener('click', () => copyFrom(input, copy));

  row.append(input, copy);
  if (canShare) {
    const share = document.createElement('button');
    share.type = 'button';
    share.className = 'secondary';
    share.textContent = 'Share…';
    share.setAttribute('aria-label', `Share link for ${name}`);
    share.addEventListener('click', () => shareUrl(url, PERSONAL_SHARE_TEXT, input, copy));
    row.append(share);
  }

  li.append(label, row);
  return li;
}

function renderOutput() {
  const base = location.href.split('#')[0];
  const code = encode(generatedMap);
  const perPerson = els.perPerson.checked;

  els.codeOutput.value = code;
  els.participantCount.textContent = `${Object.keys(generatedMap).length} people in this cycle.`;
  els.groupOutput.hidden = perPerson;
  els.personalOutput.hidden = !perPerson;
  els.codeLabel.textContent = perPerson ? 'Organizer code' : 'Code';
  els.organizerHint.hidden = !perPerson;

  if (perPerson) {
    els.linkOutput.value = '';
    els.personalLinks.replaceChildren(...personalLinks(base, generatedMap).map(createPersonalRow));
  } else {
    els.linkOutput.value = buildShareUrl(base, code);
    els.personalLinks.replaceChildren();
  }
}

els.namesForm.addEventListener('submit', (event) => {
  event.preventDefault();
  clearNameErrors();

  const { names, duplicates } = collectNames(nameValues());
  if (duplicates.length) {
    const dupKeys = new Set(duplicates.map(normalizeName));
    for (const input of nameInputs()) {
      if (dupKeys.has(normalizeName(input.value))) input.setAttribute('aria-invalid', 'true');
    }
    showError(els.namesError, `Duplicate name(s): ${duplicates.join(', ')}.`);
    nameInputs().find((input) => input.hasAttribute('aria-invalid'))?.focus();
    return;
  }
  if (names.length < 2) {
    showError(els.namesError, 'Enter at least 2 names.');
    nameInputs().find((input) => !input.value.trim())?.focus();
    return;
  }

  generatedMap = generatePairings(names);
  generatedNames = names;
  renderOutput();
  els.staleWarning.hidden = true;
  els.output.hidden = false;

  els.output.scrollIntoView({ behavior: 'smooth', block: 'start' });
  announce(`Pairings generated for ${names.length} people.`);
});

// ---------------------------------------------------------------------------
// Copy & share

const copyTimers = new WeakMap();

function flashButton(button, text, ms = 2000) {
  if (!button.dataset.label) button.dataset.label = button.textContent;
  clearTimeout(copyTimers.get(button));
  button.textContent = text;
  copyTimers.set(button, setTimeout(() => {
    button.textContent = button.dataset.label;
  }, ms));
}

async function copyFrom(field, button) {
  try {
    if (!navigator.clipboard?.writeText) throw new Error('Clipboard API unavailable');
    await navigator.clipboard.writeText(field.value);
    flashButton(button, 'Copied!');
    announce('Copied to clipboard.');
  } catch {
    field.focus();
    field.select();
    const keys = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘C' : 'Ctrl+C';
    flashButton(button, `Press ${keys} to copy`, 4000);
    announce(`Text selected. Press ${keys} to copy.`);
  }
}

/** Opens the native share sheet; falls back to copying if sharing fails. */
async function shareUrl(url, text, field, copyButton) {
  try {
    await navigator.share({ title: 'Pairings', text, url });
  } catch (err) {
    if (err?.name !== 'AbortError') copyFrom(field, copyButton);
  }
}

els.copyCode.addEventListener('click', () => copyFrom(els.codeOutput, els.copyCode));
els.copyLink.addEventListener('click', () => copyFrom(els.linkOutput, els.copyLink));

if (canShare) {
  els.shareLink.hidden = false;
  els.shareLink.addEventListener('click', () => {
    shareUrl(els.linkOutput.value, GROUP_SHARE_TEXT, els.linkOutput, els.copyLink);
  });
}

// ---------------------------------------------------------------------------
// Load existing list

els.loadForm.addEventListener('submit', (event) => {
  event.preventDefault();
  showError(els.loadError, '');
  els.loadInput.removeAttribute('aria-invalid');

  let decoded;
  try {
    decoded = decodeCode(extractCode(els.loadInput.value));
  } catch {
    els.loadInput.setAttribute('aria-invalid', 'true');
    showError(els.loadError, "This code isn't a valid pairing list.");
    return;
  }
  if (decoded.kind === 'personal') {
    els.loadInput.setAttribute('aria-invalid', 'true');
    showError(els.loadError, "This is one person's link. To load the list, use the organizer code.");
    return;
  }

  const hasNames = nameValues().some((value) => value.trim());
  if (hasNames && !confirm('Replace the current names with the names from this code?')) return;

  const names = namesFromMap(decoded.map);
  setRows(names);
  onNamesChanged();
  els.loadInput.value = '';
  els.loadPanel.open = false;
  nameInputs()[0]?.focus();
  announce(`Loaded ${names.length} names.`);
});

// ---------------------------------------------------------------------------
// Reveal

function resetReveal() {
  els.revealName.value = '';
  els.revealCode.value = '';
  els.resultName.textContent = '';
  els.revealResult.hidden = true;
  els.personalRevealBtn.hidden = false;
  showError(els.revealError, '');
}

function showResult(match) {
  els.resultName.textContent = match;
  els.revealResult.hidden = false;
  announce(`You're paired with ${match}.`);
}

function showReveal(code) {
  resetReveal();
  els.generateView.hidden = true;
  els.revealView.hidden = false;
  els.revealInvalid.hidden = true;
  els.personalReveal.hidden = true;
  els.revealForm.hidden = false;

  if (code === null) {
    // Arrived via "I have a code": ask for the code too.
    linked = null;
    els.manualCode.hidden = false;
    els.revealCode.focus();
    return;
  }

  els.manualCode.hidden = true;
  try {
    linked = decodeCode(code);
  } catch {
    linked = null;
    els.revealInvalid.hidden = false;
    els.revealForm.hidden = true;
    return;
  }

  if (linked.kind === 'personal') {
    els.personalFor.textContent = linked.name;
    els.personalReveal.hidden = false;
    els.revealForm.hidden = true;
    els.personalRevealBtn.focus();
  } else {
    els.revealName.focus();
  }
}

els.personalRevealBtn.addEventListener('click', () => {
  els.personalRevealBtn.hidden = true;
  showResult(linked.match);
  els.hideResult.focus();
});

els.revealForm.addEventListener('submit', (event) => {
  event.preventDefault();
  showError(els.revealError, '');
  els.revealResult.hidden = true;

  let decoded = linked;
  if (!decoded) {
    try {
      decoded = decodeCode(extractCode(els.revealCode.value));
    } catch {
      showError(els.revealError, "This code isn't a valid pairing list.");
      els.revealCode.focus();
      return;
    }
  }

  if (!els.revealName.value.trim()) {
    showError(els.revealError, 'Enter your name.');
    els.revealName.focus();
    return;
  }

  const name = els.revealName.value;
  const match = decoded.kind === 'personal'
    ? (normalizeName(name) === normalizeName(decoded.name) ? decoded.match : null)
    : lookup(decoded.map, name);
  if (match === null) {
    showError(els.revealError, "That name isn't in this list. Check the spelling with the organizer.");
    els.revealName.focus();
    return;
  }

  showResult(match);
});

els.hideResult.addEventListener('click', () => {
  els.resultName.textContent = '';
  els.revealResult.hidden = true;
  showError(els.revealError, '');
  if (linked?.kind === 'personal') {
    els.personalRevealBtn.hidden = false;
    els.personalRevealBtn.focus();
  } else {
    els.revealName.value = '';
    els.revealName.focus();
  }
});

// ---------------------------------------------------------------------------
// Routing (spec §4)

function showGenerate() {
  linked = null;
  resetReveal();
  els.revealView.hidden = true;
  els.generateView.hidden = false;
}

function route() {
  const code = codeFromHash(location.hash);
  if (code !== null) showReveal(code);
  else if (location.hash === '#reveal') showReveal(null);
  else showGenerate();
}

window.addEventListener('hashchange', route);

setRows(Array(INITIAL_ROWS).fill(''));
route();
