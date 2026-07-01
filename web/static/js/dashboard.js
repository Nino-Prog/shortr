async function loadLinks() {
  if (!token()) { window.location.href = '/'; return; }

  const res = await fetch('/api/links', { headers: authHeaders() });
  if (res.status === 401) { window.location.href = '/'; return; }

  const links = await res.json();
  const container = document.getElementById('links-list');

  if (!links || links.length === 0) {
    container.innerHTML = '<p class="loading">No links yet. <a href="/" style="color:var(--accent)">Create one →</a></p>';
    return;
  }

  container.innerHTML = links.map(link => `
    <div class="link-item" id="link-${link.code}">
      <div>
        <div class="link-code">
          <a href="/${link.code}" target="_blank" style="color:var(--accent);text-decoration:none">
            ${window.location.origin}/${link.code}
          </a>
        </div>
        <div class="link-url" title="${link.original_url}">${link.original_url}</div>
        <div style="font-size:0.75rem;color:var(--muted);margin-top:0.2rem">
          ${new Date(link.created_at).toLocaleDateString()}
          ${link.expires_at ? ' · expires ' + new Date(link.expires_at).toLocaleDateString() : ''}
        </div>
      </div>
      <div class="link-actions">
        <button class="btn-ghost" onclick="copyLink('${link.code}')">Copy</button>
        <button onclick="showAnalytics('${link.code}')">Analytics</button>
        <button class="btn-danger" onclick="deleteLink('${link.code}')">Delete</button>
      </div>
    </div>
  `).join('');
}

async function copyLink(code) {
  try {
    await navigator.clipboard.writeText(`${window.location.origin}/${code}`);
    alert('Copied!');
  } catch { alert(`${window.location.origin}/${code}`); }
}

async function deleteLink(code) {
  if (!confirm(`Delete /${code}? This cannot be undone.`)) return;
  const res = await fetch(`/api/links/${code}`, { method: 'DELETE', headers: authHeaders() });
  if (res.ok) {
    document.getElementById(`link-${code}`)?.remove();
  } else {
    alert('Could not delete link.');
  }
}

async function showAnalytics(code) {
  const panel = document.getElementById('analytics-panel');
  document.getElementById('analytics-code').textContent = code;
  panel.style.display = '';

  const res = await fetch(`/api/analytics/${code}`, { headers: authHeaders() });
  if (!res.ok) { panel.innerHTML = '<p class="error">Could not load analytics.</p>'; return; }

  const data = await res.json();

  document.getElementById('total-clicks').textContent = data.total_clicks;

  const topCountry = Object.entries(data.by_country || {}).sort((a,b) => b[1]-a[1])[0];
  const topCity    = Object.entries(data.by_city    || {}).sort((a,b) => b[1]-a[1])[0];
  document.getElementById('top-country').textContent = topCountry ? `${topCountry[0]} (${topCountry[1]})` : '—';
  document.getElementById('top-city').textContent    = topCity    ? `${topCity[0]} (${topCity[1]})`    : '—';

  const tbody = document.getElementById('recent-body');
  tbody.innerHTML = (data.recent || []).map(c => `
    <tr>
      <td>${new Date(c.clicked_at).toLocaleString()}</td>
      <td>${c.country || '—'}</td>
      <td>${c.city || '—'}</td>
    </tr>
  `).join('') || '<tr><td colspan="3" style="color:var(--muted)">No clicks yet.</td></tr>';

  panel.scrollIntoView({ behavior: 'smooth' });
}

function closeAnalytics() {
  document.getElementById('analytics-panel').style.display = 'none';
}

window.addEventListener('DOMContentLoaded', loadLinks);
