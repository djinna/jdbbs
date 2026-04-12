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

// ─── Font + Theme System ───
var _fonts = {
  'ibm-serif': { family: "'IBM Plex Serif',Georgia,serif", label: 'IBM Plex Serif' },
  'ibm-sans':  { family: "'IBM Plex Sans',-apple-system,sans-serif", label: 'IBM Plex Sans' },
  literata:    { family: "'Literata',Georgia,serif", label: 'Literata' },
  menlo:       { family: "'Menlo','Consolas','Monaco',monospace", label: 'Menlo' },
};
var _fontKeys = ['literata', 'ibm-serif', 'menlo', 'ibm-sans'];
var _themeState = { font: _fontKeys[Math.floor(Math.random() * _fontKeys.length)], dark: false };
try { var _saved = JSON.parse(localStorage.getItem('prodcal-theme-v1')); if (_saved) _themeState.dark = _saved.dark; } catch(e) {}

function _applyTheme() {
  var f = _fonts[_themeState.font];
  if (f) {
    document.body.style.fontFamily = f.family;
    var nameEl = document.getElementById('font-name');
    if (nameEl) nameEl.textContent = f.label;
  }
  document.documentElement.classList.toggle('dark', _themeState.dark);
  document.querySelectorAll('.theme-opt[data-font]').forEach(function(b) {
    b.classList.toggle('active', b.dataset.font === _themeState.font);
  });
  var darkBtn = document.getElementById('dark-btn');
  if (darkBtn) darkBtn.textContent = _themeState.dark ? '\u2600' : '\u263e';
}
function _saveTheme() {
  try { localStorage.setItem('prodcal-theme-v1', JSON.stringify({ dark: _themeState.dark })); } catch(e) {}
}
function _setFont(key) { _themeState.font = key; _applyTheme(); _saveTheme(); }
function _toggleDark() { _themeState.dark = !_themeState.dark; _applyTheme(); _saveTheme(); }

function _ensureThemeBar() {
  if (document.getElementById('theme-bar')) return;
  var bar = document.createElement('div');
  bar.className = 'theme-bar';
  bar.id = 'theme-bar';
  bar.innerHTML = '<div class="font-name" id="font-name" title="Click to change typeface">' + (_fonts[_themeState.font]?.label || '') + '</div>'
    + '<div class="font-options" id="font-options">'
    + '<button class="theme-opt" data-font="literata">Literata</button>'
    + '<button class="theme-opt" data-font="ibm-serif">IBM Plex Serif</button>'
    + '<button class="theme-opt" data-font="menlo">Menlo</button>'
    + '<button class="theme-opt" data-font="ibm-sans">IBM Plex Sans</button>'
    + '</div>'
    + '<div class="theme-sep"></div>'
    + '<button class="dark-btn" id="dark-btn" title="Toggle dark mode">\u263e</button>';
  document.body.appendChild(bar);
  var expanded = false;
  var nameEl = document.getElementById('font-name');
  nameEl.addEventListener('click', function() { expanded = !expanded; bar.classList.toggle('expanded', expanded); });
  document.addEventListener('click', function(e) { if (expanded && !bar.contains(e.target)) { expanded = false; bar.classList.remove('expanded'); } });
  bar.querySelectorAll('.theme-opt[data-font]').forEach(function(btn) {
    btn.addEventListener('click', function() { _setFont(this.dataset.font); expanded = false; bar.classList.remove('expanded'); });
  });
  document.getElementById('dark-btn').addEventListener('click', _toggleDark);
  _applyTheme();
}
function themeBtn() { return h('span'); }
function getTheme() { return _themeState.dark ? 'dark' : 'light'; }

// ─── State ───
let state = {
  view: 'loading', // loading, auth, form
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
  const app = $('#app');
  app.innerHTML = '';
  if (state.view === 'loading') app.appendChild(h('div', { className: 'tx-container' }, h('p', null, 'Loading...')));
  else if (state.view === 'auth') app.appendChild(renderAuth());
  else if (state.view === 'form') app.appendChild(renderForm());
  _ensureThemeBar();
  _applyTheme();
}

