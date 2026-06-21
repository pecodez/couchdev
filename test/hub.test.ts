import { describe, it, expect, beforeAll } from 'vitest'
import { hubRequest, isHubReachable } from './helpers/hub.js'

const RUN = await isHubReachable()

describe.skipIf(!RUN)('hub API (requires docker-compose up hub)', () => {
  const projectName = `test-proj-${Date.now()}`
  const sessionName = `test-session-${Date.now()}`

  describe('projects', () => {
    it('GET /projects returns an array', async () => {
      const res = await hubRequest('GET', '/projects')
      expect(res.status).toBe(200)
      const body = await res.json()
      expect(Array.isArray(body)).toBe(true)
    })

    it('POST /projects creates a project', async () => {
      const res = await hubRequest('POST', '/projects', {
        name: projectName,
        repo_path: '/mnt/couchdev/projects/sample-app',
      })
      expect(res.status).toBe(201)
      const body = await res.json()
      expect(body.name).toBe(projectName)
      expect(body.id).toBeGreaterThan(0)
    })

    it('POST /projects returns 409 on duplicate name', async () => {
      const res = await hubRequest('POST', '/projects', {
        name: projectName,
        repo_path: '/mnt/couchdev/projects/sample-app',
      })
      expect(res.status).toBe(409)
    })

    it('GET /projects includes the created project', async () => {
      const res = await hubRequest('GET', '/projects')
      const body = await res.json()
      expect(body.some((p: any) => p.name === projectName)).toBe(true)
    })
  })

  describe('sessions', () => {
    it('GET /sessions returns an array', async () => {
      const res = await hubRequest('GET', '/sessions')
      expect(res.status).toBe(200)
      const body = await res.json()
      expect(Array.isArray(body)).toBe(true)
    })

    it('POST /projects/:project/sessions creates a session', async () => {
      const res = await hubRequest('POST', `/projects/${projectName}/sessions`, {
        session: sessionName,
      })
      expect(res.status).toBe(201)
      const body = await res.json()
      expect(body.canonical_name).toBe(`${projectName}/${sessionName}`)
      expect(body.state).toBe('starting')
    })

    it('POST /projects/:project/sessions returns 409 if session is live', async () => {
      const res = await hubRequest('POST', `/projects/${projectName}/sessions`, {
        session: sessionName,
      })
      expect(res.status).toBe(409)
    })

    it('GET /sessions/:project/:session returns session status', async () => {
      const res = await hubRequest('GET', `/sessions/${projectName}/${sessionName}`)
      expect(res.status).toBe(200)
      const body = await res.json()
      expect(body.canonical_name).toBe(`${projectName}/${sessionName}`)
      expect(['starting', 'live', 'resumable']).toContain(body.state)
    })

    it('DELETE /sessions/:project/:session tears down the session', async () => {
      const res = await hubRequest('DELETE', `/sessions/${projectName}/${sessionName}`)
      expect(res.status).toBe(204)
    })

    it('GET /sessions/:project/:session returns dead after teardown', async () => {
      const res = await hubRequest('GET', `/sessions/${projectName}/${sessionName}`)
      expect(res.status).toBe(200)
      const body = await res.json()
      expect(body.state).toBe('dead')
    })

    it('POST /projects/:project/sessions returns 404 for unknown project', async () => {
      const res = await hubRequest('POST', '/projects/does-not-exist/sessions', {
        session: 'whatever',
      })
      expect(res.status).toBe(404)
    })
  })

  describe('auth', () => {
    it('rejects requests without a token', async () => {
      const res = await fetch('http://localhost:8080/api/projects')
      expect(res.status).toBe(401)
    })

    it('rejects requests with a wrong token', async () => {
      const res = await fetch('http://localhost:8080/api/projects', {
        headers: { Authorization: 'Bearer wrong-token' },
      })
      expect(res.status).toBe(401)
    })
  })
})
