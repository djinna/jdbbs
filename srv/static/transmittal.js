// Manuscript Transmittal SPA
const $ = (s, p) => (p || document).querySelector(s);
const $$ = (s, p) => [...(p || document).querySelectorAll(s)];
const h = (tag, attrs, ...kids) => {
  const el = document.createElement(tag);
  if (attrs) Object.entries(attrs).forEach(([k, v]) => {
    if (k.startsWith('on')) el.addEventListener(k.slice(2).toLowerCase(), v);
    else if (k === 'className') el.className = v;
    else if (k === 'htmlFor') el.htmlFor = v;
    else if (k === 'checked' || k === 'selected' || k === 'disabled') { if (v) el[k] = true; }
    else if (v !== undefined && v !== null && v !== false) el.setAttribute(k, v);
  });
  kids.flat(Infinity).forEach(c => {
    if (c == null || c === false) return;
    el.appendChild(c instanceof Node ? c : document.createTextNode(String(c)));
  });
  return el;
};

const api = async (url, opts = {}) => {
  const r = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  });
  const data = await r.json();
  if (!r.ok && data.error === 'unauthorized') throw new Error('unauthorized');
  if (!r.ok) throw new Error(data.error || r.statusText);
  return data;
};

let contactEmail = 'j@djinna.com';
fetch('/api/public/config').then(r => r.ok ? r.json() : null).then(c => { if (c && c.contact_email) contactEmail = c.contact_email; }).catch(() => {});

// ─── Theme (canonical JdbbTheme — see theme.js / theme.css)───
// One theme system only: state lives in localStorage `prodcal-theme-v1`,
// owned by JdbbTheme (theme.js migrates the legacy font keys).
function _ensureThemeBar() {
  var bar = document.getElementById('theme-bar');
  if (!bar) return;
  if (!window.JdbbTheme) return;
  if (bar.dataset.mounted === 'true') return;
  JdbbTheme.mount(bar);
  bar.dataset.mounted = 'true';
}
function _applyTheme() {
  if (window.JdbbTheme) JdbbTheme.apply(document.getElementById('theme-bar'));
}
function themeBtn() { return h('div', { id: 'theme-bar' }); }
function getTheme() { return window.JdbbTheme && JdbbTheme.isDark() ? 'dark' : 'light'; }

// ─── State ───
let state = {
  view: 'loading', // loading, auth, form
  // 0.17 (C): the transmittal is section 1 of the factory page. factory.js
  // calls JdbbTransmittal.mount() and we render into #fx-transmittal without
  // our own page header, step strip or sign-in gate (the factory owns those).
  embedded: false,
  mount: null,
  projectId: null,
  project: null,
  transmittal: null, // {id, project_id, status, data}
  pathClient: null,
  pathProject: null,
  saveStatus: '', // '', 'saving', 'saved', 'error'
  allProjects: [],  // for project switcher
  versions: null,   // null=not loaded, []=loaded
  showVersions: false,
  showDuplicate: false,
  showEmail: false,
  emailSending: false,
  emailResult: null, // {ok, error}
  emailConfigured: null, // null=unknown, true, false
  manuscript: null, // C7: {hasFile, filename, inspected, chapters, words, images} from the newest upload
};

// ─── Auto-save with debounce ───
let saveTimer = null;
function scheduleSave() {
  state.saveStatus = 'saving';
  updateSaveIndicator();
  clearTimeout(saveTimer);
  saveTimer = setTimeout(doSave, 600);
}

async function doSave() {
  if (!state.transmittal || !state.projectId) return;
  try {
    await api('/api/projects/' + state.projectId + '/transmittal', {
      method: 'PUT',
      body: JSON.stringify({ status: state.transmittal.status, data: state.transmittal.data }),
    });
    state.saveStatus = 'saved';
    updateSaveIndicator();
    setTimeout(() => { if (state.saveStatus === 'saved') { state.saveStatus = ''; updateSaveIndicator(); } }, 2000);
  } catch (e) {
    state.saveStatus = 'error';
    updateSaveIndicator();
    console.error('Save failed:', e);
  }
}

function updateSaveIndicator() {
  const el = $('#tx-save-status');
  if (!el) return;
  el.className = 'tx-save-status ' + state.saveStatus;
  el.textContent = state.saveStatus === 'saving' ? 'Saving...' : state.saveStatus === 'saved' ? 'Saved ✓' : state.saveStatus === 'error' ? 'Error!' : '';
}

// ─── Project switcher ───
async function loadAllProjects() {
  try {
    // Scope to current client so dropdown only shows this client's projects
    if (state.pathClient) {
      const raw = await api('/api/clients/' + state.pathClient + '/projects');
      // Normalize snake_case keys to PascalCase used by project switcher
      state.allProjects = raw.map(p => ({
        ID: p.id, Name: p.name,
        ClientSlug: p.client_slug, ProjectSlug: p.project_slug,
        StartDate: p.start_date, UpdatedAt: p.updated_at,
      }));
    } else {
      state.allProjects = await api('/api/projects');
    }
    // Re-render to show project switcher once loaded
    if (state.view === 'form') render();
  } catch (e) {
    console.error('Failed to load projects:', e);
    state.allProjects = [];
  }
}

function absoluteURL(path) {
  return new URL(path, window.location.origin + '/').toString();
}

function switchProject(proj) {
  const url = absoluteURL('/' + proj.ClientSlug + '/' + proj.ProjectSlug + '/transmittal/');
  window.location.href = url;
}

// ─── Version history ───
async function loadVersions() {
  try {
    state.versions = await api('/api/transmittals/' + state.projectId + '/versions');
  } catch (e) {
    console.error('Failed to load versions:', e);
    state.versions = [];
  }
  render();
}

async function restoreVersion(vid) {
  if (!confirm('Restore this version? Current state will be saved as a version first.')) return;
  try {
    await api('/api/transmittals/' + state.projectId + '/versions/' + vid + '/restore', { method: 'POST' });
    await loadTransmittal();
    await loadVersions();
  } catch (e) {
    alert('Restore failed: ' + e.message);
  }
}

async function previewVersion(vid) {
  try {
    const v = await api('/api/transmittals/' + state.projectId + '/versions/' + vid);
    // Temporarily show the version data
    state.transmittal = { ...state.transmittal, data: v.data, status: v.status, _preview: vid, _previewDate: v.saved_at };
    render();
  } catch (e) {
    alert('Failed to load version: ' + e.message);
  }
}

function exitPreview() {
  // Reload current live data
  loadTransmittal().then(() => loadVersions());
}

// ─── Duplicate transmittal ───
async function duplicateToProject(targetId) {
  try {
    await api('/api/transmittals/' + state.projectId + '/duplicate', {
      method: 'POST',
      body: JSON.stringify({ target_project_id: targetId }),
    });
    const target = state.allProjects.find(p => p.ID === targetId);
    if (target) {
      if (confirm('Transmittal duplicated! Go to ' + target.Name + ' transmittal?')) {
        window.location.href = absoluteURL('/' + target.ClientSlug + '/' + target.ProjectSlug + '/transmittal/');
      }
    } else {
      alert('Duplicated successfully!');
    }
  } catch (e) {
    alert('Duplicate failed: ' + e.message);
  }
}

// Helper: update a nested field in transmittal data
function setField(path, value) {
  const parts = path.split('.');
  let obj = state.transmittal.data;
  for (let i = 0; i < parts.length - 1; i++) {
    if (obj[parts[i]] === undefined) obj[parts[i]] = {};
    obj = obj[parts[i]];
  }
  obj[parts[parts.length - 1]] = value;
  scheduleSave();
  if (/^(page_iv|book|cover)\./.test(path)) refreshCopyrightPreview();
  // Live-update header title when book title changes
  if (path === 'book.title') {
    const el = document.querySelector('.page-header-title');
    if (el) el.textContent = value || state.project?.Name || 'Transmittal';
  }
}

function getField(path) {
  const parts = path.split('.');
  let obj = state.transmittal.data;
  for (const p of parts) {
    if (obj == null) return '';
    obj = obj[p];
  }
  return obj ?? '';
}

// ─── Render ───
function render() {
  const app = state.mount || $('#app');
  if (!app) return;
  app.innerHTML = '';
  if (state.view === 'auth' && state.embedded) {
    // The factory page has the sign-in gate; tell it and stay empty.
    document.dispatchEvent(new CustomEvent('tx:unauthorized'));
    return;
  }
  if (state.view === 'loading') app.appendChild(h('div', { className: 'tx-container' }, h('p', null, 'Loading...')));
  else if (state.view === 'auth') app.appendChild(renderAuth());
  else if (state.view === 'form') app.appendChild(renderForm());
  if (state.embedded) return;
  _ensureThemeBar();
  _applyTheme();
}

// Let the host page (factory.js) mirror the status in its step strip.
function emitStatus() {
  if (!state.embedded || !state.transmittal) return;
  document.dispatchEvent(new CustomEvent('tx:status', { detail: { status: state.transmittal.status } }));
}

// ─── Auth (reused pattern from calendar) ───
function renderAuth() {
  let input, errEl;
  const label = state.project ? state.project.Name : (state.pathClient || 'this project');
  const doLogin = async () => {
    try {
      await api('/api/projects/' + state.projectId + '/verify', {
        method: 'POST', body: JSON.stringify({ password: input.value }),
      });
      await loadTransmittal();
    } catch { errEl.textContent = 'Invalid password'; errEl.style.display = ''; }
  };
  return h('div', { className: 'auth-screen' },
    h('h2', null, 'Password Required'),
    h('p', { className: 'auth-sub' }, state.project ? state.project.Name + ' — Transmittal' : 'This project is protected'),
    input = h('input', { type: 'password', placeholder: 'Enter password', onKeydown: (e) => { if (e.key === 'Enter') doLogin(); } }),
    h('div', null, h('button', { className: 'btn btn-primary', onClick: doLogin }, 'Unlock')),
    errEl = h('div', { className: 'error', style: 'display:none' }),
    h('div', { className: 'forgot-link' },
      h('a', { href: '#', onClick: (e) => {
        e.preventDefault();
        const subject = encodeURIComponent('Portal access — ' + label);
        const body = encodeURIComponent(
          'Hi,\n\nI’d like a password reset (or password) for the JDBB client portal.\n\n' +
          'Project: ' + label + '\n' +
          'Portal URL: ' + window.location.href + '\n\n' +
          'Thanks.\n'
        );
        window.location.href = 'mailto:' + contactEmail + '?subject=' + subject + '&body=' + body;
      } }, 'Forgot or need a reset?')),
    h('div', { className: 'back-link' },
      h('a', { href: state.pathClient ? '/' + state.pathClient + '/' : '/' }, '← Back')),
  );
}

// ─── Boot ───
async function loadTransmittal() {
  const tx = await api('/api/projects/' + state.projectId + '/transmittal');
  state.transmittal = tx;
  state.view = 'form';
  render();
  emitStatus();
  loadManuscriptStats();
}