// ─── Auth (reused pattern from calendar) ───
function renderAuth() {
  let input;
  return h('div', { className: 'auth-screen' },
    h('h2', null, 'Password Required'),
    h('p', null, state.project ? state.project.Name + ' — Transmittal' : 'This project is protected'),
    input = h('input', { type: 'password', placeholder: 'Enter password' }),
    h('button', { className: 'btn btn-primary', onClick: async () => {
      try {
        await api('/api/projects/' + state.projectId + '/verify', {
          method: 'POST', body: JSON.stringify({ password: input.value }),
        });
        await loadTransmittal();
      } catch { alert('Invalid password'); }
    }}, 'Unlock'),
  );
}

// ─── Boot ───
async function loadTransmittal() {
  const tx = await api('/api/projects/' + state.projectId + '/transmittal');
  state.transmittal = tx;
  state.view = 'form';
  render();
}

(async function boot() {
  const parts = window.location.pathname.replace(/\/+$/, '').split('/').filter(Boolean);
  // Expect: [client, project, 'transmittal']
  if (parts.length >= 3 && parts[2] === 'transmittal') {
    state.pathClient = parts[0];
    state.pathProject = parts[1];
    try {
      const info = await api('/api/project-by-path/' + parts[0] + '/' + parts[1]);
      state.project = info.project;
      state.projectId = info.project.ID;
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
  // Production
  for (const k of ['start_date','pages_to_printer','pages_to_epub','print_run']) {
    total++; if (d.production && d.production[k]) filled++;
  }
  // Checklist — count items with explicit status, while preserving older saved data
  if (d.checklist) {
    total += d.checklist.length;
    filled += d.checklist.filter(c => !!getChecklistItemStatus(c)).length;
  }
  if (d.backmatter) {
    total += d.backmatter.length;
    filled += d.backmatter.filter(c => !!getChecklistItemStatus(c)).length;
  }
  // Design
  for (const k of ['trim','complexity']) {
    total++; if (d.design && d.design[k]) filled++;
  }
  // Editing
  total++; if (d.editing && d.editing.copyediting_level) filled++;
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
    // Header
    h('div', { className: 'page-header' },
      h('div', { className: 'page-header-top' },
        h('div', { className: 'page-header-left' },
          h('h1', { className: 'page-header-title' }, state.transmittal?.data?.book?.title || state.project?.Name || 'Transmittal'),
        ),
        h('div', { className: 'page-header-actions' },
          h('button', { className: 'btn btn-sm', onClick: () => {
            state.showVersions = !state.showVersions;
            if (state.showVersions) { state.versions = null; loadVersions(); }
            else render();
          }}, 'History'),
          h('button', { className: 'btn btn-sm', onClick: () => {
            state.showDuplicate = true; render();
          }}, 'Duplicate'),
          h('button', { className: 'btn btn-sm', onClick: () => window.print() }, 'Print'),
          h('button', { className: 'btn btn-sm', onClick: () => { state.showEmail = true; render(); }}, 'Email'),
          h('button', { className: 'btn btn-sm' + (state.transmittal.status === 'final' ? ' btn-primary' : ''),
            onClick: () => {
              const wasDraft = state.transmittal.status !== 'final';
              state.transmittal.status = wasDraft ? 'final' : 'draft';
              scheduleSave();
              if (wasDraft) {
                state.showEmail = true;
                state.emailResult = null;
              }
              render();
            }
          }, state.transmittal.status === 'final' ? 'Draft' : 'Mark Final'),
          themeBtn(),
        ),
      ),
      h('div', { className: 'page-header-sub' },
        h('button', { className: 'page-header-back', onClick: () => {
          window.location.href = '/' + state.pathClient + '/';
        }}, '← ' + (state.pathClient || 'Home').toUpperCase()),
        renderProjectSwitcher() || h('span', { style: 'font-size:13px;color:var(--text-secondary)' }, state.project.Name),
        h('a', { href: calendarUrl, style: 'font-size:0.8rem;color:var(--accent);text-decoration:none' }, 'Calendar'),
        h('span', { className: 'page-status page-status-' + state.transmittal.status },
          state.transmittal.status
        ),
        h('span', { style: 'font-size:0.78rem;color:var(--text-secondary)' }, 'Autosaves as you edit'),
        h('span', { id: 'tx-save-status', className: 'tx-save-status' }),
      ),
    ),
    // Email modal
    renderEmailModal(),
    // Version history panel (slides in from right)
    renderVersionPanel(),
    // Duplicate modal
    renderDuplicateModal(),
    // Progress
    h('div', { className: 'tx-progress' },
      h('div', { className: 'tx-progress-bar', style: 'width:' + pct + '%' }),
    ),
    // Two-column layout
    h('div', { className: 'tx-columns' + (isPreview ? ' tx-preview-mode' : '') },
      // LEFT COLUMN
      h('div', { className: 'tx-column' },
        renderBookSection(),
        renderProductionSection(),
        renderChecklistSection(),
        renderIllustrationsSection(),
        renderCoverSection(),
      ),
      // RIGHT COLUMN
      h('div', { className: 'tx-column' },
        renderEditingSection(),
        renderPermissionsSection(),
        renderPageIVSection(),
        renderDesignSection(),
        renderProofsSection(),
        renderFilesSection(),
        renderSubrightsSection(),
      ),
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
    h('div', { className: 'tx-row' },
      textField('Transmittal Date', 'book.transmittal_date', { type: 'date' }),
    ),
    h('div', { className: 'tx-row-3' },
      textField('ISBN (paper)', 'book.isbn_paper'),
      textField('ISBN (EPUB)', 'book.isbn_epub'),
      textField('ISBN (cloth)', 'book.isbn_cloth'),
    ),
  );
}

// ─── Section: Production ───
function renderProductionSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Production'),
    h('div', { className: 'tx-row' },
      textField('Transmittal Date', 'production.transmittal_date', { type: 'date' }),
      textField('Mechs Delivery', 'production.mechs_delivery', { type: 'date' }),
    ),
    h('div', { className: 'tx-row-3' },
      textField('Weeks in Prod.', 'production.weeks_in_production'),
      textField('Bound Book Date', 'production.bound_book_date', { type: 'date' }),
      textField('Print Run', 'production.print_run'),
    ),
  );
}

// ─── Section: Checklist ───
function renderChecklistSection() {
  const d = state.transmittal.data;
  const checklist = d.checklist || [];
  const backmatter = d.backmatter || [];
  const stats = d.checklist_stats || {};

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

  function checklistRow(item, i, collectionName, collection, options = {}) {
    const status = getChecklistItemStatus(item);
    const disabled = status !== 'later';
    return h('tr', null,
      h('td', { className: options.indent ? 'component-indent' : 'component-name' },
        options.label || item.component
      ),
      h('td', { style: 'width:190px' },
        h('select', {
          onChange: (e) => updateChecklistRow(collectionName, collection, i, e.target.value)
        },
          ...[
            ['', '— Select —'],
            ['included', 'In ms now'],
            ['later', 'Coming later'],
            ['not_in_book', 'Not included'],
          ].map(([value, label]) => {
            const opt = h('option', { value }, label);
            if (status === value) opt.selected = true;
            return opt;
          })
        )
      ),
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

  const checklistRows = checklist.map((item, i) =>
    checklistRow(item, i, 'checklist', checklist, { indent: item.indent })
  );

  const bmRows = backmatter.map((item, i) =>
    checklistRow(item, i, 'backmatter', backmatter, {
      label: item.component + (item.subtype ? ' (' + item.subtype + ')' : ''),
      indent: false,
    })
  );

  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Manuscript Checklist'),
    h('div', { className: 'tx-help' }, 'For each component, choose whether it is in the manuscript now, coming later, or not included in this book.'),
    h('table', { className: 'tx-checklist' },
      h('thead', null, h('tr', null,
        h('th', null, 'Component'),
        h('th', null, 'Status'),
        h('th', null, 'Expected date'),
      )),
      h('tbody', null,
        ...checklistRows,
        // Stats rows
        h('tr', { className: 'stats-row' },
          h('td', { colspan: '3' },
            h('div', { className: 'tx-row' },
              textField('Parts', 'checklist_stats.parts'),
              textField('Chapters', 'checklist_stats.chapters'),
            ),
            h('div', { className: 'tx-row-3' },
              textField('Words/Chars', 'checklist_stats.words_chars'),
              textField('MS pp', 'checklist_stats.ms_pp'),
              textField('Est. Book pp', 'checklist_stats.est_book_pp'),
            ),
          ),
        ),
        // Back matter header
        h('tr', null, h('td', { colspan: '3', style: 'padding-top:10px;font-weight:600;font-size:12px;color:var(--accent)' }, 'Back Matter')),
        ...bmRows,
      ),
    ),
  );
}

// ─── Section: Illustrations ───
function renderIllustrationsSection() {
  const types = [
    ['Figures', 'figures'], ['Tables', 'tables'], ['Photos', 'photos'], ['Other', 'other']
  ];
  return h('div', { className: 'tx-section tx-section-important' },
    h('div', { className: 'tx-section-header' }, 'Illustrations'),
    h('table', { className: 'tx-illus-table' },
      h('thead', null, h('tr', null,
        h('th', null, 'Type'), h('th', null, 'No.'), h('th', { style: 'text-align:center' }, 'Here'), h('th', null, 'To Come'),
      )),
      h('tbody', null, ...types.map(([label, key]) =>
        h('tr', null,
          h('td', null, label),
          h('td', null, h('input', { type: 'number', value: getField('illustrations.' + key + '_no') || '',
            onInput: (e) => setField('illustrations.' + key + '_no', parseInt(e.target.value) || 0) })),
          h('td', { style: 'text-align:center' },
            h('input', { type: 'checkbox', checked: getField('illustrations.' + key + '_here') ? 'checked' : undefined,
              onChange: (e) => setField('illustrations.' + key + '_here', e.target.checked) })),
          h('td', null, h('input', { type: 'text', value: getField('illustrations.' + key + '_to_come') || '',
            onInput: (e) => setField('illustrations.' + key + '_to_come', e.target.value),
            style: 'width:100%;padding:2px 4px;background:var(--bg);border:1px solid transparent;border-radius:3px;color:var(--text);font-size:12px' })),
        )
      )),
    ),
    textareaField('Art & Production Plan / Budget', 'illustrations.art_plan', {
      rows: 4,
      className: 'tx-field-important',
      helpText: 'Priority field: include art plan expectations, budget notes, and constraints.',
    }),
  );
}

// ─── Section: Permissions ───
function renderPermissionsSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Permissions & Consents'),
    h('div', { className: 'tx-field' },
      h('label', null, 'Permissions (reprint material)'),
      selectField('', 'permissions.reprint_status', [
        ['','— Select —'],['no_permissions','No permissions needed'],['permissions_needed','Permissions needed'],['permissions_pending','Permissions pending']
      ]),
    ),
    textField('Permissions to come by', 'permissions.reprint_when', { type: 'date' }),
    h('div', { className: 'tx-field' },
      h('label', null, 'Consents (original material)'),
      selectField('', 'permissions.consents_status', [
        ['','— Select —'],['no_consents','No consents needed'],['consents_needed','Consents needed'],['consents_pending','Consents pending']
      ]),
    ),
    textField('Consents to come by', 'permissions.consents_when', { type: 'date' }),
  );
}

