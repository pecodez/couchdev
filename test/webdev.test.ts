import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { McpClient, TEST_DATA_ROOT, SAMPLE_APP } from './helpers/mcp.js'

function waitFor(condition: () => boolean, timeoutMs = 15000, intervalMs = 500): Promise<void> {
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + timeoutMs
    const check = () => {
      if (condition()) return resolve()
      if (Date.now() > deadline) return reject(new Error('waitFor timed out'))
      setTimeout(check, intervalMs)
    }
    check()
  })
}

describe('webdev MCP node', () => {
  let client: McpClient

  beforeEach(() => {
    client = new McpClient()
  })

  afterEach(() => {
    client.close()
  })

  describe('tools/list', () => {
    it('returns all four webdev tools', async () => {
      const res = await client.request('tools/list')
      const names = res.result.tools.map((t: any) => t.name)
      expect(names).toContain('webdev_run_storybook')
      expect(names).toContain('webdev_run_playwright')
      expect(names).toContain('webdev_build')
      expect(names).toContain('webdev_preview')
      expect(names).toHaveLength(4)
    })

    it('all tools have project_path in their input schema', async () => {
      const res = await client.request('tools/list')
      for (const tool of res.result.tools) {
        expect(tool.inputSchema.properties).toHaveProperty('project_path')
        expect(tool.inputSchema.required).toContain('project_path')
      }
    })
  })

  describe('webdev_run_storybook', () => {
    it('returns a numeric pid immediately', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_storybook',
        arguments: { project_path: SAMPLE_APP },
      })
      const result = JSON.parse(res.result.content[0].text)
      expect(typeof result.pid).toBe('number')
      expect(result.pid).toBeGreaterThan(0)
    })

    it('creates the log directory and file under DATA_ROOT', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_storybook',
        arguments: { project_path: SAMPLE_APP },
      })
      const { pid } = JSON.parse(res.result.content[0].text)
      const logPath = join(TEST_DATA_ROOT, 'logs', 'sample-app', String(pid), 'storybook.log')

      await waitFor(() => existsSync(logPath), 5000)
      expect(existsSync(logPath)).toBe(true)
    })

    it('writes a COUCHDEV:EXIT marker to the log', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_storybook',
        arguments: { project_path: SAMPLE_APP },
      })
      const { pid } = JSON.parse(res.result.content[0].text)
      const logPath = join(TEST_DATA_ROOT, 'logs', 'sample-app', String(pid), 'storybook.log')

      await waitFor(
        () => existsSync(logPath) && readFileSync(logPath, 'utf8').includes('COUCHDEV:EXIT:'),
        25000,
      )
      const log = readFileSync(logPath, 'utf8')
      expect(log).toMatch(/COUCHDEV:EXIT:[01]/)
    })
  })

  describe('webdev_run_playwright', () => {
    it('rejects when both test_name and tag are provided', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_playwright',
        arguments: { project_path: SAMPLE_APP, test_name: 'foo', tag: 'smoke' },
      })
      expect(res.result.isError).toBe(true)
      expect(res.result.content[0].text).toContain('mutually exclusive')
    })

    it('returns a pid when called with a spec filter', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_playwright',
        arguments: { project_path: SAMPLE_APP, spec: 'e2e/home.spec.ts' },
      })
      const result = JSON.parse(res.result.content[0].text)
      expect(typeof result.pid).toBe('number')
      expect(result.pid).toBeGreaterThan(0)
    })

    it('returns a pid when called with a tag filter', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_playwright',
        arguments: { project_path: SAMPLE_APP, tag: 'smoke' },
      })
      const result = JSON.parse(res.result.content[0].text)
      expect(typeof result.pid).toBe('number')
    })

    it('creates the log file under DATA_ROOT', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_run_playwright',
        arguments: { project_path: SAMPLE_APP },
      })
      const { pid } = JSON.parse(res.result.content[0].text)
      const logPath = join(TEST_DATA_ROOT, 'logs', 'sample-app', String(pid), 'playwright.log')

      await waitFor(() => existsSync(logPath), 5000)
      expect(existsSync(logPath)).toBe(true)
    })
  })

  describe('webdev_build', () => {
    it('returns a numeric pid immediately', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_build',
        arguments: { project_path: SAMPLE_APP, mode: 'production' },
      })
      const result = JSON.parse(res.result.content[0].text)
      expect(typeof result.pid).toBe('number')
      expect(result.pid).toBeGreaterThan(0)
    })

    it('creates a mode-named log file under DATA_ROOT', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_build',
        arguments: { project_path: SAMPLE_APP, mode: 'staging' },
      })
      const { pid } = JSON.parse(res.result.content[0].text)
      const logPath = join(TEST_DATA_ROOT, 'logs', 'sample-app', String(pid), 'build-staging.log')

      await waitFor(() => existsSync(logPath), 5000)
      expect(existsSync(logPath)).toBe(true)
    })
  })

  describe('webdev_preview', () => {
    it('returns a numeric pid immediately', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_preview',
        arguments: { project_path: SAMPLE_APP, mode: 'production' },
      })
      const result = JSON.parse(res.result.content[0].text)
      expect(typeof result.pid).toBe('number')
      expect(result.pid).toBeGreaterThan(0)
    })

    it('creates a mode-named log file under DATA_ROOT', async () => {
      const res = await client.request('tools/call', {
        name: 'webdev_preview',
        arguments: { project_path: SAMPLE_APP, mode: 'production' },
      })
      const { pid } = JSON.parse(res.result.content[0].text)
      const logPath = join(TEST_DATA_ROOT, 'logs', 'sample-app', String(pid), 'preview-production.log')

      await waitFor(() => existsSync(logPath), 5000)
      expect(existsSync(logPath)).toBe(true)
    })
  })
})
