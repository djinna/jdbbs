// Production Calendar SPA
const $ = (s, p) => (p || document).querySelector(s);
const $$ = (s, p) => [...(p || document).querySelectorAll(s)];
const h = (tag, attrs, ...kids) => {
  const el = document.createElement(tag);
  if (attrs) Object.entries(attrs).forEach(([k, v]) => {
    if (k.startsWith('on')) el.addEventListener(k.slice(2).toLowerCase(), v);
    else if (k === 'className') el.className = v;
    else if (k === 'htmlFor') el.htmlFor = v;
    else if (k === 'checked' || k === 'disabled' || k === 'selected') { if (v) el[k] = true; }
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

function absoluteURL(path) {
  return new URL(path, window.location.origin + '/').toString();
}

const fmt = {
  date(s) { if (!s) return '—'; const d = new Date(s + 'T00:00:00'); return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }); },
  money(n) { return n ? '$' + n.toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 }) : '—'; },
};

let state = { view: 'projects', projectId: null, project: null, tasks: [], tab: 'gantt', editingTask: null, pathClient: null, pathProject: null, showSnapshotEmail: false, snapshotSending: false, snapshotResult: null, emailConfigured: null, siblingProjects: [], fileLog: [], journal: [], showFileLogModal: false, showJournalModal: false, showActivityEmail: false, activitySending: false, activityResult: null };

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

// Inject theme bar into page (called after render)
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
// Legacy compat
function themeBtn() { return h('span'); }
function getTheme() { return _themeState.dark ? 'dark' : 'light'; }

function render() {
  const app = $('#app');
  app.innerHTML = '';
  if (state.view === 'projects') app.appendChild(renderProjectList());
  else if (state.view === 'auth') app.appendChild(renderAuth());
  else if (state.view === 'project') app.appendChild(renderProject());
  _ensureThemeBar();
  _applyTheme();
}

// ─── Project List ───
async function loadProjects() {
  state.projects = await api('/api/projects');
  render();
}

function renderProjectList() {
  return h('div', null,
    h('div', { className: 'header' },
      h('h1', null, 'Production Calendar'),
      h('div', { className: 'header-actions' },
        h('button', { className: 'btn btn-primary', onClick: showNewProject }, '+ New Project'),
      )
    ),
    state.projects && state.projects.length
      ? h('div', { className: 'project-grid' },
          ...state.projects.map(p => {
            const path = p.ClientSlug && p.ProjectSlug ? '/' + p.ClientSlug + '/' + p.ProjectSlug + '/' : null;
            return h('div', { className: 'project-card', onClick: () => {
              if (path) window.location.href = path;
              else openProject(p.ID);
            } },
              h('h3', null, p.Name),
              path ? h('div', { className: 'meta', style: 'color:var(--accent2);font-family:monospace' }, path) : null,
              h('div', { className: 'meta' }, 'Start: ' + (p.StartDate || 'Not set')),
              h('div', { className: 'meta' }, 'Updated: ' + new Date(p.UpdatedAt).toLocaleDateString()),
            );
          })
        )
      : h('div', { className: 'empty-state' },
          h('h3', null, 'No projects yet'),
          h('p', null, 'Create your first production schedule'),
        )
  );
}

function showNewProject() {
  let nameInput, clientInput, projInput, dateInput;
  const el = h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target === el) el.remove(); } },
    h('div', { className: 'modal' },
      h('h2', null, '+ New Project'),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Project Title'),
          nameInput = h('input', { type: 'text', placeholder: 'e.g. Art of Gig' }),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Client Slug (URL)'),
          clientInput = h('input', { type: 'text', placeholder: 'e.g. vgr' }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Project Slug (URL)'),
          projInput = h('input', { type: 'text', placeholder: 'e.g. aog' }),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Start Date'),
          dateInput = h('input', { type: 'date', value: new Date().toISOString().slice(0, 10) }),
        ),
      ),
      h('div', { style: 'background:var(--surface2);border-radius:var(--radius);padding:12px;margin:8px 0;font-size:13px;color:var(--text-secondary)' },
        'URL will be: ', h('strong', { style: 'color:var(--accent2);font-family:monospace' }, '/‹client›/‹project›/'),
      ),
      h('div', { className: 'modal-actions' },
        h('button', { className: 'btn', onClick: () => el.remove() }, 'Cancel'),
        h('button', { className: 'btn btn-primary', onClick: async () => {
          const name = nameInput.value.trim();
          const cs = clientInput.value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
          const ps = projInput.value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
          if (!name || !cs || !ps) { alert('All fields required'); return; }
          try {
            const p = await api('/api/projects', {
              method: 'POST',
              body: JSON.stringify({ name, start_date: dateInput.value, client_slug: cs, project_slug: ps }),
            });
            el.remove();
            openProject(p.ID);
          } catch (e) { alert('Error: ' + e.message); }
        }}, 'Create'),
      ),
    )
  );
  document.body.appendChild(el);
  nameInput.focus();
}

async function openProject(id) {
  state.projectId = id;
  try {
    const info = await api('/api/projects/' + id);
    state.project = info.project;
    if (info.has_auth && !info.authenticated) {
      state.view = 'auth';
      render();
      return;
    }
    state.tasks = await api('/api/projects/' + id + '/tasks');
    state.view = 'project';
    render();
  } catch (e) {
    if (e.message === 'unauthorized') { state.view = 'auth'; render(); }
    else alert(e.message);
  }
}

// ─── Auth ───
function renderAuth() {
  let input;
  return h('div', { className: 'auth-screen' },
    h('h2', null, 'Password Required'),
    h('p', null, state.project ? state.project.Name : 'This project is protected'),
    input = h('input', { type: 'password', placeholder: 'Enter password' }),
    h('button', { className: 'btn btn-primary', onClick: async () => {
      try {
        await api('/api/projects/' + state.projectId + '/verify', {
          method: 'POST', body: JSON.stringify({ password: input.value }),
        });
        openProject(state.projectId);
      } catch { alert('Invalid password'); }
    }}, 'Unlock'),
    h('br'), h('br'),
    h('button', { className: 'btn', onClick: () => { state.view = 'projects'; render(); } }, '← Back'),
  );
}