// ─── Section: Page IV (Copyright) ───
function renderPageIVSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'PUB INFO & ©'),
    h('div', { className: 'tx-row' },
      textField('Copyright Year', 'page_iv.copyright_year'),
      textField('Held by', 'page_iv.held_by'),
    ),
    textField('Credit Line', 'page_iv.credit'),
    textField('Other Credit', 'page_iv.other_credit'),
    textField('Photo Credit', 'page_iv.photo_credit'),
  );
}

// ─── Section: Subrights ───
function renderSubrightsSection() {
  const items = [
    ['Copub with', 'copub'], ['Title page', 'title_page'], ['Page iv', 'page_iv'],
    ['Cover', 'cover'], ['Remove mktg pgs?', 'remove_mktg']
  ];
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Subrights'),
    ...items.map(([label, key]) => {
      const val = getField('subrights.' + key);
      const isNa = val === 'na';
      return h('div', { className: 'tx-inline' },
        h('label', { className: isNa ? 'tx-check na' : '', style: 'min-width:130px;cursor:pointer',
          onClick: () => { setField('subrights.' + key, isNa ? '' : 'na'); render(); }
        }, (isNa ? '✕ ' : '') + label),
        !isNa ? h('input', { type: 'text', value: val || '', placeholder: 'details...',
          onInput: (e) => setField('subrights.' + key, e.target.value),
          style: 'flex:1;padding:4px 6px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px'
        }) : h('span', { style: 'color:var(--text-secondary);font-size:12px;cursor:pointer',
          onClick: () => { setField('subrights.' + key, ''); render(); }
        }, 'n/a — click to enable'),
      );
    }),
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
    const next = [...styles, { name: '', type: 'paragraph', description: '' }];
    updateCustomStyles(next);
  }

  function removeCustomStyle(index) {
    const next = styles.filter((_, i) => i !== index);
    updateCustomStyles(next);
  }

  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Editing'),
    selectField('Developmental Edit Needed', 'editing.developmental_edit', [
      ['','— Select —'],
      ['none','No'],
      ['light','Light pass'],
      ['standard','Standard developmental edit'],
      ['heavy','Heavy / substantive'],
    ]),
    textareaField('Instructions for Developmental Editor', 'editing.developmental_instructions', {
      rows: 3,
      className: 'tx-field-important',
      placeholder: 'Any guidance for developmental edit focus, scope, or priorities...',
      helpText: 'Priority field: use this for anything the developmental editor must not miss.',
    }),
    selectField('Level of Copyediting', 'editing.copyediting_level', [
      ['','— Select —'],['light','Light'],['medium','Medium'],['heavy','Heavy']
    ]),
    textareaField('Instructions for Copyeditor', 'editing.instructions', {
      rows: 4,
      className: 'tx-field-important',
      helpText: 'Priority field: use this for anything the copyeditor must not miss.',
    }),
    textField('Special Characters', 'editing.special_characters'),
    textField('Mathematical Formulas', 'editing.math_formulas'),
    h('div', { className: 'tx-section-header', style: 'margin-top:16px' }, 'Custom Styles'),
    h('div', { className: 'tx-help' }, 'Add any project-specific Word styles needed for this manuscript. These will be copied into the book spec and used for Word template generation.'),
    ...styles.map((style, i) =>
      h('div', { className: 'tx-custom-style' },
        h('div', { className: 'tx-row-3' },
          textField('Style name', `custom_styles.${i}.name`),
          selectField('Type', `custom_styles.${i}.type`, [
            ['paragraph', 'Paragraph'],
            ['character', 'Character'],
          ]),
          textField('Purpose / description', `custom_styles.${i}.description`)
        ),
        h('button', { className: 'tx-reviewer-remove', type: 'button', onClick: () => removeCustomStyle(i) }, 'Remove')
      )
    ),
    h('button', { className: 'tx-add-btn', type: 'button', onClick: addCustomStyle }, '+ Add custom style'),
  );
}

