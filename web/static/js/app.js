const API = '';

function token() { return localStorage.getItem('shortr_token'); }
function setToken(t) { localStorage.setItem('shortr_token', t); }
function clearToken() { localStorage.removeItem('shortr_token'); }

function authHeaders() {
  return { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token() };
}

function logout() {
  clearToken();
  window.location.href = '/';
}

let registering = false;

function toggleAuth() {
  registering = !registering;
  document.getElementById('auth-title').textContent = registering ? 'Create account' : 'Sign in';
  document.getElementById('auth-btn').textContent = registering ? 'Register' : 'Sign in';
  document.getElementById('auth-btn').onclick = registering ? register : login;
  document.querySelector('.switch').innerHTML = registering
    ? 'Have an account? <a href="#" onclick="toggleAuth()">Sign in</a>'
    : 'No account? <a href="#" onclick="toggleAuth()">Register</a>';
}

async function login() {
  const email = document.getElementById('email').value.trim();
  const password = document.getElementById('password').value;
  const err = document.getElementById('auth-error');
  err.textContent = '';

  const res = await fetch(API + '/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!res.ok) { err.textContent = 'Invalid email or password.'; return; }
  const data = await res.json();
  setToken(data.token);
  showApp();
}

async function register() {
  const email = document.getElementById('email').value.trim();
  const password = document.getElementById('password').value;
  const err = document.getElementById('auth-error');
  err.textContent = '';

  if (password.length < 8) { err.textContent = 'Password must be at least 8 characters.'; return; }

  const res = await fetch(API + '/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!res.ok) { err.textContent = 'Email already registered.'; return; }
  const data = await res.json();
  setToken(data.token);
  showApp();
}

function showApp() {
  document.getElementById('auth-section').style.display = 'none';
  document.getElementById('app-section').style.display = '';
}

async function shorten() {
  const url = document.getElementById('url-input').value.trim();
  const code = document.getElementById('code-input').value.trim();
  const resultEl = document.getElementById('result');

  if (!url) { resultEl.textContent = 'Please enter a URL.'; resultEl.style.color = 'var(--danger)'; return; }

  const body = { url };
  if (code) body.code = code;

  const res = await fetch(API + '/api/shorten', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify(body),
  });

  if (res.status === 401) { resultEl.textContent = 'Please sign in first.'; resultEl.style.color = 'var(--danger)'; return; }
  if (!res.ok) { resultEl.textContent = 'Could not shorten — code may be taken.'; resultEl.style.color = 'var(--danger)'; return; }

  const link = await res.json();
  const shortURL = `${window.location.origin}/${link.code}`;
  resultEl.innerHTML = `<a href="${shortURL}" target="_blank" style="color:var(--success)">${shortURL}</a>`;

  // Copy to clipboard
  try { await navigator.clipboard.writeText(shortURL); resultEl.innerHTML += ' <span style="color:var(--muted);font-size:.8rem">(copied!)</span>'; } catch {}
}

// On load: if token exists, skip auth form
window.addEventListener('DOMContentLoaded', () => {
  if (token() && document.getElementById('auth-section')) {
    showApp();
  }
});