// ─── Project View ───
function renderProject() {
  const t = state.tasks;
  const origTotal = t.reduce((s, x) => s + (x.OrigBudget || 0), 0);
  const currTotal = t.reduce((s, x) => s + (x.CurrBudget || 0), 0);
  const actualTotal = t.reduce((s, x) => s + (x.ActualBudget || 0), 0);
  const done = t.filter(x => x.Status === 'done').length;

  const txUrl = state.project.ClientSlug && state.project.ProjectSlug
    ? absoluteURL('/' + state.project.ClientSlug + '/' + state.project.ProjectSlug + '/transmittal/') : null;
  const clientUrl = state.project.ClientSlug ? absoluteURL('/' + state.project.ClientSlug + '/') : absoluteURL('/');

  return h('div', null,
    h('div', { className: 'page-header' },
      h('div', { className: 'page-header-top' },
        h('div', { className: 'page-header-left' },
          h('h1', { className: 'page-header-title' }, state.project.Name),
        ),
        h('div', { className: 'page-header-actions' },
          h('button', { className: 'btn btn-sm', onClick: showAddTask }, '+ Task'),
          h('button', { className: 'btn btn-sm btn-primary', onClick: showDuplicate }, 'Make New'),
          h('button', { className: 'btn btn-sm', onClick: () => { state.showSnapshotEmail = true; state.snapshotResult = null; render(); } }, 'Email'),
          h('button', { className: 'btn btn-sm', onClick: showSettings }, 'Settings'),
        ),
      ),
      h('div', { className: 'page-header-sub' },
        h('button', { className: 'page-header-back', onClick: () => { window.location.href = clientUrl; } }, '← ' + (state.project.ClientSlug || 'Projects').toUpperCase()),
        renderProjectSwitcher() || h('span', { style: 'font-size:13px;color:var(--text-secondary)' }, state.project.Name),
        txUrl ? h('a', { href: txUrl, style: 'font-size:0.8rem;color:var(--accent);text-decoration:none' }, 'Transmittal') : null,
        h('span', { className: 'page-status' + (done === t.length && t.length > 0 ? ' page-status-final' : ' page-status-draft') },
          done + '/' + t.length + ' done'
        ),
      ),
    ),
    h('div', { className: 'budget-summary' },
      h('div', { className: 'budget-card' },
        h('div', { className: 'label' }, 'Tasks'),
        h('div', { className: 'value blue' }, done + '/' + t.length),
      ),
      h('div', { className: 'budget-card' },
        h('div', { className: 'label' }, 'Orig Budget'),
        h('div', { className: 'value' }, fmt.money(origTotal)),
      ),
      h('div', { className: 'budget-card' },
        h('div', { className: 'label' }, 'Curr Budget'),
        h('div', { className: 'value yellow' }, fmt.money(currTotal)),
      ),
      h('div', { className: 'budget-card' },
        h('div', { className: 'label' }, 'Actual Spent'),
        h('div', { className: 'value green' }, fmt.money(actualTotal)),
      ),
    ),
    h('div', { className: 'tabs' },
      h('button', { className: 'tab' + (state.tab === 'gantt' ? ' active' : ''), onClick: () => { state.tab = 'gantt'; render(); } }, 'Timeline'),
      h('button', { className: 'tab' + (state.tab === 'table' ? ' active' : ''), onClick: () => { state.tab = 'table'; render(); } }, 'Table'),
      h('button', { className: 'tab' + (state.tab === 'budget' ? ' active' : ''), onClick: () => { state.tab = 'budget'; render(); } }, 'Budget'),
      h('button', { className: 'tab' + (state.tab === 'files' ? ' active' : ''), onClick: () => { state.tab = 'files'; loadFileLog(); } }, 'Files'),
      h('button', { className: 'tab' + (state.tab === 'journal' ? ' active' : ''), onClick: () => { state.tab = 'journal'; loadJournal(); } }, 'Journal'),
    ),
    state.tab === 'gantt' ? renderGantt() : state.tab === 'table' ? renderTable() : state.tab === 'budget' ? renderBudget() : state.tab === 'files' ? renderFileLog() : state.tab === 'journal' ? renderJournal() : null,
    state.editingTask ? renderEditModal() : null,
    state.showSnapshotEmail ? renderSnapshotEmailModal() : null,
    state.showFileLogModal ? renderFileLogModal() : null,
    state.showJournalModal ? renderJournalModal() : null,
    state.showActivityEmail ? renderActivityEmailModal() : null,
  );
}