// ─── Section: Book Design ───
function renderDesignSection() {
  const trimVal = getField('design.trim') || '';
  const standardTrims = ['5.5 x 8.5', '6 x 9', '8.5 x 11', 'dont_care', ''];

  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Book Design'),
    textareaField('Trim Guidance', 'design.trim_guidance', {
      rows: 3,
      className: 'tx-field-important',
      placeholder: 'e.g. coffee table, pocket book size, gift format',
      helpText: 'Priority field: use this for trim intent, flexibility, and format direction.',
    }),
    h('div', { className: 'tx-field' },
      h('label', null, 'Trim Size'),
      h('div', { className: 'tx-check-group tx-trim-options' },
        ...['5.5 x 8.5', '6 x 9', '8.5 x 11'].map(sz =>
          h('label', { className: 'tx-check' },
            h('input', { type: 'radio', name: 'trim', value: sz,
              checked: trimVal === sz ? 'checked' : undefined,
              onChange: () => { setField('design.trim', sz); render(); }
            }), sz.toUpperCase()
          )
        ),
        h('label', { className: 'tx-check' },
          h('input', { type: 'radio', name: 'trim', value: 'dont_care',
            checked: trimVal === 'dont_care' ? 'checked' : undefined,
            onChange: () => { setField('design.trim', 'dont_care'); render(); }
          }), `DON'T CARE`
        ),
        h('label', { className: 'tx-check' },
          h('input', { type: 'radio', name: 'trim', value: 'other',
            checked: !standardTrims.includes(trimVal) ? 'checked' : undefined,
            onChange: () => { setField('design.trim', '7 x 9'); render(); }
          }), 'OTHER:'
        ),
        !standardTrims.includes(trimVal)
          ? h('input', {
              type: 'text',
              className: 'tx-trim-other-input',
              value: trimVal,
              onInput: (e) => setField('design.trim', e.target.value)
            })
          : null,
      ),
    ),

    h('div', { className: 'tx-row-3' },
      textField('Est. Book pp', 'design.est_pages'),
      textField('PPI', 'design.ppi'),
      textField('Spine Width', 'design.spine_width'),
    ),
    h('div', { className: 'tx-row' },
      selectField('Text Complexity', 'design.complexity', [
        ['','— Select —'],['simple_jdbb','Simple (jdbb)'],['complex_jdbb','Complex (jdbb)']
      ]),
      textField('Outside Designer', 'design.outside_designer'),
    ),
    textField('Reuse Previous Book', 'design.reuse_previous'),
    textareaField('Additional Design Notes', 'design.freeform_notes', {
      rows: 3,
      placeholder: 'Any free-form direction for format, feel, or production constraints...',
    }),
  );
}

