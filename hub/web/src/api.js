let _onUnauthorized = null

export function onUnauthorized(cb) {
  _onUnauthorized = cb
}

function getToken() { return localStorage.getItem('couchdev_token') || '' }

async function request(method, path, body) {
  const res = await fetch('/api' + path, {
    method,
    headers: { 'Authorization': 'Bearer ' + getToken(), 'Content-Type': 'application/json' },
    body: body != null ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401) {
    localStorage.removeItem('couchdev_token')
    if (_onUnauthorized) _onUnauthorized()
    throw new Error('unauthorized')
  }
  if (!res.ok) {
    const text = await res.text()
    let body = text
    try { body = JSON.parse(text) } catch { /* not JSON */ }
    const err = new Error(typeof body === 'string' ? body : (body.reason || res.statusText))
    err.status = res.status
    err.body = body
    throw err
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  setToken: t => localStorage.setItem('couchdev_token', t),
  clearToken: () => localStorage.removeItem('couchdev_token'),
  verify: () => request('GET', '/projects').then(() => true).catch(() => false),
  listProjects: () => request('GET', '/projects'),
  createProject: (payload) => request('POST', '/projects', payload),
  listSessions: () => request('GET', '/sessions'),
  createSession: (project, session, cwd) =>
    request('POST', `/projects/${project}/sessions`, { session, cwd: cwd || undefined }),
  getSession: (project, session) => request('GET', `/sessions/${project}/${session}`),
  deleteSession: (project, session, force) =>
    request('DELETE', `/sessions/${project}/${session}${force ? '?force=true' : ''}`),
  getChanges: (project, session) => request('GET', `/sessions/${project}/${session}/changes`),
  resumeSession: (project, session) => request('POST', `/sessions/${project}/${session}/resume`, {}),
}