// C7: Chapters / Words / Images are no longer typed in — they are read off the
// newest uploaded manuscript (its Inspect book map). Fails quietly: the row
// then says "counted at upload".
async function loadManuscriptStats() {
  try {
    const rows = await api('/api/projects/' + state.projectId + '/books');
    const list = (Array.isArray(rows) ? rows : []).filter(b => b && b.ID);
    list.sort((a, b) => (new Date(b.CreatedAt || 0) - new Date(a.CreatedAt || 0)) || (b.ID - a.ID));
    const cur = list[0];
    if (!cur) { state.manuscript = { hasFile: false }; render(); return; }
    const ms = { hasFile: true, filename: cur.SourceFilename || '', bookId: cur.ID };
    try {
      const pf = await api('/api/projects/' + state.projectId + '/preflight?book_id=' + cur.ID);
      const bm = pf && pf.book_map;
      if (bm && Array.isArray(bm.sections)) {
        ms.inspected = true;
        ms.chapters = bm.sections.filter(s => s.kind === 'body').length;
        // Reports stored before C7 carry no counts (counted=false): say so
        // rather than showing 0.
        ms.words = bm.counted ? bm.words : null;
        ms.images = bm.counted ? bm.images : (Array.isArray(pf.images) ? pf.images.length : null);
        ms.inspectedAt = pf.updated_at || '';
      }
    } catch (e) { /* not inspected yet */ }
    state.manuscript = ms;
  } catch (e) {
    state.manuscript = { hasFile: false };
  }
  // Don't yank focus from someone already typing; the row fills on the next render.
  const a = document.activeElement;
  if (a && (a.tagName === 'INPUT' || a.tagName === 'TEXTAREA' || a.tagName === 'SELECT')) return;
  render();
}

// Embedded entry point (factory page). factory.js has already resolved the
// project and passed the sign-in gate; `info` is its /api/project-by-path
// response. Errors after this point surface as tx:unauthorized (handled by
// the factory's gate) or as a line inside the section.
window.JdbbTransmittal = {
  mount: async function (el, client, project, info) {
    state.embedded = true;
    state.mount = el;
    state.pathClient = client;
    state.pathProject = project;
    state.project = info.project;
    state.projectId = info.project.ID;
    state.isAdmin = !!info.is_admin;
    state.passEmail = info.pass_email || '';
    try {
      await loadTransmittal();
      loadAllProjects();
    } catch (e) {
      if (e.message === 'unauthorized') { state.view = 'auth'; render(); }
      else { el.textContent = 'Couldn\u2019t load the transmittal: ' + e.message; }
    }
  },
};

(async function boot() {
  if (document.getElementById('fx-transmittal')) return; // factory page drives via mount()
  const parts = window.location.pathname.replace(/\/+$/, '').split('/').filter(Boolean);
  // Expect: [client, project, 'transmittal']
  if (parts.length >= 3 && parts[2] === 'transmittal') {
    state.pathClient = parts[0];
    state.pathProject = parts[1];
    try {
      const info = await api('/api/project-by-path/' + parts[0] + '/' + parts[1]);
      state.project = info.project;
      state.projectId = info.project.ID;
      state.isAdmin = !!info.is_admin;
      state.passEmail = info.pass_email || '';
      if (info.has_auth && !info.authenticated) {
        state.view = 'auth';
        render();
        return;
      }
      await loadTransmittal();
      // Load project list in background for switcher/duplicate
      loadAllProjects();
    } catch (e) {
      if (e.message === 'unauthorized') { state.view = 'auth'; render(); }
      else { document.body.textContent = 'Error: ' + e.message; }
    }
  } else {
    document.body.textContent = 'Invalid URL. Expected /{client}/{project}/transmittal/';
  }
})();

// ─── Form field helpers ───
function textField(label, path, opts = {}) {
  const currentVal = getField(path);
  const val = opts.value !== undefined ? opts.value : (currentVal || '');
  const inp = h('input', {
    type: opts.type || 'text',
    value: val,
    placeholder: opts.placeholder || '',
    readOnly: opts.readOnly ? 'readonly' : undefined,
    maxLength: opts.maxLength || undefined,
    onInput: opts.readOnly ? undefined : (e) => setField(path, opts.type === 'number' ? (parseFloat(e.target.value) || 0) : e.target.value),
  });
  return h('div', { className: `tx-field ${opts.className || ''}`.trim() },
    label ? h('label', null, label) : null,
    inp,
    opts.helpText ? h('div', { className: 'tx-help' }, opts.helpText) : null,
  );
}

function textareaField(label, path, opts = {}) {
  const val = getField(path) || '';
  const ta = h('textarea', {
    rows: opts.rows || 3,
    placeholder: opts.placeholder || '',
    onInput: (e) => setField(path, e.target.value),
  }, val);
  return h('div', { className: `tx-field ${opts.className || ''}`.trim() },
    label ? h('label', null, label) : null,
    ta,
    opts.helpText ? h('div', { className: 'tx-help' }, opts.helpText) : null,
  );
}

function selectField(label, path, options, opts = {}) {
  const val = getField(path) || '';
  const sel = h('select', { onChange: (e) => setField(path, e.target.value) },
    ...options.map(([v, l]) => {
      const opt = h('option', { value: v }, l);
      if (val === v) opt.selected = true;
      return opt;
    })
  );
  return h('div', { className: `tx-field ${opts.className || ''}`.trim() },
    label ? h('label', null, label) : null,
    sel,
    opts.helpText ? h('div', { className: 'tx-help' }, opts.helpText) : null,
  );
}

function checkField(label, path) {
  const val = !!getField(path);
  return h('label', { className: 'tx-check' },
    h('input', { type: 'checkbox', checked: val ? 'checked' : undefined, onChange: (e) => setField(path, e.target.checked) }),
    label
  );
}

// ─── Completion calc ───
function getChecklistItemStatus(item) {
  if (!item) return '';
  if (item.status) return item.status;
  if (item.here_now) return 'included';
  if (item.to_come_when) return 'later';
  return '';
}

function calcCompletion() {
  const d = state.transmittal.data;
  let filled = 0, total = 0;
  // Book fields
  for (const k of ['author','title','publisher','editor','isbn_paper','isbn_epub','isbn_cloth']) {
    total++; if (d.book && d.book[k]) filled++;
  }
  // Schedule (Production section dropped in C6; the one field left is optional)
  total++; if (d.production && d.production.target_date) filled++;
  // Checklist — count items with explicit status, while preserving older saved data
  if (d.checklist) {
    // Generated pages and the hidden CIP row are complete as they stand.
    const auto = ['Half title pg', 'Title pg', 'Copyright pg', 'Contents', 'CIP'];
    const typed = d.checklist.filter(c => !auto.includes(c.component));
    total += typed.length;
    filled += typed.filter(c => !!getChecklistItemStatus(c)).length;
  }
  if (d.backmatter) {
    total += d.backmatter.length;
    filled += d.backmatter.filter(c => !!getChecklistItemStatus(c)).length;
  }
  // Format: the trim choice (complexity tier dropped with the Format rewrite)
  total++; if (d.design && d.design.trim) filled++;
  return total > 0 ? Math.round((filled / total) * 100) : 0;
}

// ─── Format date for display ───
function fmtDate(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d)) return iso;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    + ' ' + d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
}

function parseYMD(dateStr) {
  if (!dateStr) return null;
  const d = new Date(dateStr + 'T00:00:00');
  return isNaN(d) ? null : d;
}

