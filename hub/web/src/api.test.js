import { describe, it, expect, vi, beforeEach } from 'vitest'
import { api, onUnauthorized } from './api.js'

function mockFetch({ status = 200, ok = true, body = null, text = '' } = {}) {
  globalThis.fetch = vi.fn().mockResolvedValue({
    status,
    ok,
    json: () => Promise.resolve(body),
    text: () => Promise.resolve(text),
  })
}

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('api.createProject', () => {
  it('returns parsed JSON on success', async () => {
    const project = { id: 1, name: 'my-project' }
    mockFetch({ body: project })
    const result = await api.createProject({ name: 'my-project', source_type: 'greenfield' })
    expect(result).toEqual(project)
  })

  it('sends Authorization header from localStorage', async () => {
    localStorage.setItem('couchdev_token', 'test-token')
    mockFetch({ body: {} })
    await api.createProject({ name: 'x', source_type: 'greenfield' })
    const [, opts] = globalThis.fetch.mock.calls[0]
    expect(opts.headers['Authorization']).toBe('Bearer test-token')
  })

  it('sends body as JSON', async () => {
    mockFetch({ body: {} })
    const payload = { name: 'x', source_type: 'clone', repo_url: 'git@github.com:o/r.git' }
    await api.createProject(payload)
    const [, opts] = globalThis.fetch.mock.calls[0]
    expect(JSON.parse(opts.body)).toEqual(payload)
  })

  it('throws with server error text on non-ok response', async () => {
    mockFetch({ status: 409, ok: false, text: 'UNIQUE constraint failed: projects.name' })
    await expect(api.createProject({ name: 'x', source_type: 'greenfield' }))
      .rejects.toThrow('UNIQUE constraint failed: projects.name')
  })

  it('clears token and calls onUnauthorized on 401', async () => {
    localStorage.setItem('couchdev_token', 'stale-token')
    const cb = vi.fn()
    onUnauthorized(cb)
    mockFetch({ status: 401, ok: false, text: 'unauthorized' })
    await expect(api.createProject({ name: 'x', source_type: 'greenfield' }))
      .rejects.toThrow('unauthorized')
    expect(localStorage.getItem('couchdev_token')).toBeNull()
    expect(cb).toHaveBeenCalledOnce()
  })
})

describe('api.listProjects', () => {
  it('returns null on 204', async () => {
    mockFetch({ status: 204, ok: true })
    const result = await api.listProjects()
    expect(result).toBeNull()
  })

  it('sends no body for GET requests', async () => {
    mockFetch({ body: [] })
    await api.listProjects()
    const [, opts] = globalThis.fetch.mock.calls[0]
    expect(opts.body).toBeUndefined()
  })
})