// ─── Gantt ───
function renderGantt() {
  const tasks = state.tasks;
  if (!tasks.length) return h('div', { className: 'empty-state' }, h('p', null, 'No tasks yet'));

  // Find date range
  const allDates = tasks.flatMap(t => [t.OrigDue, t.CurrDue].filter(Boolean)).map(d => new Date(d + 'T00:00:00'));
  if (!allDates.length) return h('div', { className: 'empty-state' }, h('p', null, 'No dates set'));
  const minDate = new Date(Math.min(...allDates));
  const maxDate = new Date(Math.max(...allDates));
  minDate.setDate(minDate.getDate() - 7);
  maxDate.setDate(maxDate.getDate() + 14);
  const totalDays = (maxDate - minDate) / 86400000;
  const pxPerDay = 5;
  const chartWidth = totalDays * pxPerDay;

  const toX = (dateStr) => {
    if (!dateStr) return null;
    const d = new Date(dateStr + 'T00:00:00');
    return ((d - minDate) / 86400000) * pxPerDay;
  };

  // Week headers
  const weeks = [];
  const ws = new Date(minDate);
  ws.setDate(ws.getDate() - ws.getDay());
  while (ws < maxDate) {
    weeks.push(new Date(ws));
    ws.setDate(ws.getDate() + 7);
  }

  // Build month labels positioned at month boundaries
  const months = [];
  const mStart = new Date(minDate);
  mStart.setDate(1);
  if (mStart < minDate) mStart.setMonth(mStart.getMonth() + 1);
  while (mStart < maxDate) {
    months.push(new Date(mStart));
    mStart.setMonth(mStart.getMonth() + 1);
  }

  const weekHeaders = h('div', { className: 'gantt-header-weeks', style: `width:${chartWidth}px;position:relative;height:18px` },
    // Month labels
    ...months.map(m => {
      const x = ((m - minDate) / 86400000) * pxPerDay;
      return h('div', { className: 'gantt-month-label', style: `left:${x + 2}px` },
        m.toLocaleDateString('en-US', { month: 'short' })
      );
    }),
    // Thin week ticks
    ...weeks.map(w => {
      const x = ((w - minDate) / 86400000) * pxPerDay;
      const isMonthStart = w.getDate() <= 7;
      return h('div', { className: 'gantt-header-tick' + (isMonthStart ? ' month-tick' : ''), style: `left:${x}px` });
    })
  );

  const rows = tasks.map(t => {
    const isMilestone = t.IsMilestone === 1;
    const origX = toX(t.OrigDue);
    const currX = toX(t.CurrDue);
    const origStartX = t.OrigWeeks > 0 ? origX - t.OrigWeeks * 7 * pxPerDay : origX;
    const currStartX = t.CurrWeeks > 0 ? currX - t.CurrWeeks * 7 * pxPerDay : currX;

    const bars = [];
    if (origX !== null) {
      const w = Math.max(4, (origX - origStartX));
      bars.push(h('div', { className: 'gantt-bar orig' + (isMilestone ? ' milestone-bar' : ''), style: `left:${origStartX}px;width:${w}px` }));
    }
    if (currX !== null) {
      const cls = t.Status === 'done' ? 'done' : 'curr';
      const w = Math.max(4, (currX - currStartX));
      bars.push(h('div', { className: 'gantt-bar ' + cls + (isMilestone ? ' milestone-bar' : ''), style: `left:${currStartX}px;width:${w}px` }));
    }

    const weekLines = weeks.map(w => {
      const x = ((w - minDate) / 86400000) * pxPerDay;
      return h('div', { className: 'gantt-week-line', style: `left:${x}px` });
    });

    const statusCls = 'status status-' + (t.Status || 'pending');
    const statusLabel = t.Status === 'done' ? '✓ Done' : t.Status === 'in_progress' ? '● Active' : '○ Pending';

    return h('tr', { className: isMilestone ? 'milestone' : '', onClick: () => editTask(t) },
      h('td', null, h('span', { className: 'badge badge-' + t.Assignee }, t.Assignee)),
      h('td', null, t.Title),
      h('td', null, h('span', { className: statusCls + ' clickable', onClick: (e) => { e.stopPropagation(); cycleStatus(t); }, title: 'Click to cycle status' }, statusLabel)),
      h('td', { className: 'date' }, fmt.date(t.CurrDue)),
      h('td', { className: 'gantt-bar-cell' },
        h('div', { className: 'gantt-bar-bg', style: `width:${chartWidth}px` }, ...weekLines, ...bars)
      ),
    );
  });

  return h('div', { className: 'gantt-container' },
    h('table', { className: 'gantt-table' },
      h('thead', null, h('tr', null,
        h('th', { style: 'width:50px' }, 'Who'),
        h('th', { style: 'width:200px' }, 'Task'),
        h('th', { style: 'width:80px' }, 'Status'),
        h('th', { style: 'width:80px' }, 'Due'),
        h('th', null, weekHeaders),
      )),
      h('tbody', null, ...rows),
    )
  );
}

// ─── Table ───
function renderTable() {
  const tasks = state.tasks;
  return h('div', { className: 'table-container' },
    h('table', { className: 'data-table' },
      h('thead', null, h('tr', null,
        h('th', null, 'Who'),
        h('th', null, 'Task'),
        h('th', null, 'Status'),
        h('th', null, 'Orig Wks'),
        h('th', null, 'Curr Wks'),
        h('th', null, 'Orig Due'),
        h('th', null, 'Curr Due'),
        h('th', null, 'Actual'),
        h('th', null, ''),
      )),
      h('tbody', null, ...tasks.map(t =>
        h('tr', { className: t.IsMilestone === 1 ? 'milestone' : '' },
          h('td', null, h('span', { className: 'badge badge-' + t.Assignee }, t.Assignee)),
          h('td', null, t.Title),
          h('td', null, h('span', { className: 'status status-' + (t.Status || 'pending') + ' clickable', onClick: (e) => { e.stopPropagation(); cycleStatus(t); }, title: 'Click to cycle status' },
            t.Status === 'done' ? '✓' : t.Status === 'in_progress' ? '●' : '○'
          )),
          h('td', { className: 'date' }, String(t.OrigWeeks || '—')),
          h('td', { className: 'date' }, String(t.CurrWeeks || '—')),
          h('td', { className: 'date' }, fmt.date(t.OrigDue)),
          h('td', { className: 'date' }, fmt.date(t.CurrDue)),
          h('td', { className: 'date' }, fmt.date(t.ActualDone)),
          h('td', null, h('button', { className: 'btn btn-sm', onClick: (e) => { e.stopPropagation(); editTask(t); } }, '✏')),
        )
      )),
    )
  );
}

// ─── Budget ───
function renderBudget() {
  const tasks = state.tasks.filter(t => t.OrigBudget || t.CurrBudget || t.ActualBudget);
  return h('div', { className: 'table-container' },
    h('table', { className: 'data-table' },
      h('thead', null, h('tr', null,
        h('th', null, 'Task'),
        h('th', null, 'Notes'),
        h('th', { style: 'text-align:right' }, 'Hours'),
        h('th', { style: 'text-align:right' }, 'Rate'),
        h('th', { style: 'text-align:right' }, 'Orig'),
        h('th', { style: 'text-align:right' }, 'Current'),
        h('th', { style: 'text-align:right' }, 'Actual'),
      )),
      h('tbody', null,
        ...tasks.map(t =>
          h('tr', { onClick: () => editTask(t) },
            h('td', null, t.Title),
            h('td', { style: 'color:var(--text-secondary)' }, t.BudgetNotes),
            h('td', { className: 'money' }, t.Hours ? t.Hours.toFixed(1) : '—'),
            h('td', { className: 'money' }, t.Rate ? fmt.money(t.Rate) : '—'),
            h('td', { className: 'money' }, fmt.money(t.OrigBudget)),
            h('td', { className: 'money' }, fmt.money(t.CurrBudget)),
            h('td', { className: 'money' }, fmt.money(t.ActualBudget)),
          )
        ),
        h('tr', { style: 'font-weight:700;border-top:2px solid var(--border)' },
          h('td', { colspan: '4' }, 'Total'),
          h('td', { className: 'money' }, fmt.money(tasks.reduce((s, t) => s + t.OrigBudget, 0))),
          h('td', { className: 'money' }, fmt.money(tasks.reduce((s, t) => s + t.CurrBudget, 0))),
          h('td', { className: 'money' }, fmt.money(tasks.reduce((s, t) => s + t.ActualBudget, 0))),
        )
      ),
    )
  );
}

// ─── Edit Task Modal ───
function editTask(t) { state.editingTask = { ...t }; render(); }