function toYMD(dateObj) {
  if (!dateObj || isNaN(dateObj)) return '';
  const y = dateObj.getFullYear();
  const m = String(dateObj.getMonth() + 1).padStart(2, '0');
  const d = String(dateObj.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function calcWeeksBetween(startStr, endStr) {
  const start = parseYMD(startStr);
  const end = parseYMD(endStr);
  if (!start || !end) return '';
  const ms = end.getTime() - start.getTime();
  if (ms < 0) return '';
  return (ms / (1000 * 60 * 60 * 24 * 7)).toFixed(1);
}

function addWeeks(dateStr, weeks) {
  const base = parseYMD(dateStr);
  if (!base) return '';
  const out = new Date(base);
  out.setDate(out.getDate() + (weeks * 7));
  return toYMD(out);
}

// ─── Project Switcher dropdown ───
function renderProjectSwitcher() {
  if (!state.allProjects.length) return null;
  return h('select', {
    className: 'project-switcher',
    onChange: (e) => {
      const proj = state.allProjects.find(p => p.ID === parseInt(e.target.value));
      if (proj) switchProject(proj);
    }
  },
    ...state.allProjects.map(p =>
      h('option', { value: String(p.ID), selected: p.ID === state.projectId }, p.Name)
    )
  );
}

// ─── Version History Panel ───
function renderVersionPanel() {
  if (!state.showVersions) return null;
  const versions = state.versions;
  return h('div', { className: 'tx-panel tx-version-panel' },
    h('div', { className: 'tx-panel-header' },
      h('strong', null, 'Version History'),
      h('button', { className: 'tx-panel-close', onClick: () => {
        state.showVersions = false;
        if (state.transmittal._preview) exitPreview();
        else render();
      }}, '×'),
    ),
    versions === null
      ? h('p', { style: 'padding:12px;color:var(--text-secondary)' }, 'Loading...')
      : versions.length === 0
        ? h('p', { style: 'padding:12px;color:var(--text-secondary)' }, 'No versions yet. Versions are saved automatically as you edit (up to one every 5 minutes).')
        : h('div', { className: 'tx-version-list' },
            ...versions.map(v =>
              h('div', { className: 'tx-version-item' + (state.transmittal._preview === v.id ? ' active' : '') },
                h('div', { className: 'tx-version-info' },
                  h('span', { className: 'tx-version-date' }, fmtDate(v.saved_at)),
                  h('span', { className: 'tx-version-title' }, v.title || '(untitled)'),
                  h('span', { className: 'tx-version-status' }, v.status),
                ),
                h('div', { className: 'tx-version-actions' },
                  h('button', { className: 'btn btn-xs', onClick: () => previewVersion(v.id) }, 'Preview'),
                  h('button', { className: 'btn btn-xs', onClick: () => restoreVersion(v.id) }, 'Restore'),
                ),
              )
            ),
          ),
  );
}

// ─── Duplicate Modal ───
function renderDuplicateModal() {
  if (!state.showDuplicate) return null;
  // Filter to projects that aren't the current one
  const others = state.allProjects.filter(p => p.ID !== state.projectId);
  return h('div', { className: 'tx-modal-overlay', onClick: (e) => {
    if (e.target.classList.contains('tx-modal-overlay')) { state.showDuplicate = false; render(); }
  }},
    h('div', { className: 'tx-modal' },
      h('h3', null, 'Duplicate Transmittal'),
      h('p', { style: 'color:var(--text-secondary);font-size:13px;margin-bottom:12px' },
        'Copy this transmittal to another project. Author, publisher, design, and other house settings are kept. Book-specific fields (title, dates, checklist) are cleared.'
      ),
      others.length === 0
        ? h('p', { style: 'color:var(--text-secondary)' }, 'No other projects available. Create a new project from the calendar first.')
        : h('div', { className: 'tx-duplicate-list' },
            ...others.map(p =>
              h('button', { className: 'btn btn-sm tx-duplicate-item', onClick: () => {
                state.showDuplicate = false;
                render();
                duplicateToProject(p.ID);
              }},
                h('span', null, p.Name),
                h('span', { style: 'color:var(--text-secondary);font-size:11px' }, p.ClientSlug + '/' + p.ProjectSlug),
              )
            ),
          ),
      h('div', { style: 'text-align:right;margin-top:16px' },
        h('button', { className: 'btn btn-sm', onClick: () => { state.showDuplicate = false; render(); } }, 'Cancel'),
      ),
    ),
  );
}

// ─── Main form renderer ───
function renderForm() {
  const d = state.transmittal.data;
  const pct = calcCompletion();
  const calendarUrl = '/' + state.pathClient + '/' + state.pathProject + '/';
  const isPreview = !!state.transmittal._preview;

  return h('div', { className: 'tx-container' },
    // Preview banner
    isPreview ? h('div', { className: 'tx-preview-banner' },
      h('span', null, 'Previewing version from ' + fmtDate(state.transmittal._previewDate)),
      h('button', { className: 'btn btn-sm', onClick: () => restoreVersion(state.transmittal._preview) }, 'Restore this version'),
      h('button', { className: 'btn btn-sm', onClick: exitPreview }, 'Exit preview'),
    ) : null,
    // Header: the factory page has its own; there we draw the section rule instead.
    state.embedded ? renderEmbeddedHead(isPreview) : h('div', { className: 'page-header' },
      h('div', { className: 'page-header-top' },
        h('div', { className: 'page-header-left' },
          h('h1', { className: 'page-header-title' }, state.transmittal?.data?.book?.title || state.project?.Name || 'Transmittal'),
        ),
        h('div', { className: 'page-header-actions' },
          ...headerActions('btn btn-sm', isPreview),
          themeBtn(),
        ),
      ),
      h('div', { className: 'page-header-sub' },
        h('button', { className: 'page-header-back', onClick: () => {
          window.location.href = '/' + state.pathClient + '/';
        }}, '← ' + (state.pathClient || 'Home').toUpperCase()),
        renderProjectSwitcher() || h('span', { style: 'font-size:13px;color:var(--text-secondary)' }, state.project.Name),
        // Calendar link hidden 2026-09-18 (not used by factory customers):
        // h('a', { href: calendarUrl, style: 'font-size:0.8rem;color:var(--accent);text-decoration:none' }, 'Calendar'),
        h('span', { className: 'page-status page-status-' + state.transmittal.status },
          state.transmittal.status
        ),
        h('span', { style: 'font-size:0.78rem;color:var(--text-secondary)' }, 'Autosaves as you edit'),
        h('span', { id: 'tx-save-status', className: 'tx-save-status' }),
      ),
    ),
    // Step strip: the transmittal is step 1 of the factory's five (0.13 A).
    // On the factory page the strip is the factory's own.
    (isPreview || state.embedded) ? null : renderStepStrip(),
    // Hand-off panel once final: what happened, what's next (0.13 C).
    isPreview ? null : renderHandoff(),
    // Email modal
    renderEmailModal(),
    // Version history panel (slides in from right)
    renderVersionPanel(),
    // Duplicate modal
    renderDuplicateModal(),
    // Intro: what this document is and what Mark Final does
    // Once final the hand-off panel says what to do; the intro steps aside.
    (isPreview || state.transmittal.status === 'final') ? null : h('p', { className: 'tx-intro' },
      'The transmittal is the mise en place for your book — the handoff record of what the book is, what’s in the file, and how it should be set, prepared before any typesetting starts. Fill in what you know; leave the rest. When it’s ready, ',
      h('b', null, 'Mark Final'),
      ': that generates your authoring template from it (the ',
      h('b', null, 'Authoring template'),
      state.embedded ? ' link appears beside Mark Final, above' : ' button appears above',
      ') and sends it to the studio; the build follows it. You can switch it back to Draft at any time.',
    ),
    // Completion is shown as a figure in the section head (embedded) or the
    // standalone header; the bare blue line read as an artifact (0.27).
    (isPreview || state.embedded) ? null : h('div', { className: 'tx-progress' },
      h('div', { className: 'tx-progress-bar', style: 'width:' + pct + '%' }),
    ),
    // Two-column layout
    h('div', { className: 'tx-columns' + (isPreview ? ' tx-preview-mode' : '') },
      // LEFT COLUMN — the book and who owns it (0.19, Jenna 2026-09-19:
      // rights + copyright page moved here so the columns balance; measured
      // 3712 / 3373 px at 1280 vs 2660 / 4425 before).
      h('div', { className: 'tx-column' },
        renderBookSection(),
        renderChecklistSection(),
        renderIllustrationsSection(),
        renderPermissionsSection(),
        renderPageIVSection(),
      ),
      // RIGHT COLUMN — how it looks and what comes out. Format leads, then
      // Typography (they belong together), Cover, deliverables.
      h('div', { className: 'tx-column' },
        renderDesignSection(),
        renderEditingSection(),
        renderCoverSection(),
        renderFilesSection(),
      ),
    ),
    // Finish block: the page ends with an action, not a field (0.13 A).
    isPreview ? null : renderFinish(),
  );
}

// ─── Final / Draft toggle (header button + finish block share it) ───
function toggleFinal() {
  const wasDraft = state.transmittal.status !== 'final';
  state.transmittal.status = wasDraft ? 'final' : 'draft';
  scheduleSave();
  // The studio's copy goes out from the server on Mark Final;
  // only the studio itself needs the manual email step here.
  if (wasDraft && !clientMode()) {
    state.showEmail = true;
    state.emailResult = null;
  }
  render();
  emitStatus();
  // The hand-off panel sits under the header; bring it into view.
  if (wasDraft) {
    const top = state.embedded && document.getElementById('transmittal');
    if (top) top.scrollIntoView({ behavior: 'smooth', block: 'start' });
    else window.scrollTo({ top: 0, behavior: 'smooth' });
  }
}

// Print only the transmittal when it lives on the factory page (0.17 C):
// factory.css hides the other sections while body.print-transmittal is set.
function printTransmittal() {
  if (!state.embedded) { window.print(); return; }
  document.body.classList.add('print-transmittal');
  const off = () => document.body.classList.remove('print-transmittal');
  window.addEventListener('afterprint', off, { once: true });
  window.print();
  setTimeout(off, 2000);
}

// Header actions shared by the standalone page header and the embedded
// section rule. `cls` is the class each button/link takes.
function headerActions(cls, isPreview) {
  const isFinal = state.transmittal.status === 'final';
  return [
    h('button', { className: cls, onClick: () => {
      state.showVersions = !state.showVersions;
      if (state.showVersions) { state.versions = null; loadVersions(); }
      else render();
    }}, 'History'),
    // Duplicate copies this transmittal to another project: a studio tool.
    clientMode() ? null : h('button', { className: cls, onClick: () => { state.showDuplicate = true; render(); } }, 'Duplicate'),
    h('button', { className: cls, onClick: printTransmittal }, 'Print'),
    // Customers get no email button (Jenna, 2026-09-18); the admin keeps 'Email'.
    clientMode() ? null : h('button', { className: cls, onClick: () => { state.showEmail = true; render(); } }, 'Email'),
    isFinal && !isPreview
      ? h('a', { className: cls + ' accent', href: '/api/projects/' + state.projectId + '/word-template', download: '',
          title: 'Downloads the authoring template (.docx) generated from this transmittal — optional; any editor\u2019s Heading 1/2 plus [[markers]] works too.' },
          'Template (.docx) \u2193')
      : null,
    h('button', { className: cls + (isFinal ? '' : ' accent'),
      title: isFinal
        ? 'Switches the transmittal back to Draft so you can keep editing.'
        : clientMode()
          ? 'Marks the transmittal final: generates your authoring template and emails you the link. You can switch it back to Draft.'
          : 'Marks the transmittal final: generates your authoring template and opens the email to the studio. You can switch it back to Draft.',
      onClick: toggleFinal,
    }, isFinal ? 'Return to draft' : 'Mark Final'),
  ];
}

// Section rule on the factory page: "// 1 · Transmittal [FINAL] · autosaves"
// left, the actions right (comp-017-C). Replaces the standalone page header.
function renderEmbeddedHead(isPreview) {
  const status = state.transmittal.status;
  return h('div', { className: 'tx-embed-head' + (status === 'final' ? ' is-final' : '') },
    h('div', { className: 'tx-embed-left' },
      h('span', { className: 'kicker' }, '1 \u00b7 Transmittal'),
      h('span', { className: 'tx-embed-status tx-embed-status-' + status }, '[' + String(status).toUpperCase() + ']'),
      h('span', { className: 'tx-embed-note' }, '\u00b7 ' + calcCompletion() + '% filled \u00b7 autosaves as you edit'),
      h('span', { id: 'tx-save-status', className: 'tx-save-status' }),
    ),
    h('div', { className: 'tx-embed-acts' }, ...headerActions('link-action', isPreview)),
  );
}

function factoryUrl(hash) {
  return '/' + state.pathClient + '/' + state.pathProject + '/factory/' + (hash || '');
}

// Same strip as the factory page (classes in theme.css); here step 1 is
// where you are, done once the transmittal is final.
function renderStepStrip() {
  const isFinal = state.transmittal.status === 'final';
  const step = (n, name, href, cls) => h('a', { className: 'fx-step' + (cls ? ' ' + cls : ''), href: href },
    h('span', { className: 'fx-step-n' }, String(n)), h('span', { className: 'fx-step-name' }, name));
  return h('nav', { className: 'fx-steps tx-steps', 'aria-label': 'Where you are' },
    step(1, 'Transmittal', '#', isFinal ? 'done' : 'current'),
    step(2, 'Upload', factoryUrl('#upload')),
    step(3, 'Inspect', factoryUrl('#inspect')),
    step(4, 'Build', factoryUrl('#build')),
    step(5, 'Download', factoryUrl('#download')),
    h('a', { className: 'fx-step fx-step-how', href: '/factory#how-it-runs', target: '_blank', rel: 'noopener',
      title: 'Who does what: Author, Studio, Machine, Out' }, 'How it runs →'),
  );
}

function renderHandoff() {
  if (state.transmittal.status !== 'final') return null;
  const tpl = '/api/projects/' + state.projectId + '/word-template';
  return h('div', { className: 'tx-handoff' },
    h('div', null,
      h('div', { className: 'tx-handoff-k' }, 'Transmittal is final'),
      h('p', null, 'Your authoring template has been generated from it. Two things next:'),
      h('ol', null,
        h('li', null, 'Get your manuscript into shape \u2014 in any editor. ',
          h('b', null, 'Already have a draft?'), ' Chapter titles as Heading 1, sub-heads as Heading 2, and type ',
          h('code', null, '[[quote]]'), ', ', h('code', null, '[[verse]]'), ', ', h('code', null, '[[epigraph]]'), ', ', h('code', null, '[[code]]'), '\u2026 at the start of any special paragraph; the factory applies the style and removes the marker. No template needed. ',
          h('b', null, 'Starting fresh?'), ' ', h('a', { href: tpl, download: '' }, 'Download the template'), ' (.docx for Word', h('span', { className: 'tx-fine' }, ' \u00b7 ', h('a', { href: tpl + '?format=odt', download: '' }, '.odt'), ' for LibreOffice Writer'), ') and write in its styles.'),
        state.embedded
          ? h('li', null, 'When a draft is ready, ', h('a', { href: '#upload' }, 'upload it below'), ' — step 2 — then inspect and build.')
          : h('li', null, 'When a draft is ready, ', h('a', { href: factoryUrl() }, 'go to the Factory'), ' — upload, inspect, build.'),
      ),
    ),
    state.embedded
      ? h('a', { className: 'btn-fill', href: '#upload' }, 'Continue to 2 \u00b7 Upload ', h('span', { className: 'arrow-down', 'aria-hidden': 'true' }, '\u2193'))
      : h('a', { className: 'btn-fill', href: factoryUrl() }, 'Continue to the Factory →'),
  );
}

function renderFinish() {
  const isFinal = state.transmittal.status === 'final';
  return h('div', { className: 'tx-finish' },
    isFinal
      ? h('p', null, h('b', null, 'This transmittal is final.'), ' Edit anything and it autosaves; return to draft mode if you want the studio to wait.')
      : h('p', null, h('b', null, 'That’s the whole form.'), ' Mark it final and the factory generates your authoring template from it; you can return to draft mode at any time.'),
    h('div', { className: 'tx-finish-acts' },
      isFinal
        ? h('button', { className: 'link-action', onClick: toggleFinal }, 'Return to draft mode')
        : null,
      isFinal
        ? (state.embedded
            ? h('a', { className: 'btn-fill', href: '#upload' }, 'Continue to 2 \u00b7 Upload ', h('span', { className: 'arrow-down', 'aria-hidden': 'true' }, '\u2193'))
            : h('a', { className: 'btn-fill', href: factoryUrl() }, 'Continue to the Factory →'))
        : h('button', { className: 'btn-fill', onClick: toggleFinal }, 'Mark Final'),
    ),
  );
}

// ─── Section: Book Info ───
function renderBookSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Book Information'),
    textField('Author', 'book.author'),
    textField('Title', 'book.title'),
    textField('Subtitle', 'book.subtitle'),
    h('div', { className: 'tx-row' },
      selectField('Title Status', 'book.title_status', [['firm','Firm'],['tentative','Tentative']]),
      textField('Series', 'book.series'),
    ),
    h('div', { className: 'tx-row' },
      textField('Publisher', 'book.publisher'),
      textField('In-house Editor', 'book.editor'),
    ),
    h('div', { className: 'tx-row-3' },
      textField('ISBN (paper)', 'book.isbn_paper'),
      textField('ISBN (EPUB)', 'book.isbn_epub'),
      textField('ISBN (cloth)', 'book.isbn_cloth'),
    ),
    // C6: the press-era Production section (Mechs Delivery, Weeks in Prod.,
    // Bound Book Date, a second Transmittal Date) is gone. Print Run moved
    // here, still under its old key so saved transmittals load unchanged;
    // Target date is the one schedule field the factory can act on. Old
    // production.* keys stay in the JSON, just not shown.
    h('div', { className: 'tx-row-3' },
      textField('Transmittal Date', 'book.transmittal_date', { type: 'date' }),
      textField('Target date', 'production.target_date', { type: 'date',
        helpText: 'Optional. When you would like the finished files.' }),
      textField('Print Run', 'production.print_run', { placeholder: 'e.g. 500',
        helpText: 'Optional. For your printer, not the factory.' }),
    ),
  );
}

