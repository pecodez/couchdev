function getToken() { return localStorage.getItem('couchdev_token') || '' }

async function request(method, path, body) {
  const res = await fetch('/api' + path, {
    method,
    headers: { 'Authorization': 'Bearer ' + getToken(), 'Content-Type': 'application/json' },
    body: body != null ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) throw new Error(await res.text())
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  setToken: t => localStorage.setItem('couchdev_token', t),
  listProjects: () => request('GET', '/projects'),
  createProject: (name, repo_path) => request('POST', '/projects', { name, repo_path }),
  listSessions: () => request('GET', '/sessions'),
  createSession: (project, session, cwd) =>
    request('POST', `/projects/${project}/sessions`, { session, cwd: cwd || undefined }),
  getSession: (project, session) => request('GET', `/sessions/${project}/${session}`),
  deleteSession: (project, session) => request('DELETE', `/sessions/${project}/${session}`),
}