function renderEditModal() {
  const t = state.editingTask;
  const field = (label, key, type = 'text') => {
    return h('div', { className: 'form-group' },
      h('label', null, label),
      h('input', { type, value: t[key] ?? '', onInput: (e) => { t[key] = type === 'number' ? parseFloat(e.target.value) || 0 : e.target.value; } }),
    );
  };
  const selectField = (label, key, options) => {
    const sel = h('select', { onChange: (e) => { t[key] = e.target.value; } },
      ...options.map(([v, l]) => {
        const opt = h('option', { value: v }, l);
        if (t[key] === v) opt.selected = true;
        return opt;
      })
    );
    return h('div', { className: 'form-group' }, h('label', null, label), sel);
  };

  return h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target.classList.contains('modal-backdrop')) { state.editingTask = null; render(); } } },
    h('div', { className: 'modal' },
      h('h2', null, t.ID ? 'Edit Task' : 'New Task'),
      h('div', { className: 'form-row' },
        field('Assignee', 'Assignee'),
        field('Title', 'Title'),
      ),
      h('div', { className: 'form-row' },
        selectField('Status', 'Status', [['pending', 'Pending'], ['in_progress', 'In Progress'], ['done', 'Done']]),
        selectField('Milestone', 'IsMilestone', [['0', 'Task'], ['1', 'Milestone/Sub-task']]),
      ),
      h('div', { className: 'form-row' },
        field('Orig Weeks', 'OrigWeeks', 'number'),
        field('Curr Weeks', 'CurrWeeks', 'number'),
        field('Sort Order', 'SortOrder', 'number'),
      ),
      h('div', { className: 'form-row' },
        field('Orig Due', 'OrigDue', 'date'),
        field('Curr Due', 'CurrDue', 'date'),
        field('Actual Done', 'ActualDone', 'date'),
      ),
      h('div', { className: 'form-row' },
        field('Hours', 'Hours', 'number'),
        field('Rate ($)', 'Rate', 'number'),
        field('Budget Notes', 'BudgetNotes'),
      ),
      h('div', { className: 'form-row' },
        field('Orig Budget', 'OrigBudget', 'number'),
        field('Curr Budget', 'CurrBudget', 'number'),
        field('Actual Budget', 'ActualBudget', 'number'),
      ),
      h('div', { className: 'modal-actions' },
        t.ID ? h('button', { className: 'btn btn-danger btn-sm', onClick: deleteCurrentTask }, 'Delete') : null,
        h('div', { style: 'flex:1' }),
        h('button', { className: 'btn', onClick: () => { state.editingTask = null; render(); } }, 'Cancel'),
        h('button', { className: 'btn btn-primary', onClick: saveTask }, 'Save'),
      ),
    )
  );
}

async function saveTask() {
  const t = state.editingTask;
  // Normalize IsMilestone
  t.IsMilestone = parseInt(t.IsMilestone) || 0;
  const body = {
    sort_order: t.SortOrder || 0, assignee: t.Assignee || '', title: t.Title || '',
    is_milestone: t.IsMilestone, orig_weeks: t.OrigWeeks || 0, curr_weeks: t.CurrWeeks || 0,
    orig_due: t.OrigDue || '', curr_due: t.CurrDue || '', actual_done: t.ActualDone || '',
    status: t.Status || 'pending', words: t.Words || 0, words_per_hour: t.WordsPerHour || 0,
    hours: t.Hours || 0, rate: t.Rate || 0, budget_notes: t.BudgetNotes || '',
    orig_budget: t.OrigBudget || 0, curr_budget: t.CurrBudget || 0, actual_budget: t.ActualBudget || 0,
  };
  try {
    if (t.ID) await api('/api/tasks/' + t.ID, { method: 'PUT', body: JSON.stringify(body) });
    else await api('/api/projects/' + state.projectId + '/tasks', { method: 'POST', body: JSON.stringify(body) });
    state.editingTask = null;
    state.tasks = await api('/api/projects/' + state.projectId + '/tasks');
    render();
  } catch (e) { alert('Error: ' + e.message); }
}

async function deleteCurrentTask() {
  if (!confirm('Delete this task?')) return;
  try {
    await api('/api/tasks/' + state.editingTask.ID, { method: 'DELETE' });
    state.editingTask = null;
    state.tasks = await api('/api/projects/' + state.projectId + '/tasks');
    render();
  } catch (e) { alert('Error: ' + e.message); }
}

function showAddTask() {
  const maxOrder = state.tasks.reduce((m, t) => Math.max(m, t.SortOrder || 0), 0);
  state.editingTask = {
    ID: 0, Assignee: '', Title: '', Status: 'pending', IsMilestone: 0,
    OrigWeeks: 0, CurrWeeks: 0, OrigDue: '', CurrDue: '', ActualDone: '',
    Hours: 0, Rate: 0, BudgetNotes: '', OrigBudget: 0, CurrBudget: 0, ActualBudget: 0,
    SortOrder: maxOrder + 1, Words: 0, WordsPerHour: 0,
  };
  render();
}

// ─── Duplicate / Make New ───
function showDuplicate() {
  let nameInput, clientInput, projInput, dateInput;
  const el = h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target === el) el.remove(); } },
    h('div', { className: 'modal' },
      h('h2', null, '⧉ Make New From Template'),
      h('p', { style: 'color:var(--text-secondary);font-size:14px;margin-bottom:16px' },
        'Duplicates all tasks with shifted dates. Resets status to pending, zeroes out budgets, and clears actual dates.'
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Project Title'),
          nameInput = h('input', { type: 'text', placeholder: 'e.g. New Book Title' }),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Client Slug'),
          clientInput = h('input', { type: 'text', placeholder: 'e.g. vgr', value: state.project.ClientSlug || '' }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Project Slug'),
          projInput = h('input', { type: 'text', placeholder: 'e.g. newbook' }),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'New Start Date (manuscript transmittal)'),
          dateInput = h('input', { type: 'date', value: new Date().toISOString().slice(0, 10) }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Original Start'),
          h('input', { type: 'text', value: state.project.StartDate || 'not set', disabled: true, style: 'opacity:0.6' }),
        ),
      ),
      h('div', { style: 'background:var(--surface2);border-radius:var(--radius);padding:12px;margin:12px 0;font-size:13px;color:var(--text-secondary)' },
        h('strong', { style: 'color:var(--text)' }, 'What gets copied: '),
        'Task names, assignees, durations (weeks), hours, rates, notes. ',
        h('br'),
        h('strong', { style: 'color:var(--text)' }, 'What gets reset: '),
        'All dates shifted to new start. Budgets zeroed. Status → Pending. Actuals cleared.',
        h('br'),
        h('strong', { style: 'color:var(--text)' }, 'URL: '),
        h('span', { style: 'color:var(--accent2);font-family:monospace' }, '/‹client›/‹project›/'),
      ),
      h('div', { className: 'modal-actions' },
        h('button', { className: 'btn', onClick: () => el.remove() }, 'Cancel'),
        h('button', { className: 'btn btn-primary', onClick: async () => {
          const name = nameInput.value.trim();
          const cs = clientInput.value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
          const ps = projInput.value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
          if (!name || !cs || !ps) { alert('All fields required'); return; }
          try {
            const result = await api('/api/projects/' + state.projectId + '/duplicate', {
              method: 'POST',
              body: JSON.stringify({ name, start_date: dateInput.value, client_slug: cs, project_slug: ps }),
            });
            el.remove();
            // Navigate to the new project's URL
            window.location.href = '/' + cs + '/' + ps + '/';
          } catch (e) { alert('Error: ' + e.message); }
        }}, '⧉ Create New Project'),
      ),
    )
  );
  document.body.appendChild(el);
  nameInput.focus();
}