// ─── Section: Checklist ───
function renderChecklistSection() {
  const d = state.transmittal.data;
  const checklist = d.checklist || [];
  const backmatter = d.backmatter || [];
  const manuscriptStats = renderManuscriptStats();

  function updateChecklistRow(collectionName, collection, index, nextStatus) {
    const item = collection[index];
    item.status = nextStatus;
    if (nextStatus === 'included') {
      item.here_now = true;
      item.to_come_when = '';
    } else if (nextStatus === 'later') {
      item.here_now = false;
    } else if (nextStatus === 'not_in_book') {
      item.here_now = false;
      item.to_come_when = '';
    } else {
      item.here_now = false;
      item.to_come_when = '';
    }
    setField(collectionName, collection);
    render();
  }

  // 0.20 (Jenna, 2026-09-19): one tick per status instead of a dropdown per
  // row — a 20-row list is 20 clicks, not 40, and you can run down a column.
  // Radios give the one-of-three behaviour and keyboard access for free; the
  // CSS draws them as small ticked boxes.
  let tickRow = null; // set per row below so the radio group name is unique
  function tick(value, current, title) {
    const { collectionName, collection, i } = tickRow;
    return h('td', { className: 'tx-tick' },
      h('label', { title },
        h('input', {
          type: 'radio',
          name: 'ck-' + collectionName + '-' + i,
          value,
          checked: current === value ? 'checked' : undefined,
          onChange: () => updateChecklistRow(collectionName, collection, i, value),
        }),
        h('span', { className: 'tx-tick-box', 'aria-hidden': 'true' }),
      ),
    );
  }

  function checklistRow(item, i, collectionName, collection, options = {}) {
    tickRow = { collectionName, collection, i };
    const status = getChecklistItemStatus(item);
    // Generated pages: the factory makes them from this transmittal, so the
    // only question is include or leave out. Blank (and a stale "Coming
    // later") count as included — bookspecs.go reads them the same way.
    if (options.generated) {
      const val = status === 'not_in_book' ? 'not_in_book' : 'included';
      return h('tr', { className: 'generated-row' },
        h('td', { className: 'component-name' }, options.label || item.component),
        tick('included', val, 'Included \u2014 made by the factory'),
        h('td', { className: 'tx-tick tx-muted' }, '\u2014'),
        tick('not_in_book', val, 'Leave out'),
        h('td', { style: 'width:140px', className: 'tx-muted' }, '\u2014'),
      );
    }
    const disabled = status !== 'later';
    return h('tr', null,
      h('td', { className: options.indent ? 'component-indent' : 'component-name' },
        options.label || item.component
      ),
      tick('included', status, 'In the manuscript now'),
      tick('later', status, 'Coming later'),
      tick('not_in_book', status, 'Not in this book'),
      h('td', { style: 'width:140px' },
        h('input', {
          type: 'date',
          value: item.to_come_when || '',
          placeholder: 'Expected date',
          disabled: disabled ? 'disabled' : undefined,
          onChange: (e) => {
            collection[i].status = 'later';
            collection[i].here_now = false;
            collection[i].to_come_when = e.target.value;
            setField(collectionName, collection);
            render();
          }
        })
      ),
    );
  }

  // Rows the factory generates (P4 decisions 2026-09-17: title + © pages,
  // half-title and Contents come from the transmittal, never typed). CIP is
  // the LoC/CIP line on the copyright-page builder now; its row stays in the
  // JSON for old transmittals but is not shown.
  const GENERATED = ['Half title pg', 'Title pg', 'Copyright pg', 'Contents'];
  const HIDDEN = ['CIP'];
  const groupRow = (label) => h('tr', { className: 'group-row' },
    h('td', { colspan: '5', className: 'tx-group-label' }, label));
  const generatedRows = [];
  const typedRows = [];
  checklist.forEach((item, i) => {
    if (HIDDEN.includes(item.component)) return;
    if (GENERATED.includes(item.component)) {
      generatedRows.push(checklistRow(item, i, 'checklist', checklist, { generated: true }));
    } else {
      // Front-matter rows are all flush left (Jenna, 2026-09-18); the stored `indent` flag is ignored.
      typedRows.push(checklistRow(item, i, 'checklist', checklist, { indent: false }));
    }
  });
  const checklistRows = [
    groupRow('Made by the factory from this transmittal \u2014 do not type these'),
    ...generatedRows,
    groupRow('In your manuscript file'),
    ...typedRows,
  ];

  const bmRows = backmatter.map((item, i) =>
    checklistRow(item, i, 'backmatter', backmatter, {
      label: item.component + (item.subtype ? ' (' + item.subtype + ')' : ''),
      indent: false,
    })
  );

  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Manuscript Checklist'),
    h('div', { className: 'tx-help' }, 'The first group is generated \u2014 half-title, title page, copyright page and Contents are built from this form, so leave them included unless your book should not have one. For everything else, say whether it is in the manuscript now, coming later, or not in this book.'),
    h('table', { className: 'tx-checklist' },
      h('thead', null, h('tr', null,
        h('th', null, 'Component'),
        h('th', { className: 'tx-tick' }, 'In ms'),
        h('th', { className: 'tx-tick' }, 'Later'),
        h('th', { className: 'tx-tick' }, 'Not in'),
        h('th', null, 'Expected date'),
      )),
      h('tbody', null,
        ...checklistRows,
        // Stats row. C7: only Parts is typed (it is the parts opt-in for the
        // build); Chapters / Words / Images are read off the newest upload.
        // Old checklist_stats.chapters / words_chars / ms_pp / est_book_pp
        // keys stay in the JSON, just not shown.
        h('tr', { className: 'stats-row' },
          h('td', { colspan: '5' },
            h('div', { className: 'tx-row-3 tx-stats-grid' },
              textField('Parts', 'checklist_stats.parts', {
                placeholder: 'none',
                helpText: '1 or more if the book has parts; blank or “none” if not. With parts, Heading 1 = part, Heading 2 = chapter.',
              }),
              ...manuscriptStats.fields,
            ),
            manuscriptStats.note,
          ),
        ),
        // Back matter header
        h('tr', null, h('td', { colspan: '5', className: 'tx-checklist-subhead' }, 'Back Matter')),
        ...bmRows,
      ),
    ),
  );
}

// ─── Manuscript stats (C7) ───
// Read-only "from your manuscript" values. Three states: no file yet, file
// uploaded but not inspected (or inspected before word counts existed),
// inspected. Never asks the author to type what the factory can count.
function readonlyStat(label, value, help) {
  return h('div', { className: 'tx-field tx-readonly' },
    h('label', null, label),
    h('div', { className: 'tx-readonly-value' + (value == null ? ' tx-readonly-empty' : '') },
      value == null ? '\u2014' : String(value)),
    help ? h('div', { className: 'tx-help' }, help) : null,
  );
}

