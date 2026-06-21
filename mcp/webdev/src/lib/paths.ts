import path from 'node:path'
import fs from 'node:fs'

export const DATA_ROOT = process.env.DATA_ROOT ?? '/mnt/couchdev'

export function projectName(projectPath: string): string {
  return path.basename(projectPath)
}

export function buildLogPath(projectPath: string, pid: number, tool: string): string {
  const dir = path.join(DATA_ROOT, 'logs', projectName(projectPath), String(pid))
  fs.mkdirSync(dir, { recursive: true })
  return path.join(dir, `${tool}.log`)
}

export function buildScreenshotPath(projectPath: string, pid: number, name: string): string {
  const dir = path.join(DATA_ROOT, 'screenshots', projectName(projectPath), String(pid))
  fs.mkdirSync(dir, { recursive: true })
  return path.join(dir, `${name}.png`)
}