// ─── Settings ───
function showSettings() {
  const el = h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target === el) el.remove(); } },
    h('div', { className: 'modal' },
      h('h2', null, '⚙ Project Settings'),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Project Name'),
          h('input', { type: 'text', value: state.project.Name, id: 'setting-name' }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Start Date'),
          h('input', { type: 'date', value: state.project.StartDate, id: 'setting-start' }),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Client Slug'),
          h('input', { type: 'text', value: state.project.ClientSlug || '', id: 'setting-client' }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Project Slug'),
          h('input', { type: 'text', value: state.project.ProjectSlug || '', id: 'setting-proj' }),
        ),
      ),
      state.project.ClientSlug && state.project.ProjectSlug
        ? h('div', { style: 'font-size:13px;color:var(--text-secondary);margin-bottom:12px' },
            'URL: ', h('a', { href: '/' + state.project.ClientSlug + '/' + state.project.ProjectSlug + '/', style: 'color:var(--accent2);font-family:monospace' },
              '/' + state.project.ClientSlug + '/' + state.project.ProjectSlug + '/'),
          )
        : null,
      h('button', { className: 'btn btn-primary btn-sm', onClick: async () => {
        const cs = $('#setting-client').value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
        const ps = $('#setting-proj').value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
        await api('/api/projects/' + state.projectId, {
          method: 'PUT',
          body: JSON.stringify({ name: $('#setting-name').value, start_date: $('#setting-start').value, client_slug: cs, project_slug: ps }),
        });
        state.project.Name = $('#setting-name').value;
        state.project.StartDate = $('#setting-start').value;
        state.project.ClientSlug = cs;
        state.project.ProjectSlug = ps;
        el.remove(); render();
      }}, 'Save Project'),
      h('hr', { style: 'margin:16px 0;border-color:var(--border)' }),
      h('h3', { style: 'font-size:14px;margin-bottom:8px' }, 'Password Protection'),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Set/Change Password'),
          h('input', { type: 'password', placeholder: 'New shared password', id: 'setting-pw' }),
        ),
      ),
      h('button', { className: 'btn btn-sm', onClick: async () => {
        const pw = $('#setting-pw').value;
        if (!pw) return;
        await api('/api/projects/' + state.projectId + '/auth', {
          method: 'POST', body: JSON.stringify({ password: pw }),
        });
        alert('Password set!');
      }}, 'Set Password'),
      h('hr', { style: 'margin:16px 0;border-color:var(--border)' }),
      h('h3', { style: 'font-size:14px;margin-bottom:8px' }, 'Import Seed Data'),
      h('input', { type: 'file', accept: '.json', id: 'seed-file' }),
      h('button', { className: 'btn btn-sm', style: 'margin-top:8px', onClick: async () => {
        const f = $('#seed-file').files[0];
        if (!f) return;
        const text = await f.text();
        await api('/api/projects/' + state.projectId + '/seed', {
          method: 'POST', body: text,
        });
        state.tasks = await api('/api/projects/' + state.projectId + '/tasks');
        el.remove(); render();
      }}, '📥 Import JSON'),
      h('div', { className: 'modal-actions' },
        h('button', { className: 'btn btn-danger btn-sm', onClick: async () => {
          if (!confirm('Delete this project and all tasks?')) return;
          await api('/api/projects/' + state.projectId, { method: 'DELETE' });
          state.view = 'projects'; el.remove(); loadProjects();
        }}, 'Delete Project'),
        h('div', { style: 'flex:1' }),
        h('button', { className: 'btn', onClick: () => el.remove() }, 'Close'),
      ),
    )
  );
  document.body.appendChild(el);
}

// ─── Status Toggle ───
async function cycleStatus(t) {
  const cycle = { pending: 'in_progress', in_progress: 'done', done: 'pending' };
  const newStatus = cycle[t.Status] || 'pending';
  const body = {
    sort_order: t.SortOrder || 0, assignee: t.Assignee || '', title: t.Title || '',
    is_milestone: t.IsMilestone || 0, orig_weeks: t.OrigWeeks || 0, curr_weeks: t.CurrWeeks || 0,
    orig_due: t.OrigDue || '', curr_due: t.CurrDue || '', actual_done: newStatus === 'done' ? new Date().toISOString().slice(0,10) : (t.ActualDone || ''),
    status: newStatus, words: t.Words || 0, words_per_hour: t.WordsPerHour || 0,
    hours: t.Hours || 0, rate: t.Rate || 0, budget_notes: t.BudgetNotes || '',
    orig_budget: t.OrigBudget || 0, curr_budget: t.CurrBudget || 0, actual_budget: t.ActualBudget || 0,
  };
  try {
    await api('/api/tasks/' + t.ID, { method: 'PUT', body: JSON.stringify(body) });
    state.tasks = await api('/api/projects/' + state.projectId + '/tasks');
    render();
  } catch (e) { alert('Error: ' + e.message); }
}

// ─── Snapshot Email Modal ───
const SNAPSHOT_RECIPIENTS = [
  { email: 'jdbb@agentmail.to', label: 'JDBB Archive', checked: true, editable: false },
  { email: 'j@djinna.com', label: 'Jenna', checked: true, editable: false },
];

let snapshotRecipients = null;

function initSnapshotRecipients() {
  if (snapshotRecipients) return;
  snapshotRecipients = SNAPSHOT_RECIPIENTS.map(r => ({ ...r }));
  snapshotRecipients.push({ email: '', label: 'Other', checked: false, editable: true });
}

async function checkEmailConfigCal() {
  if (state.emailConfigured !== null) return;
  try {
    const r = await api('/api/email/status');
    state.emailConfigured = r.configured;
  } catch {
    state.emailConfigured = false;
  }
}

