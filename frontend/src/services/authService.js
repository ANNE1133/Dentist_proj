const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:4000/api';

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    ...options,
  });
  const responseText = (await response.text()).trim();

  if (!response.ok) {
    throw new Error(responseText || `Request failed (${response.status})`);
  }

  if (!responseText) return null;
  return JSON.parse(responseText);
}

export async function checkSession() {
  try {
    await request('/session');
    return true;
  } catch (error) {
    if (error.message === 'authentication required') return false;
    throw error;
  }
}

export function login(username, password) {
  return request('/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
}

export function logout() {
  return request('/logout', { method: 'POST' });
}