// ─── Section: Cover ───
function renderCoverSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Cover'),
    h('div', { className: 'tx-row' },
      selectField('Paper', 'cover.paper', [['','— Select —'],['paper','Paper'],['cloth','Cloth']]),
      textField('Colors', 'cover.colors', { placeholder: 'e.g. 4C' }),
    ),
    h('div', { className: 'tx-field' },
      h('label', null, 'JDBB'),
      h('div', { className: 'tx-check-group tx-cover-options' },
        checkField('FRONT', 'cover.jdbb_front'),
        checkField('SPINE', 'cover.jdbb_spine'),
        checkField('BACK', 'cover.jdbb_back'),
      ),
    ),
    h('div', { className: 'tx-field' },
      h('label', null, 'Publisher'),
      h('div', { className: 'tx-check-group tx-cover-options' },
        checkField('FRONT', 'cover.pub_front'),
        checkField('SPINE', 'cover.pub_spine'),
        checkField('BACK', 'cover.pub_back'),
      ),
    ),
    textField('Cover Credit', 'cover.credit'),
    textareaField('Production Plan / Budget', 'cover.production_plan_budget', {
      rows: 3,
      className: 'tx-field-important',
      helpText: 'Priority field: key production-plan and budget context for cover + print timing.',
    }),
  );
}