async function sendSnapshotEmail() {
  const recipients = snapshotRecipients
    .filter(r => r.checked && r.email.trim())
    .map(r => r.email.trim());
  if (recipients.length === 0) {
    state.snapshotResult = { error: 'Select at least one recipient' };
    render();
    return;
  }
  state.snapshotSending = true;
  state.snapshotResult = null;
  render();
  try {
    const res = await api('/api/projects/' + state.projectId + '/snapshot/email', {
      method: 'POST',
      body: JSON.stringify({ recipients }),
    });
    state.snapshotSending = false;
    state.snapshotResult = { ok: true, sent_to: res.sent_to };
    render();
  } catch (e) {
    state.snapshotSending = false;
    state.snapshotResult = { error: e.message };
    render();
  }
}

function renderSnapshotEmailModal() {
  initSnapshotRecipients();
  checkEmailConfigCal();

  const closeModal = () => {
    state.showSnapshotEmail = false;
    state.snapshotResult = null;
    render();
  };

  return h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target.classList.contains('modal-backdrop')) closeModal(); } },
    h('div', { className: 'modal' },
      h('h2', null, 'Email Project Snapshot'),
      h('p', { style: 'color:var(--text-secondary);font-size:14px;margin:0 0 16px' },
        'Sends a comprehensive email with schedule overview, task list, budget summary, transmittal status, recent files, and journal entries.'
      ),

      state.emailConfigured === false
        ? h('div', { style: 'background:#fef3c7;border-radius:8px;padding:12px;margin-bottom:16px;font-size:13px;color:#92400e' },
            '⚠️ Email is not configured on the server.'
          )
        : null,

      h('div', { style: 'margin-bottom:16px' },
        h('label', { style: 'font-weight:600;display:block;margin-bottom:8px' }, 'Send to:'),
        ...snapshotRecipients.map((r, i) =>
          h('div', { style: 'display:flex;align-items:center;gap:8px;margin-bottom:6px' },
            h('input', {
              type: 'checkbox',
              checked: r.checked,
              onChange: () => { snapshotRecipients[i].checked = !snapshotRecipients[i].checked; render(); },
            }),
            r.editable
              ? h('input', {
                  type: 'email',
                  placeholder: 'email@example.com',
                  value: r.email,
                  style: 'flex:1;padding:6px 10px;border:1px solid var(--border);border-radius:6px;background:var(--surface);color:var(--text);font-size:14px',
                  onInput: (e) => {
                    snapshotRecipients[i].email = e.target.value;
                    snapshotRecipients[i].checked = e.target.value.trim().length > 0;
                  },
                })
              : h('span', { style: 'font-size:14px' }, r.email),
            h('span', { style: 'font-size:12px;color:var(--text-secondary)' }, r.label),
          )
        ),
      ),

      state.snapshotResult?.ok
        ? h('div', { style: 'background:#d1fae5;border-radius:8px;padding:12px;margin-bottom:16px;font-size:13px;color:#065f46' },
            '2713 Sent to: ' + state.snapshotResult.sent_to.join(', ')
          )
        : null,
      state.snapshotResult?.error
        ? h('div', { style: 'background:#fef2f2;border-radius:8px;padding:12px;margin-bottom:16px;font-size:13px;color:#dc2626' },
            '❌ ' + state.snapshotResult.error
          )
        : null,

      h('div', { className: 'modal-actions' },
        state.snapshotResult?.ok
          ? h('button', { className: 'btn btn-primary', onClick: closeModal }, '✓ Done')
          : [
              h('button', { className: 'btn', onClick: closeModal }, 'Cancel'),
              h('button', {
                className: 'btn btn-primary',
                disabled: state.snapshotSending || state.emailConfigured === false ? 'disabled' : undefined,
                onClick: sendSnapshotEmail,
              }, state.snapshotSending ? 'Sending…' : '📨 Send Snapshot'),
            ],
      ),
    ),
  );
}

// ─── Activity Email Modal ───
let activityRecipients = null;

function initActivityRecipients() {
  if (activityRecipients) return;
  activityRecipients = SNAPSHOT_RECIPIENTS.map(r => ({ ...r }));
  activityRecipients.push({ email: '', label: 'Other', checked: false, editable: true });
}

async function sendActivityEmail() {
  const recipients = activityRecipients
    .filter(r => r.checked && r.email.trim())
    .map(r => r.email.trim());
  if (recipients.length === 0) {
    state.activityResult = { error: 'Select at least one recipient' };
    render();
    return;
  }
  state.activitySending = true;
  state.activityResult = null;
  render();
  try {
    const res = await api('/api/projects/' + state.projectId + '/activity/email', {
      method: 'POST',
      body: JSON.stringify({ recipients }),
    });
    state.activitySending = false;
    state.activityResult = { ok: true, sent_to: res.sent_to };
    render();
  } catch (e) {
    state.activitySending = false;
    state.activityResult = { error: e.message };
    render();
  }
}

function renderActivityEmailModal() {
  initActivityRecipients();
  checkEmailConfigCal();

  const closeModal = () => {
    state.showActivityEmail = false;
    state.activityResult = null;
    render();
  };

  return h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target.classList.contains('modal-backdrop')) closeModal(); } },
    h('div', { className: 'modal' },
      h('h2', null, 'Email Activity Update'),
      h('p', { style: 'color:var(--text-secondary);font-size:14px;margin:0 0 16px' },
        'Sends recent file transfers and journal entries from the last 7 days.'
      ),

      state.emailConfigured === false
        ? h('div', { style: 'background:#fef3c7;border-radius:8px;padding:12px;margin-bottom:16px;font-size:13px;color:#92400e' },
            '⚠️ Email is not configured on the server.'
          )
        : null,

      h('div', { style: 'margin-bottom:16px' },
        h('label', { style: 'font-weight:600;display:block;margin-bottom:8px' }, 'Send to:'),
        ...activityRecipients.map((r, i) =>
          h('div', { style: 'display:flex;align-items:center;gap:8px;margin-bottom:6px' },
            h('input', {
              type: 'checkbox',
              checked: r.checked,
              onChange: () => { activityRecipients[i].checked = !activityRecipients[i].checked; render(); },
            }),
            r.editable
              ? h('input', {
                  type: 'email',
                  placeholder: 'email@example.com',
                  value: r.email,
                  style: 'flex:1;padding:6px 10px;border:1px solid var(--border);border-radius:6px;background:var(--surface);color:var(--text);font-size:14px',
                  onInput: (e) => {
                    activityRecipients[i].email = e.target.value;
                    activityRecipients[i].checked = e.target.value.trim().length > 0;
                  },
                })
              : h('span', { style: 'font-size:14px' }, r.email),
            h('span', { style: 'font-size:12px;color:var(--text-secondary)' }, r.label),
          )
        ),
      ),

      state.activityResult?.ok
        ? h('div', { style: 'background:#d1fae5;border-radius:8px;padding:12px;margin-bottom:16px;font-size:13px;color:#065f46' },
            '2713 Sent to: ' + state.activityResult.sent_to.join(', ')
          )
        : null,
      state.activityResult?.error
        ? h('div', { style: 'background:#fef2f2;border-radius:8px;padding:12px;margin-bottom:16px;font-size:13px;color:#dc2626' },
            '❌ ' + state.activityResult.error
          )
        : null,

      h('div', { className: 'modal-actions' },
        state.activityResult?.ok
          ? h('button', { className: 'btn btn-primary', onClick: closeModal }, '✓ Done')
          : [
              h('button', { className: 'btn', onClick: closeModal }, 'Cancel'),
              h('button', {
                className: 'btn btn-primary',
                disabled: state.activitySending || state.emailConfigured === false ? 'disabled' : undefined,
                onClick: sendActivityEmail,
              }, state.activitySending ? 'Sending…' : '📨 Send Activity Update'),
            ],
      ),
    ),
  );
}

