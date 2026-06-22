export function extractRepoName(url) {
  if (!url) return ''
  const clean = url.trim().replace(/\/$/, '').replace(/\.git$/, '')
  const parts = clean.includes(':') ? clean.split(':').pop().split('/') : clean.split('/')
  return parts[parts.length - 1] || ''
}

export function isValidProjectName(name) {
  return /^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(name)
}
