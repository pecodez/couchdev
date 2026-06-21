export const HUB_URL = process.env.HUB_URL ?? 'http://localhost:8080'
export const HUB_TOKEN = process.env.HUB_TOKEN ?? 'couchdev-test-token'

export async function hubRequest(
  method: string,
  path: string,
  body?: unknown,
): Promise<Response> {
  return fetch(`${HUB_URL}/api${path}`, {
    method,
    headers: {
      Authorization: `Bearer ${HUB_TOKEN}`,
      'Content-Type': 'application/json',
    },
    body: body != null ? JSON.stringify(body) : undefined,
  })
}

export async function isHubReachable(): Promise<boolean> {
  try {
    const res = await fetch(`${HUB_URL}/api/projects`, {
      headers: { Authorization: `Bearer ${HUB_TOKEN}` },
      signal: AbortSignal.timeout(2000),
    })
    return res.status !== 0
  } catch {
    return false
  }
}