// ─── Project Switcher ───
async function loadSiblingProjects() {
  if (!state.pathClient) return;
  try {
    const raw = await api('/api/clients/' + state.pathClient + '/projects');
    state.siblingProjects = raw.map(p => ({
      ID: p.id, Name: p.name,
      ClientSlug: p.client_slug, ProjectSlug: p.project_slug,
    }));
    if (state.view === 'project') render();
  } catch (e) {
    state.siblingProjects = [];
  }
}

function renderProjectSwitcher() {
  if (state.siblingProjects.length < 2) return null;
  return h('select', {
    className: 'project-switcher',
    onChange: (e) => {
      const p = state.siblingProjects.find(p => p.ID === parseInt(e.target.value));
      if (p) window.location.href = '/' + p.ClientSlug + '/' + p.ProjectSlug + '/';
    }
  },
    ...state.siblingProjects.map(p =>
      h('option', { value: String(p.ID), selected: p.ID === state.projectId }, p.Name)
    )
  );
}

// ─── File Log ───
async function loadFileLog() {
  try {
    state.fileLog = await api('/api/projects/' + state.projectId + '/file-log');
  } catch (e) { state.fileLog = []; }
  render();
}

function renderFileLog() {
  const entries = state.fileLog;
  const dirIcon = d => d === 'outbound' ? '↑ Out' : '↓ In';
  const dirClass = d => d === 'outbound' ? 'dir-out' : 'dir-in';
  return h('div', { className: 'filelog-section' },
    h('div', { style: 'display:flex;justify-content:space-between;align-items:center;margin-bottom:12px' },
      h('span', { style: 'font-size:13px;color:var(--text-secondary)' }, entries.length + ' file' + (entries.length !== 1 ? 's' : '') + ' logged'),
      h('div', { style: 'display:flex;gap:8px' },
        h('button', { className: 'btn btn-sm', style: 'font-size:12px', onClick: () => { state.showActivityEmail = true; state.activityResult = null; render(); } }, 'Email'),
        h('button', { className: 'btn btn-sm btn-primary', onClick: () => { state.showFileLogModal = true; render(); } }, '+ Log Transfer'),
      ),
    ),
    entries.length === 0
      ? h('div', { className: 'empty-state', style: 'padding:3rem' }, h('p', null, 'No file transfers logged yet'))
      : h('div', { className: 'table-container' },
          h('table', { className: 'data-table' },
            h('thead', null, h('tr', null,
              h('th', null, 'Date'),
              h('th', { style: 'width:70px' }, 'Dir'),
              h('th', null, 'Filename'),
              h('th', null, 'Type'),
              h('th', null, 'From → To'),
              h('th', null, 'Notes'),
              h('th', { style: 'width:40px' }, ''),
            )),
            h('tbody', null, ...entries.map(e =>
              h('tr', null,
                h('td', { className: 'date' }, fmt.date(e.transfer_date)),
                h('td', null, h('span', { className: 'file-dir ' + dirClass(e.direction) }, dirIcon(e.direction))),
                h('td', { style: 'font-weight:500' }, e.filename || '—'),
                h('td', null, h('span', { className: 'badge badge-dim' }, e.file_type || '—')),
                h('td', { style: 'font-size:13px' }, (e.sent_by || '?') + ' → ' + (e.received_by || '?')),
                h('td', { style: 'color:var(--text-secondary);font-size:13px;max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap' }, e.notes || ''),
                h('td', null, h('button', { className: 'btn btn-sm btn-danger', style: 'padding:2px 6px;font-size:11px', onClick: () => deleteFileLogEntry(e.id) }, '✕')),
              )
            )),
          ),
        ),
  );
}

async function deleteFileLogEntry(entryId) {
  if (!confirm('Delete this file log entry?')) return;
  try {
    await api('/api/projects/' + state.projectId + '/file-log/' + entryId, { method: 'DELETE' });
    loadFileLog();
  } catch (e) { alert('Error: ' + e.message); }
}