function renderManuscriptStats() {
  const ms = state.manuscript;
  const factoryUrl = '/' + state.pathClient + '/' + state.pathProject + '/factory/';
  const up = state.embedded ? '#upload' : factoryUrl, insp = state.embedded ? '#inspect' : factoryUrl;
  const where = state.embedded ? 'below (step 2)' : 'factory page';
  const fmtN = (n) => (typeof n === 'number' ? n.toLocaleString('en-US') : null);
  const ok = !!(ms && ms.inspected);
  let note;
  if (!ms) note = 'Checking your manuscript\u2026';
  else if (!ms.hasFile) note = h('span', null, 'Chapters, words and images are counted at upload \u2014 upload your manuscript ', h('a', { href: up }, where), '.');
  else if (!ms.inspected) note = h('span', null, 'Chapters, words and images are counted when you Inspect ', h('b', null, ms.filename || 'your file'), state.embedded ? ' in step 3 ' : ' on the ', h('a', { href: insp }, state.embedded ? 'below' : 'factory page'), '.');
  else if (ms.words == null) note = h('span', null, 'From ', h('b', null, ms.filename || 'your manuscript'), '. ', h('a', { href: insp }, 'Inspect again'), ' for a word count.');
  else note = h('span', null, 'From ', h('b', null, ms.filename || 'your manuscript'), ms.inspectedAt ? ', inspected ' + fmtDate(ms.inspectedAt) : '', '. Chapters are body sections; front and back matter are not counted.');
  return {
    fields: [
      readonlyStat('Chapters', ok ? fmtN(ms.chapters) : null),
      readonlyStat('Words', ok ? fmtN(ms.words) : null),
      readonlyStat('Images', ok ? fmtN(ms.images) : null),
    ],
    note: h('div', { className: 'tx-help tx-stats-note' }, note),
  };
}

// ─── Section: Illustrations ───
// The factory takes images from the manuscript itself; there is no separate
// art upload. The old figures/tables/photos count table is gone — Inspect
// counts inline images at upload and the report lists each one's printed
// size. Old transmittals that saved illustrations.*_no keys still load (the
// keys are simply not shown).
function renderIllustrationsSection() {
  const rule = (text) => h('li', null, text);
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Illustrations'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'Put every figure or photo in the manuscript file itself, where it belongs. The factory carries it into the print PDF and the EPUB; Inspect will tell you how many it found and how big each will print.'),
    h('ul', { className: 'tx-illus-rules' },
      rule('One image per paragraph, inline \u2014 not floating or text-wrapped.'),
      rule('Caption in the paragraph right after the image.'),
      rule('At least 1100 px wide for a full-width figure (about 300 dpi at this trim). Smaller images print smaller, not blurrier.'),
      rule('PNG for line art and screenshots, JPEG for photos.'),
    ),
    h('div', { className: 'tx-help tx-illus-guide' },
      h('b', null, 'Colour. '),
      'Leave your images in colour. The EPUB keeps them that way \u2014 a Paperwhite renders them in grey on the fly, and the same file is in colour on phones, tablets and the Kindle app. The print PDF is a black-and-white interior, so the factory converts each colour image itself rather than leave it to the printer\u2019s machine, which flattens them. A plain grey conversion is flat; the factory\u2019s is what a careful operator does first on most photographs \u2014 luminance grey, auto-level, a gentle S-curve \u2014 and it is fine for four images in five. What it cannot do is the judgement call: mixing the red channel up to lift a face out of foliage, or noticing that a chart\u2019s two lines became the same grey. If an image matters that much, convert it yourself in Photoshop and place the grey version in the Word file; it arrives already grey and the factory leaves it alone. Images that are already black and white are never touched.'),
    h('label', { className: 'tx-check tx-attest' },
      h('input', { type: 'checkbox', checked: getField('illustrations.print_colour') ? 'checked' : undefined,
        onChange: (e) => { setField('illustrations.print_colour', e.target.checked); render(); } }),
      h('span', null, 'Colour interior \u2014 keep the colour images in the print PDF too (a colour-printed book; check with your printer, it costs more per copy).'),
    ),
    textareaField('Art notes', 'illustrations.art_plan', {
      rows: 3,
      helpText: 'Anything the factory should know about the images: placement wishes, cropping, an image still to come.',
    }),
  );
}

// ─── Section: Permissions ───
// C11: the press-era status/date tracking (reprint_status, reprint_when,
// consents_status, consents_when) is gone — the factory does not clear
// rights. One courtesy line, one attestation checkbox, a timestamp. The old
// permissions.* keys stay in the JSON, just not shown.
const ATTESTATION_TEXT = 'Everything in this manuscript is mine or I have permission to reprint it. The factory typesets what I send; clearing rights is my responsibility.';

function renderPermissionsSection() {
  const attested = !!getField('permissions.attested');
  const at = getField('permissions.attested_at');
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Rights'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'A courtesy, not a check: if the book quotes at length, reprints someone else\u2019s work or uses photographs you did not take, make sure you hold the permissions before you build. The factory cannot tell, and does not look. The full terms are on the ',
      h('a', { href: '/factory/terms', target: '_blank', rel: 'noopener' }, 'terms page'), '.'),
    h('label', { className: 'tx-check tx-attest' + (attested ? ' tx-attest-on' : '') },
      h('input', { type: 'checkbox', checked: attested ? 'checked' : undefined,
        onChange: (e) => {
          setField('permissions.attested', e.target.checked);
          setField('permissions.attested_at', e.target.checked ? new Date().toISOString() : '');
          render();
        } }),
      h('span', null, ATTESTATION_TEXT),
    ),
    attested && at ? h('div', { className: 'tx-help' }, 'Attested ' + fmtDate(at) + '.') : null,
  );
}

// ─── Section: Copyright page (C12) ───
// Pub Info & © is now the copyright-page builder: the fields a good p. iv
// needs, and nothing else. The build writes this page (Typst) and the Word
// template shows the same block. Legacy page_iv.credit / other_credit /
// photo_credit keys stay in the JSON; if they hold text they are shown
// read-only below so nothing typed is lost.
const INTERIOR_CREDIT_DEFAULT = 'Typeset by jdbb studio in {typeface}';

// Rights choice (5.22). Mirrors srv/rights.go — the server resolves the same
// wording for the PDF, the Word template and the EPUB's dc:rights.
const RIGHTS_OPTIONS = [
  ['all_rights', 'All rights reserved', null, null],
  ['cc_by', 'CC BY \u2014 attribution', 'Attribution 4.0 International', 'licenses/by/4.0/'],
  ['cc_by_sa', 'CC BY-SA \u2014 attribution, share-alike', 'Attribution-ShareAlike 4.0 International', 'licenses/by-sa/4.0/'],
  ['cc_by_nc', 'CC BY-NC \u2014 attribution, non-commercial', 'Attribution-NonCommercial 4.0 International', 'licenses/by-nc/4.0/'],
  ['cc_by_nc_sa', 'CC BY-NC-SA \u2014 attribution, non-commercial, share-alike', 'Attribution-NonCommercial-ShareAlike 4.0 International', 'licenses/by-nc-sa/4.0/'],
  ['cc_by_nd', 'CC BY-ND \u2014 attribution, no derivatives', 'Attribution-NoDerivatives 4.0 International', 'licenses/by-nd/4.0/'],
  ['cc_by_nc_nd', 'CC BY-NC-ND \u2014 attribution, non-commercial, no derivatives', 'Attribution-NonCommercial-NoDerivatives 4.0 International', 'licenses/by-nc-nd/4.0/'],
  ['cc0', 'CC0 \u2014 public domain dedication', 'CC0 1.0 Universal', 'publicdomain/zero/1.0/'],
];

function rightsLine(code, year, holder) {
  const o = RIGHTS_OPTIONS.find(r => r[0] === code) || RIGHTS_OPTIONS[0];
  const who = [year, holder].filter(Boolean).join(' ');
  const cline = who ? 'Copyright \u00a9 ' + who + '.' : '';
  if (o[0] === 'cc0') {
    return (holder || 'The author') + ' has dedicated this work to the public domain under the Creative Commons CC0 1.0 Universal dedication. To view a copy, visit https://creativecommons.org/' + o[3];
  }
  if (o[3]) {
    const lic = 'This work is licensed under a Creative Commons ' + o[2] + ' License. To view a copy of this license, visit https://creativecommons.org/' + o[3];
    return cline ? cline + ' ' + lic : lic;
  }
  return cline ? cline + ' All rights reserved.' : 'All rights reserved.';
}

function copyrightPageLines() {
  const g = (p) => String(getField(p) || '').trim();
  const lines = [];
  const title = g('book.title');
  if (title) lines.push({ text: title, strong: true });
  const year = g('page_iv.copyright_year'), holder = g('page_iv.held_by') || g('book.author');
  lines.push({ text: rightsLine(g('page_iv.rights'), year, holder) });
  const pub = g('book.publisher'), city = g('page_iv.publisher_city');
  if (pub) lines.push({ text: 'Published by ' + pub + (city ? ', ' + city : '') + '.' });
  if (g('page_iv.edition_line')) lines.push({ text: g('page_iv.edition_line') });
  if (g('cover.credit')) lines.push({ text: g('cover.credit') });
  lines.push({ text: (g('page_iv.interior_credit') || INTERIOR_CREDIT_DEFAULT).replace('{typeface}', 'the book\u2019s typeface'), muted: !g('page_iv.interior_credit') });
  if (g('page_iv.credit')) lines.push({ text: g('page_iv.credit') });
  if (g('page_iv.loc_line')) lines.push({ text: g('page_iv.loc_line') });
  const isbnP = g('book.isbn_paper'), isbnE = g('book.isbn_epub');
  if (isbnP) lines.push({ text: 'ISBN ' + isbnP + ' (paperback)' });
  if (isbnE) lines.push({ text: 'ISBN ' + isbnE + ' (ebook)' });
  if (g('page_iv.additional_notices')) lines.push({ text: g('page_iv.additional_notices'), pre: true });
  if (g('page_iv.printed_in')) lines.push({ text: g('page_iv.printed_in') });
  return lines;
}

function renderCopyrightPreview() {
  return h('div', { id: 'tx-cr-preview', className: 'tx-cr-preview' },
    ...copyrightPageLines().map(l => h('div', {
      className: 'tx-cr-line' + (l.strong ? ' tx-cr-strong' : '') + (l.muted ? ' tx-cr-muted' : '') + (l.pre ? ' tx-cr-pre' : ''),
    }, l.text)),
  );
}

function refreshCopyrightPreview() {
  const el = document.getElementById('tx-cr-preview');
  if (el) el.replaceWith(renderCopyrightPreview());
}

