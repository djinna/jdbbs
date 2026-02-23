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

// ─── Theme ───
function getTheme() { return localStorage.getItem('prodcal-theme') || 'dark'; }
function setTheme(t) { localStorage.setItem('prodcal-theme', t); document.documentElement.setAttribute('data-theme', t); }
function toggleTheme() { setTheme(getTheme() === 'dark' ? 'light' : 'dark'); render(); }
function themeBtn() { return h('button', { className: 'theme-toggle', onClick: toggleTheme, title: 'Toggle light/dark mode' }, getTheme() === 'dark' ? '☀️' : '🌙'); }
if (getTheme() === 'light') document.documentElement.setAttribute('data-theme', 'light');

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
    state.allProjects = await api('/api/projects');
    // Re-render to show project switcher once loaded
    if (state.view === 'form') render();
  } catch (e) {
    console.error('Failed to load projects:', e);
    state.allProjects = [];
  }
}

function switchProject(proj) {
  const url = '/' + proj.ClientSlug + '/' + proj.ProjectSlug + '/transmittal/';
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
        window.location.href = '/' + target.ClientSlug + '/' + target.ProjectSlug + '/transmittal/';
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
}

// ─── Auth (reused pattern from calendar) ───
function renderAuth() {
  let input;
  return h('div', { className: 'auth-screen' },
    h('h2', null, '🔒 Password Required'),
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
  const val = getField(path) || '';
  const inp = h('input', {
    type: opts.type || 'text',
    value: val,
    placeholder: opts.placeholder || '',
    onInput: (e) => setField(path, opts.type === 'number' ? (parseFloat(e.target.value) || 0) : e.target.value),
  });
  return h('div', { className: 'tx-field' }, h('label', null, label), inp);
}

function textareaField(label, path, opts = {}) {
  const val = getField(path) || '';
  const ta = h('textarea', {
    rows: opts.rows || 3,
    placeholder: opts.placeholder || '',
    onInput: (e) => setField(path, e.target.value),
  }, val);
  return h('div', { className: 'tx-field' }, h('label', null, label), ta);
}

function selectField(label, path, options) {
  const val = getField(path) || '';
  const sel = h('select', { onChange: (e) => setField(path, e.target.value) },
    ...options.map(([v, l]) => {
      const opt = h('option', { value: v }, l);
      if (val === v) opt.selected = true;
      return opt;
    })
  );
  return h('div', { className: 'tx-field' }, h('label', null, label), sel);
}

function checkField(label, path) {
  const val = !!getField(path);
  return h('label', { className: 'tx-check' },
    h('input', { type: 'checkbox', checked: val ? 'checked' : undefined, onChange: (e) => setField(path, e.target.checked) }),
    label
  );
}

// ─── Completion calc ───
function calcCompletion() {
  const d = state.transmittal.data;
  let filled = 0, total = 0;
  // Book fields
  for (const k of ['author','title','publisher','editor','transmittal_date']) {
    total++; if (d.book && d.book[k]) filled++;
  }
  // Production
  for (const k of ['transmittal_date','mechs_delivery','print_run']) {
    total++; if (d.production && d.production[k]) filled++;
  }
  // Checklist — count items with here_now or to_come_when
  if (d.checklist) {
    total += d.checklist.length;
    filled += d.checklist.filter(c => c.here_now || c.to_come_when).length;
  }
  if (d.backmatter) {
    total += d.backmatter.length;
    filled += d.backmatter.filter(c => c.here_now || c.to_come_when).length;
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

// ─── Project Switcher dropdown ───
function renderProjectSwitcher() {
  if (!state.allProjects.length) return null;
  return h('select', {
    className: 'tx-project-switcher',
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
      ? h('p', { style: 'padding:12px;color:var(--text2)' }, 'Loading...')
      : versions.length === 0
        ? h('p', { style: 'padding:12px;color:var(--text2)' }, 'No versions yet. Versions are saved automatically as you edit (up to one every 5 minutes).')
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
      h('p', { style: 'color:var(--text2);font-size:13px;margin-bottom:12px' },
        'Copy this transmittal to another project. Author, publisher, design, and other house settings are kept. Book-specific fields (title, dates, checklist) are cleared.'
      ),
      others.length === 0
        ? h('p', { style: 'color:var(--text2)' }, 'No other projects available. Create a new project from the calendar first.')
        : h('div', { className: 'tx-duplicate-list' },
            ...others.map(p =>
              h('button', { className: 'btn btn-sm tx-duplicate-item', onClick: () => {
                state.showDuplicate = false;
                render();
                duplicateToProject(p.ID);
              }},
                h('span', null, p.Name),
                h('span', { style: 'color:var(--text2);font-size:11px' }, p.ClientSlug + '/' + p.ProjectSlug),
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
      h('span', null, '👁 Previewing version from ' + fmtDate(state.transmittal._previewDate)),
      h('button', { className: 'btn btn-sm', onClick: () => restoreVersion(state.transmittal._preview) }, 'Restore this version'),
      h('button', { className: 'btn btn-sm', onClick: exitPreview }, 'Exit preview'),
    ) : null,
    // Header
    h('div', { className: 'tx-header' },
      h('div', null,
        h('h1', null, '📋 Manuscript Transmittal'),
        h('div', { className: 'tx-subtitle' },
          renderProjectSwitcher() || state.project.Name,
          ' · ',
          h('span', { className: 'tx-status tx-status-' + state.transmittal.status },
            state.transmittal.status
          ),
        ),
      ),
      h('div', { className: 'tx-header-actions' },
        h('a', { className: 'btn btn-sm', href: calendarUrl }, '📅 Calendar'),
        h('button', { className: 'btn btn-sm', onClick: () => {
          state.showVersions = !state.showVersions;
          if (state.showVersions) { state.versions = null; loadVersions(); }
          else render();
        }}, '🕒 History'),
        h('button', { className: 'btn btn-sm', onClick: () => {
          state.showDuplicate = true; render();
        }}, '⧉ Duplicate'),
        h('button', { className: 'btn btn-sm', onClick: () => window.print() }, '🖨 Print'),
        h('button', { className: 'btn btn-sm' + (state.transmittal.status === 'final' ? ' btn-primary' : ''),
          onClick: () => {
            state.transmittal.status = state.transmittal.status === 'final' ? 'draft' : 'final';
            scheduleSave(); render();
          }
        }, state.transmittal.status === 'final' ? '↩ Back to Draft' : '✓ Mark Final'),
        themeBtn(),
        h('span', { id: 'tx-save-status', className: 'tx-save-status' }),
      ),
    ),
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
      ),
      // RIGHT COLUMN
      h('div', { className: 'tx-column' },
        renderPermissionsSection(),
        renderPageIVSection(),
        renderSubrightsSection(),
        renderEditingSection(),
        renderDesignSection(),
        renderCoverSection(),
        renderFilesSection(),
        renderProofsSection(),
        renderOtherSection(),
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
    h('div', { className: 'tx-row' },
      textField('ISBN (paper)', 'book.isbn_paper'),
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

  const checklistRows = checklist.map((item, i) =>
    h('tr', null,
      h('td', { className: item.indent ? 'component-indent' : 'component-name' }, item.component),
      h('td', { style: 'text-align:center;width:60px' },
        h('input', { type: 'checkbox', checked: item.here_now ? 'checked' : undefined,
          onChange: (e) => { checklist[i].here_now = e.target.checked; setField('checklist', checklist); }
        }),
      ),
      h('td', { style: 'width:100px' },
        h('input', { type: 'text', value: item.to_come_when || '', placeholder: 'date',
          onInput: (e) => { checklist[i].to_come_when = e.target.value; setField('checklist', checklist); }
        }),
      ),
    )
  );

  const bmRows = backmatter.map((item, i) =>
    h('tr', null,
      h('td', { className: 'component-name' },
        item.component + (item.subtype ? ' (' + item.subtype + ')' : ''),
      ),
      h('td', { style: 'text-align:center;width:60px' },
        h('input', { type: 'checkbox', checked: item.here_now ? 'checked' : undefined,
          onChange: (e) => { backmatter[i].here_now = e.target.checked; setField('backmatter', backmatter); }
        }),
      ),
      h('td', { style: 'width:100px' },
        h('input', { type: 'text', value: item.to_come_when || '', placeholder: 'date',
          onInput: (e) => { backmatter[i].to_come_when = e.target.value; setField('backmatter', backmatter); }
        }),
      ),
    )
  );

  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Manuscript Checklist'),
    h('table', { className: 'tx-checklist' },
      h('thead', null, h('tr', null,
        h('th', null, 'Component'),
        h('th', { style: 'text-align:center' }, 'Here'),
        h('th', null, 'To Come'),
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
  return h('div', { className: 'tx-section' },
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
    textareaField('Art & Production Plan / Budget', 'illustrations.art_plan', { rows: 4 }),
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
    h('div', { className: 'tx-section-header' }, 'Page IV (Copyright)'),
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
        }) : h('span', { style: 'color:var(--text2);font-size:12px;cursor:pointer',
          onClick: () => { setField('subrights.' + key, ''); render(); }
        }, 'n/a — click to enable'),
      );
    }),
  );
}

// ─── Section: Editing ───
function renderEditingSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Editing'),
    selectField('Level of Copyediting', 'editing.copyediting_level', [
      ['','— Select —'],['light','Light'],['medium','Medium'],['heavy','Heavy']
    ]),
    textField('Special Characters', 'editing.special_characters'),
    textField('Mathematical Formulas', 'editing.math_formulas'),
    textareaField('Other Instructions for Copyeditor', 'editing.instructions'),
  );
}

// ─── Section: Book Design ───
function renderDesignSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Book Design'),
    h('div', { className: 'tx-field' },
      h('label', null, 'Trim Size'),
      h('div', { className: 'tx-check-group' },
        ...['5.5 x 8.5', '6 x 9', '8.5 x 11'].map(sz =>
          h('label', { className: 'tx-check' },
            h('input', { type: 'radio', name: 'trim', value: sz,
              checked: getField('design.trim') === sz ? 'checked' : undefined,
              onChange: () => { setField('design.trim', sz); render(); }
            }), sz
          )
        ),
        h('label', { className: 'tx-check' },
          h('input', { type: 'radio', name: 'trim', value: 'other',
            checked: !['5.5 x 8.5','6 x 9','8.5 x 11',''].includes(getField('design.trim')) ? 'checked' : undefined,
            onChange: () => { setField('design.trim', '7x9'); render(); }
          }), 'other:'
        ),
        !['5.5 x 8.5','6 x 9','8.5 x 11',''].includes(getField('design.trim'))
          ? h('input', { type: 'text', value: getField('design.trim'), style: 'width:80px;padding:2px 4px;font-size:12px;background:var(--bg);border:1px solid var(--border);border-radius:3px;color:var(--text)',
              onInput: (e) => setField('design.trim', e.target.value) })
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
      h('div', { className: 'tx-check-group' },
        checkField('Front', 'cover.jdbb_front'),
        checkField('Spine', 'cover.jdbb_spine'),
        checkField('Back', 'cover.jdbb_back'),
      ),
    ),
    h('div', { className: 'tx-field' },
      h('label', null, 'Publisher'),
      h('div', { className: 'tx-check-group' },
        checkField('Front', 'cover.pub_front'),
        checkField('Spine', 'cover.pub_spine'),
        checkField('Back', 'cover.pub_back'),
      ),
    ),
    textField('Cover Credit', 'cover.credit'),
  );
}

// ─── Section: Files ───
function renderFilesSection() {
  const d = state.transmittal.data;
  const archives = (d.files && d.files.archives) || [];
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Files & Delivery'),
    h('div', { className: 'tx-field' },
      h('label', null, 'Printer Delivery Format'),
      h('div', { className: 'tx-check-group' },
        ...['native', 'postscript', 'PDF'].map(fmt =>
          h('label', { className: 'tx-check' },
            h('input', { type: 'radio', name: 'printer_format', value: fmt,
              checked: getField('files.printer_format') === fmt ? 'checked' : undefined,
              onChange: () => { setField('files.printer_format', fmt); render(); }
            }), fmt
          )
        ),
      ),
    ),
    h('div', { className: 'tx-field' },
      h('label', null, 'Customer Archives'),
      h('div', { className: 'tx-check-group', style: 'flex-direction:column;gap:4px' },
        ...['native', 'export final text', 'PDF: for web, by chapter', 'PDF: for print, by chapter'].map(opt => {
          const has = archives.includes(opt);
          return h('label', { className: 'tx-check' },
            h('input', { type: 'checkbox', checked: has ? 'checked' : undefined,
              onChange: (e) => {
                let newArr = [...archives];
                if (e.target.checked) newArr.push(opt);
                else newArr = newArr.filter(x => x !== opt);
                setField('files.archives', newArr);
              }
            }), opt
          );
        }),
      ),
    ),
  );
}

// ─── Section: Page Proofs ───
function renderProofsSection() {
  const d = state.transmittal.data;
  const reviewers = (d.proofs && d.proofs.reviewers) || [];
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Page Proofs'),
    h('div', { style: 'font-size:12px;color:var(--text2);margin-bottom:8px' }, '1st pages to be reviewed by:'),
    ...reviewers.map((rev, i) =>
      h('div', { className: 'tx-reviewer' },
        h('input', { type: 'text', value: rev.name || '', placeholder: 'Name',
          onInput: (e) => { reviewers[i].name = e.target.value; setField('proofs.reviewers', reviewers); }
        }),
        h('input', { type: 'text', value: rev.contact || '', placeholder: 'Address / Tel',
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

// ─── Section: Other Instructions ───
function renderOtherSection() {
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Other Instructions'),
    textareaField('', 'other_instructions', { rows: 4, placeholder: 'Any other instructions...' }),
  );
}