function renderFileLogModal() {
  let dirSel, fnameInput, typeInput, sentInput, recvInput, notesInput, dateInput;
  const fileTypes = ['.docx', '.pdf', '.epub', '.tiff', '.jpg', '.png', '.eps', '.indd', '.ai', '.psd'];
  const close = () => { state.showFileLogModal = false; render(); };
  return h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target.classList.contains('modal-backdrop')) close(); } },
    h('div', { className: 'modal' },
      h('h2', null, '+ Log File Transfer'),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Direction'),
          dirSel = h('select', null,
            h('option', { value: 'inbound' }, '↓ Inbound (received)'),
            h('option', { value: 'outbound' }, '↑ Outbound (sent)'),
          ),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Transfer Date'),
          dateInput = h('input', { type: 'date', value: new Date().toISOString().slice(0, 10) }),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Filename'),
          fnameInput = h('input', { type: 'text', placeholder: 'e.g. manuscript_v2.docx' }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'File Type'),
          typeInput = h('select', null,
            h('option', { value: '' }, '— select —'),
            ...fileTypes.map(t => h('option', { value: t }, t)),
            h('option', { value: 'other' }, 'other'),
          ),
        ),
      ),
      h('div', { className: 'form-row' },
        h('div', { className: 'form-group' },
          h('label', null, 'Sent By'),
          sentInput = h('input', { type: 'text', placeholder: 'Name' }),
        ),
        h('div', { className: 'form-group' },
          h('label', null, 'Received By'),
          recvInput = h('input', { type: 'text', placeholder: 'Name' }),
        ),
      ),
      h('div', { className: 'form-group' },
        h('label', null, 'Notes'),
        notesInput = h('textarea', { rows: '2', style: 'width:100%;padding:8px;border:1px solid var(--border);border-radius:8px;background:var(--surface);color:var(--text);font-size:14px;resize:vertical' }),
      ),
      h('div', { className: 'modal-actions' },
        h('button', { className: 'btn', onClick: close }, 'Cancel'),
        h('button', { className: 'btn btn-primary', onClick: async () => {
          const fname = fnameInput.value.trim();
          if (!fname) { alert('Filename required'); return; }
          // Auto-detect file type from filename if not selected
          let ft = typeInput.value;
          if (!ft && fname.includes('.')) {
            ft = '.' + fname.split('.').pop().toLowerCase();
          }
          try {
            await api('/api/projects/' + state.projectId + '/file-log', {
              method: 'POST',
              body: JSON.stringify({
                direction: dirSel.value,
                filename: fname,
                file_type: ft,
                sent_by: sentInput.value.trim(),
                received_by: recvInput.value.trim(),
                notes: notesInput.value.trim(),
                transfer_date: dateInput.value,
              }),
            });
            state.showFileLogModal = false;
            loadFileLog();
          } catch (e) { alert('Error: ' + e.message); }
        }}, 'Save'),
      ),
    ),
  );
}

// ─── Journal ───
async function loadJournal() {
  try {
    state.journal = await api('/api/projects/' + state.projectId + '/journal');
  } catch (e) { state.journal = []; }
  render();
}

const journalTypeEmoji = { call: '📞', decision: '⚖️', approval: '✅', note: '📝' };
const journalTypeLabel = { call: 'Call', decision: 'Decision', approval: 'Approval', note: 'Note' };

function renderJournal() {
  const entries = state.journal;
  return h('div', { className: 'journal-section' },
    h('div', { style: 'display:flex;justify-content:space-between;align-items:center;margin-bottom:12px' },
      h('span', { style: 'font-size:13px;color:var(--text-secondary)' }, entries.length + ' entr' + (entries.length !== 1 ? 'ies' : 'y')),
      h('div', { style: 'display:flex;gap:8px' },
        h('button', { className: 'btn btn-sm', style: 'font-size:12px', onClick: () => { state.showActivityEmail = true; state.activityResult = null; render(); } }, 'Email'),
        h('button', { className: 'btn btn-sm btn-primary', onClick: () => { state.showJournalModal = true; render(); } }, '+ Add Entry'),
      ),
    ),
    entries.length === 0
      ? h('div', { className: 'empty-state', style: 'padding:3rem' }, h('p', null, 'No journal entries yet'))
      : h('div', { className: 'journal-feed' },
          ...entries.map(e => {
            const emoji = journalTypeEmoji[e.entry_type] || '📝';
            const label = journalTypeLabel[e.entry_type] || e.entry_type;
            const dt = new Date(e.created_at);
            const dateStr = dt.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
            const timeStr = dt.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
            return h('div', { className: 'journal-entry' },
              h('div', { className: 'journal-entry-header' },
                h('span', { className: 'journal-type journal-type-' + e.entry_type }, emoji + ' ' + label),
                h('span', { className: 'journal-date' }, dateStr + ' · ' + timeStr),
                h('button', { className: 'btn btn-sm btn-danger', style: 'padding:2px 6px;font-size:11px;margin-left:auto', onClick: () => deleteJournalEntry(e.id) }, '✕'),
              ),
              h('div', { className: 'journal-content' }, e.content),
            );
          }),
        ),
  );
}

async function deleteJournalEntry(entryId) {
  if (!confirm('Delete this journal entry?')) return;
  try {
    await api('/api/projects/' + state.projectId + '/journal/' + entryId, { method: 'DELETE' });
    loadJournal();
  } catch (e) { alert('Error: ' + e.message); }
}

function renderJournalModal() {
  let typeSel, contentArea;
  const close = () => { state.showJournalModal = false; render(); };
  return h('div', { className: 'modal-backdrop', onClick: (e) => { if (e.target.classList.contains('modal-backdrop')) close(); } },
    h('div', { className: 'modal' },
      h('h2', null, '+ Journal Entry'),
      h('div', { className: 'form-group' },
        h('label', null, 'Type'),
        typeSel = h('select', null,
          h('option', { value: 'note' }, '📝 Note'),
          h('option', { value: 'call' }, '📞 Call'),
          h('option', { value: 'decision' }, '⚖️ Decision'),
          h('option', { value: 'approval' }, '✅ Approval'),
        ),
      ),
      h('div', { className: 'form-group' },
        h('label', null, 'Content'),
        contentArea = h('textarea', {
          rows: '5',
          placeholder: 'What happened?',
          style: 'width:100%;padding:10px;border:1px solid var(--border);border-radius:8px;background:var(--surface);color:var(--text);font-size:14px;resize:vertical;font-family:inherit',
        }),
      ),
      h('div', { className: 'modal-actions' },
        h('button', { className: 'btn', onClick: close }, 'Cancel'),
        h('button', { className: 'btn btn-primary', onClick: async () => {
          const content = contentArea.value.trim();
          if (!content) { alert('Content required'); return; }
          try {
            await api('/api/projects/' + state.projectId + '/journal', {
              method: 'POST',
              body: JSON.stringify({ entry_type: typeSel.value, content }),
            });
            state.showJournalModal = false;
            loadJournal();
          } catch (e) { alert('Error: ' + e.message); }
        }}, 'Save'),
      ),
    ),
  );
}

// ─── Boot ───
(async function boot() {
  // Check if URL is /{client}/{project}/ — if so, go directly to that project
  const parts = window.location.pathname.replace(/\/+$/, '').split('/').filter(Boolean);
  if (parts.length >= 2 && parts[0] !== 'api' && parts[0] !== 'static') {
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
      state.tasks = await api('/api/projects/' + info.project.ID + '/tasks');
      state.view = 'project';
      render();
      loadSiblingProjects();
    } catch (e) {
      if (e.message === 'unauthorized') { state.view = 'auth'; render(); }
      else { state.view = 'projects'; loadProjects(); }
    }
  } else {
    loadProjects();
  }
})();