function renderPageIVSection() {
  const legacy = [['Credit line', 'page_iv.credit'], ['Other credit', 'page_iv.other_credit'], ['Photo credit', 'page_iv.photo_credit']]
    .filter(([, k]) => String(getField(k) || '').trim());
  const fromBook = (label, path) => readonlyStat(label, String(getField(path) || '').trim() || null);
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Copyright page'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'Page iv is generated from these fields \u2014 don\u2019t type one in your manuscript. Title, publisher and ISBNs come from Book Information above.'),
    h('div', { className: 'tx-row' },
      textField('Copyright year', 'page_iv.copyright_year', { placeholder: String(new Date().getFullYear()) }),
      textField('Rights holder', 'page_iv.held_by', { placeholder: getField('book.author') || 'Author name',
        helpText: 'Who holds the copyright \u2014 usually the author. Blank uses the author\u2019s name.' }),
    ),
    h('div', { className: 'tx-row' },
      fromBook('Publisher (from Book)', 'book.publisher'),
      textField('Publisher city', 'page_iv.publisher_city', { placeholder: 'e.g. Hong Kong' }),
    ),
    selectField('Rights', 'page_iv.rights', RIGHTS_OPTIONS.map(r => [r[0], r[1]]), {
      helpText: 'Printed after the \u00a9 line, in the wording Creative Commons recommends; the EPUB\u2019s rights metadata matches. All rights reserved unless you choose otherwise.' }),
    textField('Edition / printing line', 'page_iv.edition_line', { placeholder: 'e.g. First edition, 2026' }),
    h('div', { className: 'tx-row' },
      fromBook('ISBN paper (from Book)', 'book.isbn_paper'),
      fromBook('ISBN EPUB (from Book)', 'book.isbn_epub'),
    ),
    textField('Cover design credit', 'cover.credit', { placeholder: 'e.g. Cover design by \u2026',
      helpText: 'Designer or artist to name. Leave blank for none.' }),
    textField('Interior / typesetting credit', 'page_iv.interior_credit', { placeholder: INTERIOR_CREDIT_DEFAULT,
      helpText: 'Blank prints the default; {typeface} is filled with the book\u2019s body typeface at build.' }),
    textField('Library of Congress / CIP line', 'page_iv.loc_line', { placeholder: 'optional \u2014 e.g. Library of Congress Control Number: 2026xxxxxx' }),
    textField('Printed in', 'page_iv.printed_in', { placeholder: 'e.g. Printed in the United States of America' }),
    textareaField('Additional notices', 'page_iv.additional_notices', { rows: 3,
      placeholder: 'Permissions acknowledgements, a disclaimer, a dedication of the type\u2026',
      helpText: 'Printed as typed, after the ISBNs.' }),
    legacy.length ? h('div', { className: 'tx-legacy' },
      h('div', { className: 'tx-help' }, 'From the earlier form (not printed \u2014 move what you still want into Additional notices):'),
      ...legacy.map(([label, k]) => h('div', { className: 'tx-legacy-line' }, h('b', null, label + ': '), getField(k))),
    ) : null,
    h('div', { className: 'tx-field' },
      h('label', null, 'Preview'),
      renderCopyrightPreview(),
    ),
  );
}

// Subrights (copub / marketing pages) deleted 2026-09-18: press-only.
// Saved subrights.* values still ride along in the JSON.

// ─── Custom style row (docs/reviews/CUSTOM-STYLES-MARKERS-2026-09-22.md) ───
// Mirrors srv/customstyles.go: the parents, which parents take which delta,
// and the resolved-form wording. Keep the two in step.
const CUSTOM_STYLE_PARENTS = ['Normal', 'First Paragraph', 'Heading 1', 'Heading 2', 'Heading 3',
  'Block Quote', 'Epigraph', 'Verse', 'Code Block', 'Section Break', 'Copyright', 'Signature', 'Glossary Entry'];
const CUSTOM_STYLE_CHAR_PARENTS = [['none', 'Nothing (plain text)'], ['italic', 'Italic'], ['small caps', 'Small caps'], ['code', 'Code (monospace)']];
const CUSTOM_STYLE_INDENT_PARENTS = new Set(['Verse', 'Block Quote', 'Code Block']);
const CUSTOM_STYLE_SPACE_PARENTS = new Set(['Section Break']);
const DESIGNED_PRESETS = new Set(['tweet-block', 'metadata-paragraph', 'metadata-inline', 'ascii-block']);

function customStyleResolvedForm(style) {
  const name = (style.name || '').trim();
  if (!name) return '';
  const type = style.type === 'character' ? 'character' : 'paragraph';
  const typst = (style.typst || '').trim();
  // A snippet or designed preset the studio wrote (read-only here; admin edits it).
  if (DESIGNED_PRESETS.has(style.preset)) return `${name} \u2192 designed by the studio (preset: ${style.preset})`;
  if (typst && !/^#let [a-z0-9-]+\(content\) = \{\s*content\s*\}$/.test(typst) && !/text\(font: config\.body-font, size: 0\.9em\)\[#content\]/.test(typst)) {
    return `${name} \u2192 designed by the studio (custom typesetting)`;
  }
  if (type === 'character') {
    const p = style.based_on || 'none';
    return p === 'none' ? `${name} \u2192 plain text (no change)` : `${name} \u2192 ${p}`;
  }
  const parent = CUSTOM_STYLE_PARENTS.includes(style.based_on) ? style.based_on : 'Normal';
  let out = `${name} \u2192 ${parent}`;
  const indent = CUSTOM_STYLE_INDENT_PARENTS.has(parent) ? parseInt(style.indent || 0, 10) : 0;
  if (indent === 1) out += ', indent one level';
  if (indent === 2) out += ', indent two levels';
  const space = CUSTOM_STYLE_SPACE_PARENTS.has(parent) ? parseInt(style.space_before || 1, 10) : 1;
  if (space === 3) out += ', 3\u00d7 the stanza gap before';
  if (space === 6) out += ', 6\u00d7 the stanza gap before';
  return out;
}

function renderCustomStyleRow(style, i, removeCustomStyle) {
  const type = style.type === 'character' ? 'character' : 'paragraph';
  const parent = type === 'character'
    ? (style.based_on || 'none')
    : (CUSTOM_STYLE_PARENTS.includes(style.based_on) ? style.based_on : 'Normal');
  const base = `custom_styles.${i}`;
  const onParentChange = (e) => {
    const s = state.transmittal.data.custom_styles[i];
    s.based_on = e.target.value;
    if (!CUSTOM_STYLE_INDENT_PARENTS.has(s.based_on)) delete s.indent;
    if (!CUSTOM_STYLE_SPACE_PARENTS.has(s.based_on)) delete s.space_before;
    setField(`${base}.based_on`, s.based_on);
    render();
  };
  const onTypeChange = (e) => {
    const s = state.transmittal.data.custom_styles[i];
    s.type = e.target.value;
    s.based_on = s.type === 'character' ? 'none' : 'Normal';
    delete s.indent; delete s.space_before;
    setField(`${base}.type`, s.type);
    render();
  };
  const basedOnOptions = type === 'character'
    ? CUSTOM_STYLE_CHAR_PARENTS
    : CUSTOM_STYLE_PARENTS.map(p => [p, p]);
  const basedOn = h('div', { className: 'tx-field' },
    h('label', null, 'Based on'),
    h('select', { onChange: onParentChange }, ...basedOnOptions.map(([v, l]) => {
      const o = h('option', { value: v }, l); if (v === parent) o.selected = true; return o;
    })));
  const deltas = [];
  if (type === 'paragraph' && CUSTOM_STYLE_INDENT_PARENTS.has(parent)) {
    deltas.push(selectField('Indent', `${base}.indent`, [
      ['0', 'none'], ['1', 'one level (1.5 em)'], ['2', 'two levels (3 em)'],
    ]));
  }
  if (type === 'paragraph' && CUSTOM_STYLE_SPACE_PARENTS.has(parent)) {
    deltas.push(selectField('Space before', `${base}.space_before`, [
      ['1', 'as usual'], ['3', '3\u00d7 the stanza gap'], ['6', '6\u00d7 the stanza gap'],
    ]));
  }
  const resolved = customStyleResolvedForm(style);
  return h('div', { className: 'tx-custom-style' },
    h('div', { className: 'tx-row-3' },
      textField('Style name', `${base}.name`, { placeholder: 'e.g. verse2' }),
      h('div', { className: 'tx-field' },
        h('label', null, 'Type'),
        h('select', { onChange: onTypeChange },
          ...[['paragraph', 'Paragraph'], ['character', 'Character (inside a line)']].map(([v, l]) => {
            const o = h('option', { value: v }, l); if (v === type) o.selected = true; return o;
          }))),
      basedOn,
    ),
    h('div', { className: 'tx-row-3' },
      ...deltas,
      textField('Purpose', `${base}.description`, { placeholder: type === 'character' ? 'e.g. commands typed at the terminal' : 'e.g. second indent level in a poem' }),
    ),
    resolved ? h('div', { className: 'tx-help tx-custom-style-resolved' }, 'In the book: ', h('strong', null, resolved)) : null,
    h('button', { className: 'tx-reviewer-remove', type: 'button', onClick: () => removeCustomStyle(i) }, 'Remove')
  );
}

// ─── Section: Editing ───
function renderEditingSection() {
  const styles = state.transmittal.data.custom_styles || [];

  function duplicateCustomStyleName(nextStyles) {
    const seen = new Set();
    for (const item of nextStyles) {
      const key = (item && item.name ? String(item.name) : '').trim().toLowerCase();
      if (!key) continue;
      if (seen.has(key)) return key;
      seen.add(key);
    }
    return '';
  }

  function updateCustomStyles(nextStyles) {
    const duplicate = duplicateCustomStyleName(nextStyles);
    if (duplicate) {
      alert(`Duplicate custom style name: ${duplicate}. Use distinct names such as metadata-p and metadata-c.`);
      return;
    }
    setField('custom_styles', nextStyles);
    render();
  }

  function addCustomStyle() {
    const next = [...styles, { name: '', type: 'paragraph', based_on: 'Normal', description: '' }];
    updateCustomStyles(next);
  }

  function removeCustomStyle(index) {
    const next = styles.filter((_, i) => i !== index);
    updateCustomStyles(next);
  }

  // C10: Developmental Edit and Level of Copyediting (with their instruction
  // boxes) are gone — the factory does not sell editing. What is left is
  // what drives the template and the typography. Old editing.* keys
  // (developmental_edit, developmental_instructions, copyediting_level,
  // instructions) stay in the JSON, just not shown.
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Typography'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'Four choices are yours; margins, running heads and the rest follow from the trim and the series design. Each row\u2019s first answer is the studio default \u2014 leave them if you have no preference.'),
    ...renderTypographyChoices(),
    h('div', { className: 'tx-section-header', style: 'margin-top:16px' }, 'Typography notes'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'The factory typesets what you send; it does not edit. Tell it here about anything in the text that needs special handling in type.'),
    textField('Special Characters', 'editing.special_characters', {
      placeholder: 'e.g. Greek, IPA, accented names, arrows',
      helpText: 'Scripts, symbols or diacritics beyond ordinary English, so the right fonts are checked before the build.',
    }),
    textField('Mathematical Formulas', 'editing.math_formulas', {
      placeholder: 'none',
      helpText: 'Inline symbols, or displayed equations? Word\u2019s equation editor, or typed?',
    }),
    h('div', { className: 'tx-section-header', style: 'margin-top:16px' }, 'Custom Styles'),
    h('div', { className: 'tx-help' }, 'A style you use in the manuscript that is not one of the factory\u2019s. Name it, say which factory style it is based on, and (for a few parents) how it differs: an indent level, or a bigger gap. It then goes into the book spec, the Word template, and the [[marker]] list Inspect accepts. Anything a based-on style cannot express is designed by the studio and shown here once it is.'),
    ...styles.map((style, i) => renderCustomStyleRow(style, i, removeCustomStyle)),
    h('button', { className: 'tx-add-btn', type: 'button', onClick: addCustomStyle }, '+ Add custom style'),
  );
}

