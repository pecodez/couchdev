import { spawn, ChildProcess } from 'node:child_process'
import { createInterface } from 'node:readline'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))

export const WEBDEV_BIN = resolve(__dirname, '../../mcp/webdev/dist/index.js')
export const TEST_DATA_ROOT = process.env.TEST_DATA_ROOT
  ?? resolve(__dirname, '../data')
export const SAMPLE_APP = resolve(__dirname, '../fixtures/sample-app')

export class McpClient {
  private proc: ChildProcess
  private pending = new Map<number, (msg: any) => void>()
  private idCounter = 0

  constructor(env: Record<string, string> = {}) {
    this.proc = spawn('node', [WEBDEV_BIN], {
      env: { ...process.env, DATA_ROOT: TEST_DATA_ROOT, ...env },
      stdio: ['pipe', 'pipe', 'pipe'],
    })

    const rl = createInterface({ input: this.proc.stdout! })
    rl.on('line', (line) => {
      if (!line.trim()) return
      try {
        const msg = JSON.parse(line)
        if (msg.id != null) {
          const resolve = this.pending.get(msg.id)
          if (resolve) {
            this.pending.delete(msg.id)
            resolve(msg)
          }
        }
      } catch { /* ignore non-JSON lines */ }
    })
  }

  async request(method: string, params: Record<string, unknown> = {}): Promise<any> {
    const id = ++this.idCounter
    return new Promise((resolve) => {
      this.pending.set(id, resolve)
      this.proc.stdin!.write(
        JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n',
      )
    })
  }

  close() {
    this.proc.kill()
  }
}