// ─── Section: Deliverables ───
function renderFilesSection() {
  const d = state.transmittal.data;
  const deliverables = (d.files && d.files.deliverables) || [];
  const deliverableOptions = [
    'Print-ready PDF (interior)',
    'EPUB file',
    'Cover files (print + digital)',
    'Typst source files',
    'Final manuscript (Word/DOCX)',
    'Fonts used (if licensable)',
  ];
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Deliverables'),
    h('div', { className: 'tx-field' },
      h('label', null, 'Client receives at project end'),
      h('div', { className: 'tx-check-group tx-delivery-archives' },
        ...deliverableOptions.map(opt => {
          const has = deliverables.includes(opt);
          return h('label', { className: 'tx-check' },
            h('input', { type: 'checkbox', checked: has ? 'checked' : undefined,
              onChange: (e) => {
                let newArr = [...deliverables];
                if (e.target.checked) newArr.push(opt);
                else newArr = newArr.filter(x => x !== opt);
                setField('files.deliverables', newArr);
              }
            }), opt
          );
        }),
      ),
    ),
    h('div', { className: 'tx-field' },
      h('label', null, 'Printer Delivery'),
      h('div', { className: 'tx-check-group tx-delivery-options' },
        ...['PDF/X', 'Other'].map(fmt =>
          h('label', { className: 'tx-check' },
            h('input', { type: 'radio', name: 'printer_format', value: fmt,
              checked: getField('files.printer_format') === fmt ? 'checked' : undefined,
              onChange: () => { setField('files.printer_format', fmt); render(); }
            }), fmt
          )
        ),
      ),
      getField('files.printer_format') === 'Other'
        ? h('input', { type: 'text', value: getField('files.printer_format_other') || '', placeholder: 'Specify format',
            className: 'tx-input', style: 'margin-top:6px',
            onInput: (e) => setField('files.printer_format_other', e.target.value)
          })
        : null,
    ),
  );
}