// ─── Typography choices (punch list 0.8 part 2) ───
// Four radio rows, stored under transmittal.typography and mirrored into the
// book spec by the server (srv/typochoices.go). Keys and defaults must match
// that file: pairing studio|classic|house|literary · size compact|standard|
// generous · section_break space|breve|ornament|custom (+ section_break_text) ·
// paragraphs indented|block.
const TYPO_PAIRINGS = [
  { value: 'studio', name: 'Studio\u2019s choice', blurb: 'We pick for the manuscript \u2014 today that is the open classic pairing below.' },
  { value: 'classic', name: 'Open classic', blurb: 'Libertinus Serif for the text, Source Sans for headings. Even, quiet, reads well at any size.' },
  { value: 'house', name: 'Studio house', blurb: 'Plantin for the text, Proxima Nova for headings \u2014 the jdbb series look. Warm, slightly dark on the page.' },
  { value: 'literary', name: 'Literary', blurb: 'EB Garamond for text and headings alike. Light, old-style, fiction and essays.' },
];
const TYPO_SIZES = [
  { value: 'compact', name: 'Compact', factor: 0.95, blurb: 'A little smaller; more words to the page, so a shorter book.' },
  { value: 'standard', name: 'Standard', factor: 1, blurb: 'The size the trim calls for.' },
  { value: 'generous', name: 'Generous', factor: 1.06, blurb: 'A little larger; easier on the eyes, a longer book.' },
];
const TYPO_BREAKS = [
  { value: 'space', name: 'White space', blurb: 'A blank line between scenes. Quiet; can be missed at the foot of a page.' },
  { value: 'breve', name: 'Breve', blurb: 'Three small centred marks (\u02d8 \u02d8 \u02d8) \u2014 the studio\u2019s house mark.' },
  { value: 'ornament', name: 'Ornament', blurb: 'A single centred fleuron (\u2767).' },
  { value: 'custom', name: 'Your own', blurb: 'Type the mark you want centred between scenes \u2014 an asterism (\u2042), three asterisks, a tilde, a word.' },
];
const TYPO_PARAS = [
  { value: 'indented', name: 'Indented', blurb: 'First line indented, no space between paragraphs. The book convention.' },
  { value: 'block', name: 'Block', blurb: 'No indent; half a line of space between paragraphs. Manuals, workbooks, some nonfiction.' },
];

// Trim → body size and a rough words-a-page figure, mirroring the server's
// trim-derived defaults (srv/typodefaults.go): margins are fixed fractions
// of the trim, 10.5pt from 5\u00bd in wide, 10pt below; leading 1.28\u00d7.
function typoTrimDims() {
  const t = String(getField('design.trim') || '');
  const m = t.match(/^\s*(\d+(?:\.\d+)?)\s*[x\u00d7]\s*(\d+(?:\.\d+)?)\s*$/i);
  if (m) return { w: parseFloat(m[1]), h: parseFloat(m[2]), assumed: false };
  return { w: 6, h: 9, assumed: true };
}
function typoSizeDescription(opt) {
  const d = typoTrimDims();
  const base = d.w >= 5.5 ? 10.5 : 10;
  const pt = Math.round(base * opt.factor * 4) / 4;
  const lead = Math.round(pt * 1.28 * 10) / 10;
  const textW = d.w * (1 - 0.146 - 0.125) * 72, textH = d.h * (1 - 0.089 - 0.111) * 72;
  const charsPerLine = textW / pt * 2.3, lines = Math.floor(textH / (pt * 1.28));
  const words = Math.round(charsPerLine * lines * 0.85 / 6 / 10) * 10;
  return `${pt} on ${lead} pt, roughly ${words} words a page` + (d.assumed ? ' at 6 \u00d7 9' : '') + '.';
}

function renderTypographyChoices() {
  const row = (label, path, name, options, extra) => {
    const cur = String(getField(path) || options[0].value);
    const pick = (v) => () => { setField(path, v); render(); };
    return h('div', { className: 'tx-field' },
      h('label', null, label),
      h('div', { className: 'tx-trim-tiers tx-typo-rows' },
        ...options.map(o => {
          const on = cur === o.value;
          return h('label', { className: 'tx-check tx-trim-tier' + (on ? ' is-on' : '') },
            h('input', { type: 'radio', name, value: o.value, checked: on ? 'checked' : undefined, onChange: pick(o.value) }),
            h('span', { className: 'tx-trim-tier-text' },
              h('strong', null, o.name), ' \u2014 ',
              h('span', { className: 'tx-trim-blurb' }, extra ? extra(o) + ' ' + o.blurb : o.blurb)));
        })));
  };
  return [
    row('Typeface', 'typography.pairing', 'typo-pairing', TYPO_PAIRINGS),
    renderTypoSamples(),
    row('Text size', 'typography.size', 'typo-size', TYPO_SIZES, typoSizeDescription),
    row('Section breaks', 'typography.section_break', 'typo-break', TYPO_BREAKS),
    (getField('typography.section_break') === 'custom')
      ? textField('Your break mark', 'typography.section_break_text',
          { placeholder: '* * *', className: 'tx-typo-custom', maxLength: 24,
            helpText: 'Set in the text face, centred, with a line of space above and below. Up to 24 characters; leave empty for the breve.' })
      : null,
    row('Paragraphs', 'typography.paragraphs', 'typo-paras', TYPO_PARAS),
  ];
}

// ─── Typography sampler chips (punch list 0.8 part 3) ───
// One text page per pairing (recto 3 of the sampler, same story in each),
// rendered at 2× by typesetting/scripts/build-sampler.sh into
// srv/static/samples/. Click enlarges in a plain full-screen overlay (no
// modal lib on this page); click anywhere or Escape closes. Hidden in print.
const TYPO_SAMPLES = [
  { key: 'classic', name: 'Open classic', faces: 'Libertinus Serif · Source Sans 3' },
  { key: 'house', name: 'Studio house', faces: 'Plantin · Proxima Nova' },
  { key: 'literary', name: 'Literary', faces: 'EB Garamond' },
];
const TYPO_SAMPLES_V = '20260919';
function typoSampleUrl(file) { return '/static/samples/' + file + '?v=' + TYPO_SAMPLES_V; }

function openTypoLightbox(sample) {
  closeTypoLightbox();
  const onKey = (e) => { if (e.key === 'Escape') closeTypoLightbox(); };
  const box = h('div', { className: 'tx-lightbox', role: 'dialog', 'aria-label': sample.name + ' sample page', onClick: closeTypoLightbox },
    h('img', { src: typoSampleUrl('page-' + sample.key + '.png'), alt: 'A text page set in the ' + sample.name + ' pairing (' + sample.faces + ')' }),
    h('div', { className: 'tx-lightbox-caption' },
      sample.name + ' \u00b7 ' + sample.faces + ' \u00b7 6 \u00d7 9 in, standard size',
      h('span', { className: 'tx-lightbox-hint' }, 'click or Esc to close')));
  box._onKey = onKey;
  document.addEventListener('keydown', onKey);
  document.body.appendChild(box);
}
function closeTypoLightbox() {
  const box = document.querySelector('.tx-lightbox');
  if (!box) return;
  if (box._onKey) document.removeEventListener('keydown', box._onKey);
  box.remove();
}

function renderTypoSamples() {
  return h('div', { className: 'tx-typo-samples' },
    h('div', { className: 'tx-typo-chips' },
      ...TYPO_SAMPLES.map(s => h('button', {
        type: 'button', className: 'tx-typo-chip',
        title: 'Enlarge the ' + s.name + ' sample page',
        onClick: () => openTypoLightbox(s),
      },
        h('img', { src: typoSampleUrl('chip-' + s.key + '.png'), width: 240, height: 360, loading: 'lazy',
          alt: 'Text page in the ' + s.name + ' pairing' }),
        h('span', { className: 'tx-typo-chip-name' }, s.name),
        h('span', { className: 'tx-typo-chip-faces' }, s.faces)))),
    h('div', { className: 'tx-typo-samples-foot' },
      h('span', { className: 'tx-typo-samples-note' }, 'Same page, three settings \u2014 click a page to enlarge.'),
      h('a', { className: 'link-action', href: typoSampleUrl('typography-sampler.pdf'), target: '_blank', rel: 'noopener' },
        'Download the full sampler (PDF)')));
}

// ─── Section: Format ───
// Was "Book Design", a production editor's worksheet: trim radios with a
// DON'T CARE, est. pages, PPI, spine width, a jdbb complexity tier, outside
// designer, reuse-previous. The factory needs one physical decision from
// the author — the page size — and the spine belongs to the printer, who
// alone knows the paper. Old keys (design.est_pages, .ppi, .spine_width,
// .complexity, .outside_designer, .reuse_previous) still load and save;
// they just no longer have inputs.
const TRIM_TIERS = [
  { value: '5.5 x 8.5', name: 'Small', size: '5\u00bd \u00d7 8\u00bd in',
    blurb: 'Fits a hand or a coat pocket. Fiction, essays, poetry. Fewer words to a page, so a longer book.' },
  { value: '6 x 9', name: 'Medium', size: '6 \u00d7 9 in',
    blurb: 'The standard trade paperback. Nonfiction, memoir, anything with notes or the occasional figure. If you are unsure, choose this.' },
  { value: '8.5 x 11', name: 'Large', size: '8\u00bd \u00d7 11 in',
    blurb: 'Workbooks, manuals, wide tables, many images. Heavy in the hand; not a book for reading in bed.' },
];
const TRIM_STUDIO = 'studio';            // legacy 'dont_care' shows as this
const TRIM_STUDIO_ALIASES = [TRIM_STUDIO, 'dont_care'];

