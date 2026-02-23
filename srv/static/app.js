// Production Calendar SPA
const $ = (s, p) => (p || document).querySelector(s);
const $$ = (s, p) => [...(p || document).querySelectorAll(s)];
const h = (tag, attrs, ...kids) => {
  const el = document.createElement(tag);
  if (attrs) Object.entries(attrs).forEach(([k, v]) => {
    if (k.startsWith('on')) el.addEventListener(k.slice(2).toLowerCase(), v);
    else if (k === 'className') el.className = v;
    else if (k === 'htmlFor') el.htmlFor = v;
    else el.setAttribute(k, v);
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

const fmt = {
  date(s) { if (!s) return '—'; const d = new Date(s + 'T00:00:00'); return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }); },
  money(n) { return n ? '$' + n.toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 }) : '—'; },
};

let state = { view: 'projects', projectId: null, project: null, tasks: [], tab: 'gantt', editingTask: null };

function render() {
  const app = $('#app');
  app.innerHTML = '';
  if (state.view === 'projects') app.appendChild(renderProjectList());
  else if (state.view === 'auth') app.appendChild(renderAuth());
  else if (state.view === 'project') app.appendChild(renderProject());
}

// ─── Project List ───
async function loadProjects() {
  state.projects = await api('/api/projects');
  render();
}

function renderProjectList() {
  return h('div', null,
    h('div', { className: 'header' },
      h('h1', null, '📅 Production Calendar'),
      h('div', { className: 'header-actions' },
        h('button', { className: 'btn btn-primary', onClick: showNewProject }, '+ New Project'),
      )
    ),
    state.projects && state.projects.length
      ? h('div', { className: 'project-grid' },
          ...state.projects.map(p =>
            h('div', { className: 'project-card', onClick: () => openProject(p.ID) },
              h('h3', null, p.Name),
              h('div', { className: 'meta' }, 'Start: ' + (p.StartDate || 'Not set')),
              h('div', { className: 'meta' }, 'Updated: ' + new Date(p.UpdatedAt).toLocaleDateString()),
            )
          )
        )
      : h('div', { className: 'empty-state' },
          h('h3', null, 'No projects yet'),
          h('p', null, 'Create your first production schedule'),
        )
  );
}

async function showNewProject() {
  const name = prompt('Project name:');
  if (!name) return;
  const start = prompt('Start date (YYYY-MM-DD):', new Date().toISOString().slice(0, 10));
  const p = await api('/api/projects', { method: 'POST', body: JSON.stringify({ name, start_date: start || '' }) });
  openProject(p.ID);
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
    h('h2', null, '🔒 Password Required'),
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

  return h('div', null,
    h('div', { className: 'header' },
      h('div', null,
        h('button', { className: 'btn btn-sm', onClick: () => { state.view = 'projects'; loadProjects(); }, style: 'margin-bottom:8px' }, '← Projects'),
        h('h1', null, state.project.Name),
      ),
      h('div', { className: 'header-actions' },
        h('button', { className: 'btn btn-sm', onClick: showAddTask }, '+ Task'),
        h('button', { className: 'btn btn-sm', onClick: showSettings }, '⚙ Settings'),
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
    ),
    state.tab === 'gantt' ? renderGantt() : state.tab === 'table' ? renderTable() : renderBudget(),
    state.editingTask ? renderEditModal() : null,
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
  const pxPerDay = 3;
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

  const weekHeaders = h('div', { className: 'gantt-header-weeks', style: `width:${chartWidth}px` },
    ...weeks.map(w => {
      const left = ((w - minDate) / 86400000) * pxPerDay;
      const width = 7 * pxPerDay;
      return h('div', { className: 'gantt-header-week', style: `width:${width}px` },
        w.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
      );
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
      h('td', null, h('span', { className: statusCls }, statusLabel)),
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
          h('td', null, h('span', { className: 'status status-' + (t.Status || 'pending') },
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
            h('td', { style: 'color:var(--text2)' }, t.BudgetNotes),
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
      h('button', { className: 'btn btn-primary btn-sm', onClick: async () => {
        await api('/api/projects/' + state.projectId, {
          method: 'PUT',
          body: JSON.stringify({ name: $('#setting-name').value, start_date: $('#setting-start').value }),
        });
        state.project.Name = $('#setting-name').value;
        state.project.StartDate = $('#setting-start').value;
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
      }}, '🔒 Set Password'),
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

// ─── Boot ───
loadProjects();