// ─── Section: Page Proofs ───
function renderProofsSection() {
  const d = state.transmittal.data;
  const reviewers = (d.proofs && d.proofs.reviewers) || [];
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Page Proofs'),
    h('div', { style: 'font-size:12px;color:var(--text-secondary);margin-bottom:8px' }, '1st pages to be reviewed by:'),
    ...reviewers.map((rev, i) =>
      h('div', { className: 'tx-reviewer' },
        h('input', { type: 'text', value: rev.name || '', placeholder: 'Name',
          onInput: (e) => { reviewers[i].name = e.target.value; setField('proofs.reviewers', reviewers); }
        }),
        h('input', { type: 'text', value: rev.contact || '', placeholder: 'Email',
          onInput: (e) => { reviewers[i].contact = e.target.value; setField('proofs.reviewers', reviewers); }
        }),
        h('button', { className: 'tx-reviewer-remove', onClick: () => {
          reviewers.splice(i, 1); setField('proofs.reviewers', reviewers); render();
        }}, '×'),
      )
    ),
    h('button', { className: 'tx-add-btn', onClick: () => {
      reviewers.push({ name: '', contact: '' }); setField('proofs.reviewers', reviewers); render();
    }}, '+ Add Reviewer'),
  );
}


// ─── Email Modal ───

const DEFAULT_RECIPIENTS = [
  { email: 'jdbb@agentmail.to', label: 'JDBB Archive', checked: true, editable: false },
  { email: 'j@djinna.com', label: 'Jenna', checked: true, editable: false },
];

let emailRecipients = null; // initialized on first open

function initEmailRecipients() {
  if (emailRecipients) return;
  emailRecipients = DEFAULT_RECIPIENTS.map(r => ({ ...r }));
  // Add a blank "custom" row
  emailRecipients.push({ email: '', label: 'Other', checked: false, editable: true });
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
        h('h2', null, 'Email Transmittal'),
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
              'Set AGENTMAIL_API_KEY and AGENTMAIL_INBOX_ID environment variables.'
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
              '2713 Sent to: ' + state.emailResult.sent_to.join(', ')
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
            ? h('button', { className: 'btn btn-primary', onClick: closeModal }, '2713 Done')
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