function renderDesignSection() {
  const trimVal = String(getField('design.trim') || '');
  const isStudio = TRIM_STUDIO_ALIASES.includes(trimVal);
  const isTier = TRIM_TIERS.some(t => t.value === trimVal);
  const isOther = trimVal !== '' && !isStudio && !isTier;
  const radio = (value, checked, onChange, ...label) =>
    h('label', { className: 'tx-check tx-trim-tier' + (checked ? ' is-on' : '') },
      h('input', { type: 'radio', name: 'trim', value, checked: checked ? 'checked' : undefined, onChange }),
      h('span', { className: 'tx-trim-tier-text' }, ...label));
  const pick = (v) => () => { setField('design.trim', v); render(); };

  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Format'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'One physical decision is yours: the page size, which printers call the trim. Margins, type size and running heads follow from the series design.'),
    textareaField('What kind of book is it, as an object?', 'design.trim_guidance', {
      rows: 2,
      className: 'tx-field-important',
      placeholder: 'e.g. a paperback novel \u00b7 a workbook people write in \u00b7 a small gift book \u00b7 a reference with wide tables',
      helpText: 'Say it in words. It lets us sanity-check the size you choose below.',
    }),
    h('div', { className: 'tx-field' },
      h('label', null, 'Trim size'),
      h('div', { className: 'tx-trim-tiers' },
        ...TRIM_TIERS.map(t => radio(t.value, trimVal === t.value, pick(t.value),
          h('strong', null, t.name), ' \u2014 ', t.size, '. ',
          h('span', { className: 'tx-trim-blurb' }, t.blurb))),
        radio(TRIM_STUDIO, isStudio, pick(TRIM_STUDIO),
          h('strong', null, 'Let the studio choose'), '. ',
          h('span', { className: 'tx-trim-blurb' }, 'We pick from the manuscript and your note above, and tell you what we chose.')),
        radio('other', isOther, pick('7 x 10'),
          h('strong', null, 'Exact size'), isOther ? ': ' : '. ',
          isOther
            ? h('input', { type: 'text', className: 'tx-trim-other-input', value: trimVal,
                placeholder: '7 x 10', onInput: (e) => setField('design.trim', e.target.value) })
            : null,
          h('span', { className: 'tx-trim-blurb' }, ' When a printer, a series or a distributor already fixes the size. Width \u00d7 height in inches, e.g. 7 x 10.')),
      ),
    ),
    h('div', { className: 'tx-field' },
      h('label', null, 'Spine, paper and the cover template'),
      h('div', { className: 'tx-help tx-illus-guide' },
        'A book\u2019s spine width depends on the PPI, pages per inch, of the paper from your printer that you choose. When your interior PDF is final, you give the printer the trim and the page count; they send back a cover template your cover designer will use with the front-spine-back set up correctly.'),
    ),
    textareaField('Format notes', 'design.freeform_notes', {
      rows: 3,
      placeholder: 'e.g. match the look of my previous book with you \u00b7 a designer is involved \u00b7 a size or paper your printer has already fixed',
    }),
  );
}

// Press-era fields (paper, colours, who does front/spine/back) came out with
// the factory: the pass makes an interior, not a jacket. What's left is the
// split of responsibilities and the two things the interior does need.
function renderCoverSection() {
  const rule = (text) => h('li', null, text);
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Cover'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'The factory builds the inside of the book. The cover is yours to design or commission \u2014 here is how the two meet.'),
    h('ul', { className: 'tx-illus-rules' },
      rule('What you get: an EPUB with your front cover embedded, so the book shows its face in readers\u2019 libraries; and a print-ready interior PDF at your trim size, with the page count your cover designer needs for the spine.'),
      rule('What you bring: one front-cover image (JPEG or PNG, portrait, at least 1600 \u00d7 2400 px), uploaded on the factory page under step 2. Replace it any time; the next build picks it up.'),
      rule('For print, your cover is a separate file that goes straight to the printer \u2014 they publish a template once they know your trim, page count and paper. Nothing to upload here for that.'),
    ),
    // Cover credit moved to the Copyright page section (C12); same key.
    textareaField('Cover notes', 'cover.production_plan_budget', {
      rows: 2,
      helpText: 'Anything about the cover the factory should know \u2014 a designer still working, an image to come.',
    }),
  );
}

// ─── Section: What you get ───
// Was Deliverables (checkboxes incl. Typst source, fonts, cover files) plus a
// Page Proofs reviewer list and a PDF/X printer radio. Decision 2026-09-17
// (RUN C13): every pass yields the same four things; Typst source stays with
// the factory (house template + filters, useless without the toolchain);
// fonts never (licensed); cover files never (theirs); no proof routing —
// the author downloads the PDF and reads it. Old files.* / proofs.* keys
// still ride along in the JSON.
function renderFilesSection() {
  const rule = (...c) => h('li', null, ...c);
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'What you get'),
    h('div', { className: 'tx-help tx-illus-guide' },
      'Every pass produces the same four things, on the factory page, every time you build:'),
    h('ul', { className: 'tx-illus-rules' },
      rule(h('strong', null, 'Print-interior PDF'), ' at your trim size — an RGB PDF; your printer converts to their colour profile.'),
      rule(h('strong', null, 'EPUB'), ' with your front cover embedded — the file Kindle, Apple Books and Kobo want.'),
      rule(h('strong', null, 'Authoring template'), ' (.docx / .odt) generated from this transmittal, with the factory styles and your copyright page in place.'),
      rule(h('strong', null, 'Inspect report'), ' — what the factory found in your manuscript and what to fix.'),
    ),
    h('div', { className: 'tx-help tx-illus-guide' },
      'Not included: the Typst source (the house template and filters — they stay with the factory), the fonts (licensed), and cover files (yours). There is no proof-routing step: download the PDF and read it; rebuild as often as you like while your pass is live.'),
  );
}

// ─── Email Modal ───

let emailRecipients = null; // initialized on first open

// A signed-in pass holder (not the studio) is looking at their own transmittal.
function clientMode() { return !state.isAdmin && !!state.passEmail; }

function initEmailRecipients() {
  if (emailRecipients) return;
  // Pass holders: "me" first, studio opt-in. The studio already hears
  // about Mark Final through its own notification.
  if (clientMode()) {
    emailRecipients = [
      { email: state.passEmail, label: 'You', checked: true, editable: false },
      { email: contactEmail, label: 'Studio', checked: false, editable: false },
      { email: '', label: 'Other', checked: false, editable: true },
    ];
    return;
  }
  emailRecipients = [
    { email: 'jdbb@agentmail.to', label: 'JDBB Archive', checked: true, editable: false },
    { email: contactEmail, label: 'Studio', checked: true, editable: false },
    { email: '', label: 'Other', checked: false, editable: true },
  ];
}

async function checkEmailConfig() {
  if (state.emailConfigured !== null) return;
  try {
    const r = await api('/api/email/status');
    state.emailConfigured = r.configured;
  } catch {
    state.emailConfigured = false;
  }
}

async function sendTransmittalEmail() {
  const recipients = emailRecipients
    .filter(r => r.checked && r.email.trim())
    .map(r => r.email.trim());

  if (recipients.length === 0) {
    state.emailResult = { error: 'Select at least one recipient' };
    render();
    return;
  }

  state.emailSending = true;
  state.emailResult = null;
  render();

  try {
    const res = await api('/api/projects/' + state.projectId + '/transmittal/email', {
      method: 'POST',
      body: JSON.stringify({ recipients }),
    });
    state.emailSending = false;
    state.emailResult = { ok: true, sent_to: res.sent_to };
    render();
  } catch (e) {
    state.emailSending = false;
    state.emailResult = { error: e.message };
    render();
  }
}

function renderEmailModal() {
  if (!state.showEmail) return h('div');
  initEmailRecipients();
  checkEmailConfig();

  const title = state.transmittal?.data?.book?.title || 'Untitled';
  const status = state.transmittal?.status || 'draft';

  const closeModal = () => {
    state.showEmail = false;
    state.emailResult = null;
    render();
  };

  return h('div', { className: 'tx-modal-overlay', onClick: (e) => { if (e.target.classList.contains('tx-modal-overlay')) closeModal(); } },
    h('div', { className: 'tx-modal email-modal' },
      h('div', { className: 'tx-modal-header' },
        h('h2', null, clientMode() ? 'Email me a copy' : 'Email Transmittal'),
        h('button', { className: 'tx-modal-close', onClick: closeModal }, '×'),
      ),
      h('div', { className: 'tx-modal-body' },
        // Status line
        h('div', { className: 'email-summary' },
          h('strong', null, title),
          ' · ',
          h('span', { className: 'tx-status tx-status-' + status }, status),
        ),

        state.emailConfigured === false
          ? h('div', { className: 'email-warning' },
              'Email is not configured on the server. ',
              'Set PRODCAL_MAIL_FROM (Resend) or AGENTMAIL_API_KEY + AGENTMAIL_INBOX_ID on the server.'
            )
          : null,

        // Recipients
        h('div', { className: 'email-recipients' },
          h('label', { className: 'email-label' }, 'Send to:'),
          ...emailRecipients.map((r, i) =>
            h('div', { className: 'email-recipient-row' },
              h('input', {
                type: 'checkbox',
                checked: r.checked ? 'checked' : undefined,
                onChange: () => { emailRecipients[i].checked = !emailRecipients[i].checked; render(); },
              }),
              r.editable
                ? h('input', {
                    type: 'email',
                    className: 'email-input',
                    placeholder: 'email@example.com',
                    value: r.email,
                    onInput: (e) => {
                      emailRecipients[i].email = e.target.value;
                      emailRecipients[i].checked = e.target.value.trim().length > 0;
                    },
                    onFocus: () => {
                      // Auto-add another row if this is the last editable one
                      const editables = emailRecipients.filter(x => x.editable);
                      if (editables.indexOf(r) === editables.length - 1 && r.email.trim()) {
                        emailRecipients.push({ email: '', label: 'Other', checked: false, editable: true });
                        render();
                      }
                    },
                  })
                : h('span', { className: 'email-addr' }, r.email),
              h('span', { className: 'email-recipient-label' }, r.label),
            )
          ),
        ),

        // Result
        state.emailResult?.ok
          ? h('div', { className: 'email-success' },
              '✓ Sent to: ' + state.emailResult.sent_to.join(', ')
            )
          : null,
        state.emailResult?.error
          ? h('div', { className: 'email-error' },
              '' + state.emailResult.error
            )
          : null,

        // Actions
        h('div', { className: 'email-actions' },
          state.emailResult?.ok
            ? h('button', { className: 'btn btn-primary', onClick: closeModal }, '✓ Done')
            : [
                h('button', {
                  className: 'btn btn-primary',
                  disabled: state.emailSending || state.emailConfigured === false ? 'disabled' : undefined,
                  onClick: sendTransmittalEmail,
                }, state.emailSending ? 'Sending…' : 'Send Email'),
                h('button', { className: 'btn btn-sm', onClick: closeModal }, 'Cancel'),
              ],
        ),
      ),
    ),
  );
}
